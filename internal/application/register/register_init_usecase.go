package register

import (
	"fmt"

	"github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/enums"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/domain/ports/messaging"
	"go.uber.org/zap"
)

// registerInitUseCaseImpl implementa o caso de uso de inicialização de registro
type registerInitUseCaseImpl struct {
	authClient          client.AuthClient
	userClient          client.UserClient
	companyClient       client.CompanyClient
	communicationClient client.CommunicationClient
	messagePublisher    messaging.MessagePublisher
	logger              *zap.Logger
}

// NewRegisterInitUseCase cria um novo caso de uso de inicialização de registro
func NewRegisterInitUseCase(
	authClient client.AuthClient,
	userClient client.UserClient,
	companyClient client.CompanyClient,
	communicationClient client.CommunicationClient,
	messagePublisher messaging.MessagePublisher,
	logger *zap.Logger,
) RegisterInitUseCase {
	return &registerInitUseCaseImpl{
		authClient:          authClient,
		userClient:          userClient,
		companyClient:       companyClient,
		communicationClient: communicationClient,
		messagePublisher:    messagePublisher,
		logger:              logger,
	}
}

// Execute executa o caso de uso de inicialização de registro
func (uc *registerInitUseCaseImpl) Execute(command appdto.RegisterInitCommand) (dto.RegisterInitResponseDTO, error) {

	// Passo 1: Verificar se a empresa existe consultando o Company Service
	company, err := uc.companyClient.GetByTenantId(command.Context, command.TenantId, command.CorrelationID)
	if err != nil {
		return dto.RegisterInitResponseDTO{}, err
	}

	// Passo 2: Criar requisição de registro
	registerRequest := userDto.MSUserRegisterInitRequestDTO{
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
	registerResponse, err := uc.userClient.InitRegister(client.WithCompanyID(command.Context, company.ID), registerRequest, command.TenantId, command.CorrelationID)
	if err != nil {
		return dto.RegisterInitResponseDTO{}, err
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

					emailReq := messaging.MessageDTO{
						TenantId:          command.TenantId,
						CorrelationID:     command.CorrelationID,
						XCorrelationID:    command.CorrelationID,
						MessageType:       enums.MessageTypeEmail.String(),
						CommunicationType: enums.CommunicationTypeEmail.String(),
						TemplateType:      enums.TemplateTypeAutenticacaoEmailToken.String(),
						Recipient:         registerResponse.Email,
						CodeUser:          registerResponse.RegistrationSessionID,
						Variables:         emailVars,
					}
					_ = uc.messagePublisher.PublishMessage(command.Context, emailReq)
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

					smsReq := messaging.MessageDTO{
						TenantId:          command.TenantId,
						CorrelationID:     command.CorrelationID,
						XCorrelationID:    command.CorrelationID,
						MessageType:       enums.MessageTypeSMS.String(),
						CommunicationType: enums.CommunicationTypeSMS.String(),
						TemplateType:      enums.TemplateTypeAutenticacaoSMSToken.String(),
						Recipient:         command.Phone,
						CodeUser:          registerResponse.RegistrationSessionID,
						Variables:         smsVars,
					}
					_ = uc.messagePublisher.PublishMessage(command.Context, smsReq)
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

					whatsReq := messaging.MessageDTO{
						TenantId:          command.TenantId,
						CorrelationID:     command.CorrelationID,
						XCorrelationID:    command.CorrelationID,
						MessageType:       enums.MessageTypeWhatsApp.String(),
						CommunicationType: enums.CommunicationTypeWhatsApp.String(),
						TemplateType:      enums.TemplateTypeAutenticacaoWhatsAppToken.String(),
						Recipient:         command.Phone,
						CodeUser:          registerResponse.RegistrationSessionID,
						Variables:         whatsVars,
					}
					_ = uc.messagePublisher.PublishMessage(command.Context, whatsReq)
				}
			}
		}
	} else {
		// Fallback default: Envia EMAIL
		requiredChannels = append(requiredChannels, "EMAIL")
		messageReq := messaging.MessageDTO{
			TenantId:          command.TenantId,
			CorrelationID:     command.CorrelationID,
			XCorrelationID:    command.CorrelationID,
			MessageType:       enums.MessageTypeEmail.String(),
			CommunicationType: enums.CommunicationTypeEmail.String(),
			TemplateType:      enums.TemplateTypeAutenticacaoEmailToken.String(),
			Recipient:         registerResponse.Email,
			CodeUser:          registerResponse.RegistrationSessionID,
			Variables:         interfaceVariables,
		}
		_ = uc.messagePublisher.PublishMessage(command.Context, messageReq)
	}

	// Passo 7: Retornar resposta com canais exigidos
	response := dto.RegisterInitResponseDTO{
		RegistrationSessionID: registerResponse.RegistrationSessionID,
		Email:                 registerResponse.Email,
		Phone:                 command.Phone,
		ExpiresIn:             registerResponse.ExpiresIn,
		RequiredChannels:      requiredChannels,
	}

	return response, nil
}
