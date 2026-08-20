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

type RegisterResendUseCase interface {
	Execute(command appdto.RegisterResendCommand) (dto.RegisterResendResponseDTO, error)
}

type registerResendUseCaseImpl struct {
	userClient          client.UserClient
	companyClient       client.CompanyClient
	communicationClient client.CommunicationClient
	messagePublisher    messaging.MessagePublisher
	logger              *zap.Logger
}

func NewRegisterResendUseCase(
	userClient client.UserClient,
	companyClient client.CompanyClient,
	communicationClient client.CommunicationClient,
	messagePublisher messaging.MessagePublisher,
	logger *zap.Logger,
) RegisterResendUseCase {
	return &registerResendUseCaseImpl{
		userClient:          userClient,
		companyClient:       companyClient,
		communicationClient: communicationClient,
		messagePublisher:    messagePublisher,
		logger:              logger,
	}
}

func (uc *registerResendUseCaseImpl) Execute(command appdto.RegisterResendCommand) (dto.RegisterResendResponseDTO, error) {
	uc.logger.Info("Reenviando token de registro",
		zap.String("email", command.Email),
		zap.String("correlation_id", command.CorrelationID))

	// Passo 1: Buscar informações da empresa
	company, err := uc.companyClient.GetByTenantId(command.Context, command.TenantId, command.CorrelationID)
	if err != nil {
		return dto.RegisterResendResponseDTO{}, err
	}

	// Passo 2: Chamar ms-user para incrementar contador e validar sessão
	req := userDto.MSUserRegisterResendRequestDTO{
		Email:                 command.Email,
		RegistrationSessionID: command.RegistrationSessionID,
	}

	resp, err := uc.userClient.ResendRegisterToken(command.Context, req, command.TenantId, command.CorrelationID)
	if err != nil {
		return dto.RegisterResendResponseDTO{}, err
	}

	// Passo 3: Preparar variáveis para o template (mesmas do register_init)
	variables := map[string]string{
		"userName":  resp.NameFull,
		"token":     resp.Token,
		"expiresIn": fmt.Sprintf("%d", resp.ExpiresIn/60), // Converter para minutos
		"appName":   company.Name,
	}

	// Converte para map[string]interface{}
	interfaceVariables := make(map[string]interface{})
	for k, v := range variables {
		interfaceVariables[k] = v
	}

	// Passo 4: Enviar email com token usando novo template RESEND
	messageReq := messaging.MessageDTO{
		TenantId:      command.TenantId,
		XCorrelationID:    command.CorrelationID,
		MessageType:       enums.MessageTypeEmail.String(),
		CommunicationType: enums.CommunicationTypeEmail.String(),
		TemplateType:      enums.TemplateTypeAutenticacaoEmailTokenResend.String(),
		Recipient:         resp.Email,
		CodeUser:          resp.RegistrationSessionID,
		Variables:         interfaceVariables,
	}

	err = uc.messagePublisher.PublishMessage(command.Context, messageReq)
	if err != nil {
		// Não falha o reenvio se o email não for enviado
		uc.logger.Error("Erro ao enviar email de reenvio de token",
			zap.String("email", resp.Email),
			zap.Error(err))
	} else {
		uc.logger.Info("Email de reenvio enviado com sucesso",
			zap.String("email", resp.Email))
	}

	// Passo 5: Retornar resposta
	return dto.RegisterResendResponseDTO{
		Message:                 resp.Message,
		ResendAttemptsRemaining: resp.ResendAttemptsRemaining,
		ExpiresIn:               resp.ExpiresIn,
	}, nil
}
