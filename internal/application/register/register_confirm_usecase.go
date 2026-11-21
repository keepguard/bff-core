package register

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	client "github.com/keepguard/bff-core/internal/application/port"
	outport "github.com/keepguard/bff-core/internal/application/port/out"
	"github.com/keepguard/bff-core/internal/domain/enums"
	"github.com/keepguard/bff-core/internal/domain/saga"
	"github.com/keepguard/bff-core/internal/pkg"
	"go.uber.org/zap"
)

// registerConfirmUseCaseImpl implementa o caso de uso de confirmação de registro
type registerConfirmUseCaseImpl struct {
	userClient          client.UserClient
	companyClient       client.CompanyClient
	authClient          client.AuthClient
	userConsentClient   client.UserConsentClient
	communicationClient client.CommunicationClient
	messagePublisher    outport.MessagePublisher
	sagaExecutor        *saga.InMemorySagaExecutor
	logger              *zap.Logger
	opts                RegisterOptions
	lockTTL             time.Duration
	resultTTL           time.Duration
	waitTimeout         time.Duration
	pollInterval        time.Duration
}

// NewRegisterConfirmUseCase cria um novo caso de uso de confirmação de registro
func NewRegisterConfirmUseCase(
	userClient client.UserClient,
	companyClient client.CompanyClient,
	authClient client.AuthClient,
	userConsentClient client.UserConsentClient,
	communicationClient client.CommunicationClient,
	messagePublisher outport.MessagePublisher,
	logger *zap.Logger,
	opts RegisterOptions,
) RegisterConfirmUseCase {
	return &registerConfirmUseCaseImpl{
		userClient:          userClient,
		companyClient:       companyClient,
		authClient:          authClient,
		userConsentClient:   userConsentClient,
		communicationClient: communicationClient,
		messagePublisher:    messagePublisher,
		sagaExecutor:        saga.NewInMemorySagaExecutor(logger),
		logger:              logger,
		opts:                opts,
		lockTTL:             registerLockTTL,
		resultTTL:           registerResultTTL,
		waitTimeout:         registerWaitTimeout,
		pollInterval:        registerPollInterval,
	}
}

// Execute executa o caso de uso de confirmação de registro. Com saga_hardening ligado, o mesmo confirm repetido
// devolve o resultado já concluído e dois confirms simultâneos não rodam a saga duas vezes.
func (uc *registerConfirmUseCaseImpl) Execute(ctx context.Context, command appdto.RegisterConfirmCommand) (appdto.RegisterConfirmViewDTO, error) {
	if !uc.opts.SagaHardening || uc.opts.State == nil {
		return uc.runOnce(ctx, command)
	}
	return uc.runIdempotent(ctx, command)
}

// runIdempotent: replay do resultado, trava por confirm e espera pelo resultado de quem já está rodando.
func (uc *registerConfirmUseCaseImpl) runIdempotent(ctx context.Context, command appdto.RegisterConfirmCommand) (appdto.RegisterConfirmViewDTO, error) {
	material := confirmMaterial(command)
	fingerprint := confirmFingerprint(material)
	resultKey := registerResultKeyPrefix + fingerprint
	lockKey := registerLockKeyPrefix + fingerprint
	deadline := time.Now().Add(uc.waitTimeout)

	for {
		if view, ok := uc.loadReplay(ctx, resultKey, material, command.CorrelationID); ok {
			uc.metrics().RecordStep("ConfirmRegister", "replayed")
			return view, nil
		}

		acquired, err := uc.opts.State.TryLock(ctx, lockKey, uc.lockTTL)
		if err != nil {
			// Redis fora do ar: segue sem trava, como antes da saga endurecida (o banco e o ms-user decidem).
			uc.logger.Warn("Trava do cadastro indisponível; seguindo sem trava",
				zap.String("correlation_id", command.CorrelationID),
				zap.Error(err))
			uc.metrics().RecordStep("ConfirmLock", "store_error")
			return uc.runOnce(ctx, command)
		}
		if acquired {
			return uc.runAsOwner(ctx, command, material, resultKey, lockKey)
		}

		if time.Now().After(deadline) {
			uc.metrics().RecordStep("ConfirmLock", "in_progress")
			return appdto.RegisterConfirmViewDTO{}, pkg.NewAppError(
				"REGISTRATION_IN_PROGRESS",
				"Cadastro em andamento. Tente novamente em instantes.",
				http.StatusConflict,
			)
		}

		select {
		case <-ctx.Done():
			return appdto.RegisterConfirmViewDTO{}, ctx.Err()
		case <-time.After(uc.pollInterval):
		}
	}
}

