package register

import (
	"context"
	"fmt"
	"time"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	client "github.com/keepguard/bff-core/internal/application/port"
	outport "github.com/keepguard/bff-core/internal/application/port/out"
	"github.com/keepguard/bff-core/internal/domain/enums"
	"go.uber.org/zap"
)

// registerInitUseCaseImpl implementa o caso de uso de inicialização de registro
type registerInitUseCaseImpl struct {
	authClient          client.AuthClient
	userClient          client.UserClient
	companyClient       client.CompanyClient
	communicationClient client.CommunicationClient
	messagePublisher    outport.MessagePublisher
	logger              *zap.Logger
	opts                RegisterOptions
}

// NewRegisterInitUseCase cria um novo caso de uso de inicialização de registro
func NewRegisterInitUseCase(
	authClient client.AuthClient,
	userClient client.UserClient,
	companyClient client.CompanyClient,
	communicationClient client.CommunicationClient,
	messagePublisher outport.MessagePublisher,
	logger *zap.Logger,
	opts RegisterOptions,
) RegisterInitUseCase {
	return &registerInitUseCaseImpl{
		authClient:          authClient,
		userClient:          userClient,
		companyClient:       companyClient,
		communicationClient: communicationClient,
		messagePublisher:    messagePublisher,
		logger:              logger,
		opts:                opts,
	}
}

