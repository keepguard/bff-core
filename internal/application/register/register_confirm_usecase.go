package register

import (
	"context"
	"fmt"
	"time"

	"github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	userConsentDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user_consent"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/enums"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/domain/ports/messaging"
	"github.com/keepguard/bff-core/internal/domain/saga"
	"go.uber.org/zap"
)

// registerConfirmUseCaseImpl implementa o caso de uso de confirmação de registro
type registerConfirmUseCaseImpl struct {
	userClient          client.UserClient
	companyClient       client.CompanyClient
	authClient          client.AuthClient
	userConsentClient   client.UserConsentClient
	communicationClient client.CommunicationClient
	messagePublisher    messaging.MessagePublisher
	sagaExecutor        *saga.InMemorySagaExecutor
	logger              *zap.Logger
}

// NewRegisterConfirmUseCase cria um novo caso de uso de confirmação de registro
func NewRegisterConfirmUseCase(
	userClient client.UserClient,
	companyClient client.CompanyClient,
	authClient client.AuthClient,
	userConsentClient client.UserConsentClient,
	communicationClient client.CommunicationClient,
	messagePublisher messaging.MessagePublisher,
	logger *zap.Logger,
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
	}
}

// Execute executa o caso de uso de confirmação de registro usando SAGA em memória
func (uc *registerConfirmUseCaseImpl) Execute(command appdto.RegisterConfirmCommand) (dto.RegisterConfirmResponseDTO, error) {
	uc.logger.Info("Iniciando confirmação de registro com SAGA",
		zap.String("email", command.Email),
		zap.String("correlation_id", command.CorrelationID))

	// Preparar dados para o SAGA
	sagaData := map[string]interface{}{
		"command": command,
	}

	// Construir e executar SAGA
	registerSaga := uc.buildRegisterConfirmSaga()

	err := uc.sagaExecutor.Execute(command.Context, registerSaga, sagaData)
	if err != nil {
		uc.logger.Error("SAGA de registro falhou",
			zap.String("email", command.Email),
			zap.String("correlation_id", command.CorrelationID),
			zap.Error(err))
		return dto.RegisterConfirmResponseDTO{}, err
	}

	// Extrair resultado do login do sagaData
	loginResponse, ok := sagaData["loginResponse"].(authDto.AuthLoginResponseDTO)
	if !ok {
		return dto.RegisterConfirmResponseDTO{}, fmt.Errorf("resposta de login não encontrada no SAGA")
	}

	// Enviar email de boas-vindas APENAS se SAGA completou com sucesso
	uc.sendWelcomeEmail(command.Context, sagaData, command.TenantId, command.CorrelationID)

	uc.logger.Info("Registro confirmado com sucesso via SAGA",
		zap.String("email", command.Email),
		zap.String("correlation_id", command.CorrelationID))

	return dto.RegisterConfirmResponseDTO{
		Token:          loginResponse.Token,
		TokenExpiresIn: loginResponse.ExpiresIn,
	}, nil
}