// runAsOwner roda a saga com a trava; guarda o resultado (cifrado) antes de soltar a trava.
func (uc *registerConfirmUseCaseImpl) runAsOwner(ctx context.Context, command appdto.RegisterConfirmCommand, material, resultKey, lockKey string) (appdto.RegisterConfirmViewDTO, error) {
	// Soltar a trava não pode depender do request: se o cliente caiu, a trava ainda precisa sair.
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	defer func() {
		if err := uc.opts.State.Unlock(releaseCtx, lockKey); err != nil {
			uc.logger.Warn("Falha ao soltar a trava do cadastro (expira pelo prazo)",
				zap.String("correlation_id", command.CorrelationID),
				zap.Error(err))
		}
	}()

	view, err := uc.runOnce(ctx, command)
	if err != nil {
		return view, err
	}
	uc.storeReplay(releaseCtx, resultKey, material, view, command.CorrelationID)
	return view, nil
}

// loadReplay devolve o resultado já concluído do mesmo confirm. Falha de leitura ou de decifra vira "não há".
func (uc *registerConfirmUseCaseImpl) loadReplay(ctx context.Context, resultKey, material, correlationID string) (appdto.RegisterConfirmViewDTO, bool) {
	sealed, found, err := uc.opts.State.GetResult(ctx, resultKey)
	if err != nil {
		uc.logger.Warn("Leitura do resultado do cadastro falhou; seguindo sem replay",
			zap.String("correlation_id", correlationID),
			zap.Error(err))
		return appdto.RegisterConfirmViewDTO{}, false
	}
	if !found {
		return appdto.RegisterConfirmViewDTO{}, false
	}
	plain, err := openReplay(material, sealed)
	if err != nil {
		uc.logger.Warn("Resultado do cadastro não pôde ser aberto; ignorado",
			zap.String("correlation_id", correlationID))
		return appdto.RegisterConfirmViewDTO{}, false
	}
	var stored struct {
		Token          string `json:"token"`
		TokenExpiresIn int64  `json:"tokenExpiresIn"`
	}
	if err := json.Unmarshal(plain, &stored); err != nil {
		return appdto.RegisterConfirmViewDTO{}, false
	}
	uc.logger.Info("Confirmação repetida: devolvendo o resultado já concluído",
		zap.String("correlation_id", correlationID))
	return appdto.RegisterConfirmViewDTO{Token: stored.Token, TokenExpiresIn: stored.TokenExpiresIn}, true
}

// storeReplay guarda o resultado cifrado por registerResultTTL. Falha só vira log: o cadastro já foi feito.
func (uc *registerConfirmUseCaseImpl) storeReplay(ctx context.Context, resultKey, material string, view appdto.RegisterConfirmViewDTO, correlationID string) {
	plain, err := json.Marshal(struct {
		Token          string `json:"token"`
		TokenExpiresIn int64  `json:"tokenExpiresIn"`
	}{Token: view.Token, TokenExpiresIn: view.TokenExpiresIn})
	if err != nil {
		uc.logger.Warn("Falha ao montar o resultado do cadastro para replay", zap.String("correlation_id", correlationID))
		return
	}
	sealed, err := sealReplay(material, plain)
	if err != nil {
		uc.logger.Warn("Falha ao cifrar o resultado do cadastro para replay", zap.String("correlation_id", correlationID))
		return
	}
	if err := uc.opts.State.SetResult(ctx, resultKey, sealed, uc.resultTTL); err != nil {
		uc.logger.Warn("Falha ao guardar o resultado do cadastro para replay",
			zap.String("correlation_id", correlationID),
			zap.Error(err))
		uc.metrics().RecordStep("ConfirmResult", "store_error")
	}
}

