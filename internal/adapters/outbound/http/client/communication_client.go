package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// communicationClient implementa CommunicationClient usando HTTP
type communicationClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewCommunicationClient cria uma nova instância do CommunicationClient
func NewCommunicationClient(config *config.Config, logger *zap.Logger) client.CommunicationClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(2)
	httpClient.SetRetryWaitTime(200 * time.Millisecond)

	return &communicationClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// SendNotification envia uma notificação
func (c *communicationClient) SendNotification(ctx context.Context, req client.SendNotificationRequestDTO, xApplication, correlationID string) error {
	url := fmt.Sprintf("%s/api/v1/notifications/send", c.config.Services.Communication.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return fmt.Errorf("erro ao comunicar com communication service: %w", err)
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "communication service")
	}

	return nil
}

// SendMessage envia uma mensagem através do ms-communication
func (c *communicationClient) SendMessage(ctx context.Context, req communicationDto.SendMessageRequestDTO, xApplication, correlationID string) (communicationDto.SendMessageResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/messages/send", c.config.Services.Communication.BaseURL)

	var response communicationDto.SendMessageResponseDTO
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&response).
		Post(url)

	if err != nil {
		return communicationDto.SendMessageResponseDTO{}, fmt.Errorf("erro ao comunicar com ms-communication: %w", err)
	}

	// Trata erros HTTP do ms-communication
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return communicationDto.SendMessageResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "ms-communication service")
	}

	return response, nil
}