// buildRegisterConfirmSaga constrói os steps da SAGA de confirmação de registro
func (uc *registerConfirmUseCaseImpl) buildRegisterConfirmSaga() saga.InMemorySaga {
	return saga.InMemorySaga{
		Name: "RegisterConfirmSaga",
		Steps: []saga.InMemoryStep{
			// Step 1: Validar Company (read-only, sem compensação)
			{
				Name: "ValidateCompany",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					command := data["command"].(appdto.RegisterConfirmCommand)
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
					confirmRequest := userDto.MSUserRegisterConfirmRequestDTO{
						Email:                 command.Email,
						RegistrationSessionID: command.RegistrationSessionId,
						Token:                 command.Token,
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
					company := data["company"].(companyDto.MSCompanyResponseDTO)
					confirmResponse := data["confirmResponse"].(userDto.MSUserRegisterConfirmResponseDTO)

					userRequest := userDto.MSUserCreateRequestDTO{
						CompanyID:       company.ID,
						Type:            confirmResponse.Type,
						Email:           confirmResponse.Email,
						PhoneE164:       confirmResponse.Phone,
						PreferredLocale: "pt-BR",
						Timezone:        "America/Sao_Paulo",
						Status:          "ACTIVE",
						PersonProfile: &userDto.PersonProfileDTO{
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
				Compensate: func(ctx context.Context, data map[string]interface{}) error {
					command := data["command"].(appdto.RegisterConfirmCommand)
					user := data["user"].(userDto.MSUserResponseDTO)
					return uc.userClient.DeleteUser(ctx, user.ID, command.TenantId, command.CorrelationID)
				},
				MaxRetries: 1, // Sem retry - delegado ao decorator se necessário
				Timeout:    10 * time.Second,
			},
			// Step 4: Criar NotifyPreferences (compensação em cascata via User)
			{
				Name: "CreateUserNotify",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					command := data["command"].(appdto.RegisterConfirmCommand)
					user := data["user"].(userDto.MSUserResponseDTO)

					notifyRequest := userDto.MSUserNotifyCreateRequestDTO{
						UserID:          user.ID,
						EmailEnabled:    true,
						SmsEnabled:      true,
						PushEnabled:     true,
						WhatsAppEnabled: true,
					}
					_, err := uc.userClient.CreateUserNotify(ctx, notifyRequest, command.TenantId, command.CorrelationID)
					return err
				},
				Compensate: nil, // compensação em cascata via DELETE do User
				MaxRetries: 1,   // Sem retry - delegado ao decorator se necessário
				Timeout:    5 * time.Second,
			},
			// Step 5: Criar AuthUser (COM compensação)
			{
				Name: "CreateAuthUser",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					command := data["command"].(appdto.RegisterConfirmCommand)
					company := data["company"].(companyDto.MSCompanyResponseDTO)
					confirmResponse := data["confirmResponse"].(userDto.MSUserRegisterConfirmResponseDTO)
					user := data["user"].(userDto.MSUserResponseDTO)

					authUserRequest := authDto.AuthUserCreateRequestDTO{
						Username:       user.Email,
						Email:          user.Email,
						Password:       confirmResponse.PasswordHash,
						IDUserExternal: user.ID,
						CodeUser:       user.CodeUser,
						CompanyID:      user.CompanyID,
						CompanyCode:    company.CodeCompany,
						TenantId:   command.TenantId,
					}
					_, err := uc.authClient.CreateUser(ctx, authUserRequest, command.TenantId, command.CorrelationID)
					return err
				},
				Compensate: func(ctx context.Context, data map[string]interface{}) error {
					command := data["command"].(appdto.RegisterConfirmCommand)
					user := data["user"].(userDto.MSUserResponseDTO)
					return uc.authClient.HardDeleteUser(ctx, user.ID, command.TenantId, command.CorrelationID)
				},
				MaxRetries: 1, // Sem retry - delegado ao decorator se necessário
				Timeout:    10 * time.Second,
			},
			// Step 6: Aceitar Consentimentos (COM compensação)
			{
				Name: "AcceptConsents",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					command := data["command"].(appdto.RegisterConfirmCommand)
					confirmResponse := data["confirmResponse"].(userDto.MSUserRegisterConfirmResponseDTO)
					user := data["user"].(userDto.MSUserResponseDTO)

					acceptAllRequest := userConsentDto.UserConsentAcceptAllRequestDTO{
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
					user := data["user"].(userDto.MSUserResponseDTO)
					return uc.userConsentClient.DeleteAllByUserId(ctx, user.ID, command.TenantId, command.CorrelationID)
				},
				MaxRetries: 1, // Sem retry - delegado ao decorator se necessário
				Timeout:    5 * time.Second,
			},
			// Step 7: Fazer Login (gera token, sem compensação)
			{
				Name: "ExecuteLogin",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					command := data["command"].(appdto.RegisterConfirmCommand)
					confirmResponse := data["confirmResponse"].(userDto.MSUserRegisterConfirmResponseDTO)
					user := data["user"].(userDto.MSUserResponseDTO)

					registerLoginRequest := authDto.AuthRegisterLoginRequestDTO{
						Username:     user.Email,
						PasswordHash: confirmResponse.PasswordHash,
						TenantId: command.TenantId,
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
				Timeout:    5 * time.Second,
			},
		},
	}
}

// sendWelcomeEmail envia email de boas-vindas (fora do SAGA)
func (uc *registerConfirmUseCaseImpl) sendWelcomeEmail(ctx context.Context, sagaData map[string]interface{}, tenantId, correlationID string) {
	company := sagaData["company"].(companyDto.MSCompanyResponseDTO)
	user := sagaData["user"].(userDto.MSUserResponseDTO)
	confirmResponse := sagaData["confirmResponse"].(userDto.MSUserRegisterConfirmResponseDTO)

	variables := map[string]string{
		"userName": confirmResponse.NameFull,
		"appName":  company.Name,
	}
	interfaceVariables := make(map[string]interface{})
	for k, v := range variables {
		interfaceVariables[k] = v
	}

	messageReq := messaging.MessageDTO{
		TenantId:      tenantId,
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
			zap.String("email", confirmResponse.Email),
			zap.String("code_user", user.CodeUser),
			zap.Error(err))
	} else {
		uc.logger.Info("Email de boas-vindas enviado com sucesso",
			zap.String("email", confirmResponse.Email),
			zap.String("code_user", user.CodeUser))
	}
}