// runOnce executa a saga uma vez (sem trava nem replay) e envia o e-mail de boas-vindas se ela completou.
func (uc *registerConfirmUseCaseImpl) runOnce(ctx context.Context, command appdto.RegisterConfirmCommand) (appdto.RegisterConfirmViewDTO, error) {
	uc.logger.Info("Iniciando confirmação de registro com SAGA",
		zap.String("correlation_id", command.CorrelationID))

	// Preparar dados para o SAGA
	sagaData := map[string]interface{}{
		"command": command,
	}

	// Construir e executar SAGA
	registerSaga := uc.buildRegisterConfirmSaga()

	err := uc.sagaExecutor.Execute(ctx, registerSaga, sagaData)
	if err != nil {
		uc.logger.Error("SAGA de registro falhou",
			zap.String("correlation_id", command.CorrelationID),
			zap.Error(err))
		return appdto.RegisterConfirmViewDTO{}, err
	}

	// Extrair resultado do login do sagaData
	loginResponse, ok := sagaData["loginResponse"].(appdto.AuthLoginResponseDTO)
	if !ok {
		return appdto.RegisterConfirmViewDTO{}, fmt.Errorf("resposta de login não encontrada no SAGA")
	}

	// Enviar email de boas-vindas APENAS se SAGA completou com sucesso
	uc.sendWelcomeEmail(ctx, sagaData, command.TenantId, command.CorrelationID)

	uc.logger.Info("Registro confirmado com sucesso via SAGA",
		zap.String("correlation_id", command.CorrelationID))

	return appdto.RegisterConfirmViewDTO{
		Token:          loginResponse.Token,
		TokenExpiresIn: loginResponse.ExpiresIn,
	}, nil
}

