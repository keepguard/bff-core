package client

import (
	"context"

	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
)

// SendNotificationRequestDTO representa a requisição para enviar notificação
type SendNotificationRequestDTO struct {
	UserID       string            `json:"userId"`
	TemplateType string            `json:"templateType"`
	Channel      string            `json:"channel"`
	Recipient    string            `json:"recipient"`
	Data         map[string]string `json:"data"`
}

// CommunicationClient interface para comunicação com o serviço de comunicação
type CommunicationClient interface {
	// SendNotification envia uma notificação
	SendNotification(ctx context.Context, req SendNotificationRequestDTO, tenantId, correlationID string) error
	// SendMessage envia uma mensagem através do ms-communication
	SendMessage(ctx context.Context, req communicationDto.SendMessageRequestDTO, tenantId, correlationID string) (communicationDto.SendMessageResponseDTO, error)
}