// Execute executa o caso de uso de inicialização de registro
func (uc *registerInitUseCaseImpl) Execute(ctx context.Context, command appdto.RegisterInitCommand) (appdto.RegisterInitViewDTO, error) {

	// Passo 1: Verificar se a empresa existe consultando o Company Service
	company, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
	if err != nil {
		return appdto.RegisterInitViewDTO{}, err
	}

	// Passo 2: Criar requisição de registro
	registerRequest := appdto.MSUserRegisterInitRequestDTO{
		CompanyID:                  company.ID,
		Email:                      command.Email,
		NameFull:                   command.NameFull,
		Password:                   command.Password,
		Phone:                      command.Phone,
		HasAcceptedTermsAndPrivacy: command.HasAcceptedTermsAndPrivacy,
		AcceptedMarketing:          command.AcceptedMarketing,
		IPAddress:                  command.IPAddress,
		UserAgent:                  command.UserAgent,
		Geolocation:                command.Geolocation,
		Type:                       command.Type,
	}

	// Passo 3: Inicializar registro no User Service
	registerResponse, err := uc.userClient.InitRegister(client.WithCompanyID(ctx, company.ID), registerRequest, command.TenantId, command.CorrelationID)
	if err != nil {
		return appdto.RegisterInitViewDTO{}, err
	}

	// Passo 4: Preparar variáveis para o template
	variables := map[string]string{
		"userName":  command.NameFull,
		"token":     registerResponse.Token,
		"expiresIn": fmt.Sprintf("%d", registerResponse.ExpiresIn/60),
		"appName":   company.Name,
	}

	interfaceVariables := make(map[string]interface{})
	for k, v := range variables {
		interfaceVariables[k] = v
	}

	// Passo 5: Enviar mensagem através dos canais de MFA habilitados na Empresa
	// Se a empresa não tiver canais customizados definidos, o padrão é EMAIL
	hasCustomChannels := len(company.MfaChannels) > 0
	emailSent := false
	smsSent := false
	var requiredChannels []string

	if hasCustomChannels {
		for _, mfaChannel := range company.MfaChannels {
			if !mfaChannel.Enabled || !mfaChannel.Required {
				continue
			}

			requiredChannels = append(requiredChannels, mfaChannel.Channel)

			switch mfaChannel.Channel {
			case "EMAIL":
				if !emailSent {
					emailToken := registerResponse.EmailToken
					if emailToken == "" {
						emailToken = registerResponse.Token
					}
					emailVars := make(map[string]interface{})
					for k, v := range variables {
						emailVars[k] = v
					}
					emailVars["token"] = emailToken

					emailReq := outport.MessageDTO{
						CompanyID:         company.ID,
						CorrelationID:     command.CorrelationID,
						XCorrelationID:    command.CorrelationID,
						MessageType:       enums.MessageTypeEmail.String(),
						CommunicationType: enums.CommunicationTypeEmail.String(),
						TemplateType:      enums.TemplateTypeAutenticacaoEmailToken.String(),
						Recipient:         registerResponse.Email,
						CodeUser:          registerResponse.RegistrationSessionID,
						Variables:         emailVars,
					}
					uc.sendCode(ctx, emailReq, "EMAIL")
					emailSent = true
				}
			case "SMS":
				if !smsSent && command.Phone != "" {
					smsToken := registerResponse.SmsToken
					if smsToken == "" {
						smsToken = registerResponse.Token
					}
					smsVars := make(map[string]interface{})
					for k, v := range variables {
						smsVars[k] = v
					}
					smsVars["token"] = smsToken

					smsReq := outport.MessageDTO{
						CompanyID:         company.ID,
						CorrelationID:     command.CorrelationID,
						XCorrelationID:    command.CorrelationID,
						MessageType:       enums.MessageTypeSMS.String(),
						CommunicationType: enums.CommunicationTypeSMS.String(),
						TemplateType:      enums.TemplateTypeAutenticacaoSMSToken.String(),
						Recipient:         command.Phone,
						CodeUser:          registerResponse.RegistrationSessionID,
						Variables:         smsVars,
					}
					uc.sendCode(ctx, smsReq, "SMS")
					smsSent = true
				}
			case "WHATSAPP":
				if command.Phone != "" {
					whatsToken := registerResponse.WhatsAppToken
					if whatsToken == "" {
						whatsToken = registerResponse.Token
					}
					whatsVars := make(map[string]interface{})
					for k, v := range variables {
						whatsVars[k] = v
					}
					whatsVars["token"] = whatsToken

					whatsReq := outport.MessageDTO{
						CompanyID:         company.ID,
						CorrelationID:     command.CorrelationID,
						XCorrelationID:    command.CorrelationID,
						MessageType:       enums.MessageTypeWhatsApp.String(),
						CommunicationType: enums.CommunicationTypeWhatsApp.String(),
						TemplateType:      enums.TemplateTypeAutenticacaoWhatsAppToken.String(),
						Recipient:         command.Phone,
						CodeUser:          registerResponse.RegistrationSessionID,
						Variables:         whatsVars,
					}
					uc.sendCode(ctx, whatsReq, "WHATSAPP")
				}
			}
		}
	} else {
		// Fallback default: Envia EMAIL
		requiredChannels = append(requiredChannels, "EMAIL")
		messageReq := outport.MessageDTO{
			CompanyID:         company.ID,
			CorrelationID:     command.CorrelationID,
			XCorrelationID:    command.CorrelationID,
			MessageType:       enums.MessageTypeEmail.String(),
			CommunicationType: enums.CommunicationTypeEmail.String(),
			TemplateType:      enums.TemplateTypeAutenticacaoEmailToken.String(),
			Recipient:         registerResponse.Email,
			CodeUser:          registerResponse.RegistrationSessionID,
			Variables:         interfaceVariables,
		}
		uc.sendCode(ctx, messageReq, "EMAIL")
	}

	// Passo 7: Retornar resposta com canais exigidos
	response := appdto.RegisterInitViewDTO{
		RegistrationSessionID: registerResponse.RegistrationSessionID,
		Email:                 registerResponse.Email,
		Phone:                 command.Phone,
		ExpiresIn:             registerResponse.ExpiresIn,
		RequiredChannels:      requiredChannels,
	}
	if uc.opts.SagaHardening {
		// Aditivo: o front sabe que o reenvio (resend) é a saída se o código não chegou.
		resendAvailable := true
		response.ResendAvailable = &resendAvailable
	}

	return response, nil
}

// sendCodeAttempts: com saga_hardening, o envio do código tenta 3 vezes com backoff curto (200 ms, 400 ms).
// Sem a flag, uma tentativa só, como antes (o erro é contado e logado, mas não muda a resposta).
const sendCodeAttempts = 3

// sendCode envia um código de cadastro. Falha não derruba o init (a resposta segue 201), mas é contada e logada
// sem dado pessoal (sem destinatário, código ou template com valores).
func (uc *registerInitUseCaseImpl) sendCode(ctx context.Context, msg outport.MessageDTO, channel string) {
	attempts := 1
	if uc.opts.SagaHardening {
		attempts = sendCodeAttempts
	}
	backoff := 200 * time.Millisecond
	metrics := metricsOrNoop(uc.opts.Metrics)

	for attempt := 1; attempt <= attempts; attempt++ {
		err := uc.messagePublisher.PublishMessage(ctx, msg)
		if err == nil {
			metrics.RecordStep("InitSendCode", "ok")
			return
		}
		metrics.RecordStep("InitSendCode", "error")
		uc.logger.Warn("Envio do código de cadastro falhou",
			zap.String("channel", channel),
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", attempts),
			zap.String("correlation_id", msg.CorrelationID),
			zap.Error(err))

		if attempt < attempts {
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff *= 2
		}
	}
}