// buildRegisterConfirmSaga constrói os steps da SAGA de confirmação de registro
func (uc *registerConfirmUseCaseImpl) buildRegisterConfirmSaga() saga.InMemorySaga {
	steps := []saga.InMemoryStep{
		// Step 1: Validar Company (read-only, sem compensação)
		{
			Name: "ValidateCompany",
			Execute: func(ctx context.Context, data map[string]interface{}) error {
				command := data["command"].(appdto.RegisterConfirmCommand)
				if company, ok := data["company"].(appdto.MSCompanyResponseDTO); ok {
					ctx = client.WithCompanyID(ctx, company.ID)
				}
				company, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
				if err != nil {
					return err
				}
				data["company"] = company
				return nil
			},
			Compensate: nil, // read-only
			MaxRetries: 1,   // Sem retry - delegado ao decorator se necessário
			Timeout:    5 * time.Second,
		},
		// Step 2: Confirmar Registro (stateless, sem compensação)
		{
			Name: "ConfirmRegisterSession",
			Execute: func(ctx context.Context, data map[string]interface{}) error {
				command := data["command"].(appdto.RegisterConfirmCommand)
				if company, ok := data["company"].(appdto.MSCompanyResponseDTO); ok {
					ctx = client.WithCompanyID(ctx, company.ID)
				}
				confirmRequest := appdto.MSUserRegisterConfirmRequestDTO{
					Email:                 command.Email,
					RegistrationSessionID: command.RegistrationSessionId,
					Token:                 command.Token,
					EmailToken:            command.EmailToken,
					SmsToken:              command.SmsToken,
					WhatsAppToken:         command.WhatsAppToken,
				}
				confirmResponse, err := uc.userClient.ConfirmRegister(ctx, confirmRequest, command.TenantId, command.CorrelationID)
				if err != nil {
					return err
				}
				data["confirmResponse"] = confirmResponse
				return nil
			},
			Compensate: nil, // stateless
			MaxRetries: 1,   // Sem retry - delegado ao decorator se necessário
			Timeout:    5 * time.Second,
		},
		// Step 3: Criar Usuário (COM compensação)
		{
			Name: "CreateUser",
			Execute: func(ctx context.Context, data map[string]interface{}) error {
				command := data["command"].(appdto.RegisterConfirmCommand)
				if company, ok := data["company"].(appdto.MSCompanyResponseDTO); ok {
					ctx = client.WithCompanyID(ctx, company.ID)
				}
				company := data["company"].(appdto.MSCompanyResponseDTO)
				confirmResponse := data["confirmResponse"].(appdto.MSUserRegisterConfirmResponseDTO)

				userRequest := appdto.MSUserCreateRequestDTO{
					CompanyID:       company.ID,
					Type:            confirmResponse.Type,
					Email:           confirmResponse.Email,
					PhoneE164:       confirmResponse.Phone,
					PreferredLocale: registerLocale(confirmResponse.PreferredLocale),
					CountryCode:     confirmResponse.CountryCode,
					Timezone:        "America/Sao_Paulo",
					Status:          "ACTIVE",
					PersonProfile: &appdto.PersonProfileDTO{
						FullName:  confirmResponse.NameFull,
						KYCLevel:  "BASIC",
						KYCStatus: "NOT_STARTED",
						PEP:       false,
					},
				}
				userResponse, err := uc.userClient.CreateUser(ctx, userRequest, command.TenantId, command.CorrelationID)
				if err != nil {
					return err
				}
				data["user"] = userResponse
				return nil
			},
			// Cleanup delegated to the ms-user reaper (PENDING profile expires in 30 min, with reconciliation).
			Compensate: uc.delegateToReaper("CreateUser", "ms-user"),
			MaxRetries: 1, // Sem retry - delegado ao decorator se necessário
			Timeout:    10 * time.Second,
		},
		// Step 4: Criar NotifyPreferences (compensação em cascata via User)
		{
			Name: "CreateUserNotify",
			Execute: func(ctx context.Context, data map[string]interface{}) error {
				command := data["command"].(appdto.RegisterConfirmCommand)
				if company, ok := data["company"].(appdto.MSCompanyResponseDTO); ok {
					ctx = client.WithCompanyID(ctx, company.ID)
				}
				user := data["user"].(appdto.MSUserResponseDTO)

				notifyRequest := appdto.MSUserNotifyCreateRequestDTO{
					UserID:          user.ID,
					EmailEnabled:    true,
					SmsEnabled:      true,
					PushEnabled:     true,
					WhatsAppEnabled: true,
				}
				_, err := uc.userClient.CreateUserNotify(ctx, notifyRequest, command.TenantId, command.CorrelationID)
				return err
			},
			Compensate: nil, // removed together with the user profile by the ms-user reaper
			MaxRetries: 1,   // Sem retry - delegado ao decorator se necessário
			Timeout:    5 * time.Second,
		},
		// Step 5: Criar AuthUser (COM compensação)
		{
			Name: "CreateAuthUser",
			Execute: func(ctx context.Context, data map[string]interface{}) error {
				command := data["command"].(appdto.RegisterConfirmCommand)
				if company, ok := data["company"].(appdto.MSCompanyResponseDTO); ok {
					ctx = client.WithCompanyID(ctx, company.ID)
				}
				company := data["company"].(appdto.MSCompanyResponseDTO)
				confirmResponse := data["confirmResponse"].(appdto.MSUserRegisterConfirmResponseDTO)
				user := data["user"].(appdto.MSUserResponseDTO)

				authUserRequest := appdto.AuthUserCreateRequestDTO{
					Username:       user.Email,
					Email:          user.Email,
					Password:       confirmResponse.PasswordHash,
					IDUserExternal: user.ID,
					CodeUser:       user.CodeUser,
					CompanyID:      user.CompanyID,
					CompanyCode:    company.CodeCompany,
					TenantId:       command.TenantId,
				}
				_, err := uc.authClient.CreateUser(ctx, authUserRequest, command.TenantId, command.CorrelationID)
				if err == nil {
					return nil
				}
				if !uc.opts.SagaHardening || !isAmbiguousOutcome(err) {
					return err
				}
				return uc.recoverAuthUserCreate(ctx, err, user.ID, command)
			},
			// Cleanup delegated to the ms-auth reaper (PENDING_ACTIVATION account expires in 15 min).
			Compensate: uc.delegateToReaper("CreateAuthUser", "ms-auth"),
			MaxRetries: 1, // Sem retry - delegado ao decorator se necessário
			Timeout:    10 * time.Second,
		},
		// Step 6: Aceitar Consentimentos (COM compensação)
		{
			Name: "AcceptConsents",
			Execute: func(ctx context.Context, data map[string]interface{}) error {
				command := data["command"].(appdto.RegisterConfirmCommand)
				if company, ok := data["company"].(appdto.MSCompanyResponseDTO); ok {
					ctx = client.WithCompanyID(ctx, company.ID)
				}
				confirmResponse := data["confirmResponse"].(appdto.MSUserRegisterConfirmResponseDTO)
				user := data["user"].(appdto.MSUserResponseDTO)

				acceptAllRequest := appdto.UserConsentAcceptAllRequestDTO{
					UserID:      user.ID,
					Email:       user.Email,
					AcceptedAt:  time.Now(),
					Geolocation: confirmResponse.Geolocation,
				}
				_, err := uc.userConsentClient.AcceptAll(ctx, acceptAllRequest, command.TenantId, command.CorrelationID)
				return err
			},
			Compensate: func(ctx context.Context, data map[string]interface{}) error {
				command := data["command"].(appdto.RegisterConfirmCommand)
				user := data["user"].(appdto.MSUserResponseDTO)
				return uc.userConsentClient.DeleteAllByUserId(ctx, user.ID, command.TenantId, command.CorrelationID)
			},
			MaxRetries: 1, // Sem retry - delegado ao decorator se necessário
			// accept-all no ms-user-consents costuma levar ~4s+; 5s gerava falso timeout e retries sem sessão Redis
			Timeout: 30 * time.Second,
		},
		// Step 7: Fazer Login (gera token, sem compensação)
		{
			Name: "ExecuteLogin",
			Execute: func(ctx context.Context, data map[string]interface{}) error {
				command := data["command"].(appdto.RegisterConfirmCommand)
				if company, ok := data["company"].(appdto.MSCompanyResponseDTO); ok {
					ctx = client.WithCompanyID(ctx, company.ID)
				}
				confirmResponse := data["confirmResponse"].(appdto.MSUserRegisterConfirmResponseDTO)
				user := data["user"].(appdto.MSUserResponseDTO)

				registerLoginRequest := appdto.AuthRegisterLoginRequestDTO{
					Username:     user.Email,
					PasswordHash: confirmResponse.PasswordHash,
					TenantId:     command.TenantId,
					CompanyID:    user.CompanyID,
				}
				loginResponse, err := uc.authClient.RegisterLogin(ctx, registerLoginRequest, command.TenantId, command.CorrelationID, command.ClientId)
				if err != nil {
					return err
				}
				data["loginResponse"] = loginResponse
				return nil
			},
			Compensate: nil, // token expira naturalmente
			MaxRetries: 1,   // Sem retry - delegado ao decorator se necessário
			Timeout:    15 * time.Second,
		},
	}

	// Métricas por passo e por compensação (o executor não conhece métricas).
	for i := range steps {
		steps[i].Execute = uc.observeStep(steps[i].Name, steps[i].Execute)
		if steps[i].Compensate != nil && !delegatedSteps[steps[i].Name] {
			steps[i].Compensate = uc.observeCompensation(steps[i].Name, steps[i].Compensate)
		}
	}

	return saga.InMemorySaga{
		Name:  "RegisterConfirmSaga",
		Steps: steps,
	}
}

// defaultRegisterLocale is the locale of every registration before the country field (21/11/2026).
const defaultRegisterLocale = "pt-BR"

// registerLocale: the locale the person sent at init (already accepted by the ms-user), else pt-BR as before.
func registerLocale(fromSession string) string {
	if fromSession == "" {
		return defaultRegisterLocale
	}
	return fromSession
}

// recoverAuthUserCreate trata o passo 5 com timeout ambíguo: o ms-auth pode ter gravado a conta e a resposta
// se perdido. Consulta por id_user_external; se existe, o passo é tratado como executado (e a compensação
// do passo 5 continua disponível se a saga cair depois). Se não existe ou a consulta falha, devolve o erro original.
func (uc *registerConfirmUseCaseImpl) recoverAuthUserCreate(ctx context.Context, original error, userID string, command appdto.RegisterConfirmCommand) error {
	recoveryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), authRecoveryTimeout)
	defer cancel()

	exists, err := uc.authClient.ExistsByIDExternal(recoveryCtx, userID, command.TenantId, command.CorrelationID)
	if err != nil {
		uc.logger.Warn("Passo 5 com timeout: não foi possível confirmar a conta no ms-auth",
			zap.String("correlation_id", command.CorrelationID),
			zap.Error(err))
		return original
	}
	if !exists {
		return original
	}
	uc.logger.Warn("Passo 5 com timeout, mas a conta existe no ms-auth: tratando como executada",
		zap.String("correlation_id", command.CorrelationID))
	uc.metrics().RecordStep("CreateAuthUser", "ambiguous_recovered")
	return nil
}

// isAmbiguousOutcome reports whether the call may have been executed by the remote service even though no
// answer arrived: deadline/timeout, connection reset or broken pipe, or the response lost mid-way (EOF).
// An explicit HTTP answer (4xx/5xx) and a refused connection (request never sent) are NOT ambiguous.
func isAmbiguousOutcome(err error) bool {
	if err == nil {
		return false
	}
	var httpErr *appdto.HTTPError
	if errors.As(err, &httpErr) {
		return false
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNABORTED) || errors.Is(err, syscall.EPIPE) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "read" {
		return true
	}
	return false
}

// observeStep conta o resultado de um passo (ok | error).
func (uc *registerConfirmUseCaseImpl) observeStep(name string, fn func(context.Context, map[string]interface{}) error) func(context.Context, map[string]interface{}) error {
	metrics := metricsOrNoop(uc.opts.Metrics)
	return func(ctx context.Context, data map[string]interface{}) error {
		err := fn(ctx, data)
		if err != nil {
			metrics.RecordStep(name, "error")
			return err
		}
		metrics.RecordStep(name, "ok")
		return nil
	}
}

// delegatedSteps are the saga steps whose cleanup belongs to the owning service's reaper: the bff-core no longer
// calls destructive routes of ms-auth or ms-user (spec-cadastro-saga-reaper §4.4/§4.6).
var delegatedSteps = map[string]bool{"CreateUser": true, "CreateAuthUser": true}

// delegateToReaper returns a compensation that only logs and counts result="delegated_to_reaper". It never fails,
// so the saga keeps compensating the remaining steps and returns the original error.
func (uc *registerConfirmUseCaseImpl) delegateToReaper(step, owner string) func(context.Context, map[string]interface{}) error {
	return func(_ context.Context, data map[string]interface{}) error {
		correlationID := ""
		if command, ok := data["command"].(appdto.RegisterConfirmCommand); ok {
			correlationID = command.CorrelationID
		}
		uc.logger.Info("Registration compensation delegated to reaper",
			zap.String("step", step),
			zap.String("owner", owner),
			zap.String("correlation_id", correlationID))
		metricsOrNoop(uc.opts.Metrics).RecordCompensation(step, "delegated_to_reaper")
		return nil
	}
}

// observeCompensation conta o resultado de uma compensação (ok | error).
func (uc *registerConfirmUseCaseImpl) observeCompensation(name string, fn func(context.Context, map[string]interface{}) error) func(context.Context, map[string]interface{}) error {
	metrics := metricsOrNoop(uc.opts.Metrics)
	return func(ctx context.Context, data map[string]interface{}) error {
		err := fn(ctx, data)
		if err != nil {
			metrics.RecordCompensation(name, "error")
			return err
		}
		metrics.RecordCompensation(name, "ok")
		return nil
	}
}

// sendWelcomeEmail envia email de boas-vindas (fora do SAGA)
func (uc *registerConfirmUseCaseImpl) sendWelcomeEmail(ctx context.Context, sagaData map[string]interface{}, tenantId, correlationID string) {
	company := sagaData["company"].(appdto.MSCompanyResponseDTO)
	user := sagaData["user"].(appdto.MSUserResponseDTO)
	confirmResponse := sagaData["confirmResponse"].(appdto.MSUserRegisterConfirmResponseDTO)

	variables := map[string]string{
		"userName": confirmResponse.NameFull,
		"appName":  company.Name,
	}
	interfaceVariables := make(map[string]interface{})
	for k, v := range variables {
		interfaceVariables[k] = v
	}

	messageReq := outport.MessageDTO{
		CompanyID:         company.ID,
		CorrelationID:     correlationID,
		XCorrelationID:    correlationID,
		MessageType:       enums.MessageTypeEmail.String(),
		CommunicationType: enums.CommunicationTypeEmail.String(),
		TemplateType:      enums.TemplateTypeCadastroSucesso.String(),
		Recipient:         confirmResponse.Email,
		CodeUser:          user.CodeUser,
		Variables:         interfaceVariables,
	}

	err := uc.messagePublisher.PublishMessage(ctx, messageReq)
	if err != nil {
		uc.logger.Error("Erro ao enviar email de boas-vindas",
			zap.String("code_user", user.CodeUser),
			zap.String("correlation_id", correlationID),
			zap.Error(err))
	} else {
		uc.logger.Info("Email de boas-vindas enviado com sucesso",
			zap.String("code_user", user.CodeUser),
			zap.String("correlation_id", correlationID))
	}
}
