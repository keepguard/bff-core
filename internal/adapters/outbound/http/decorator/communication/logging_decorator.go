package communication

import (
	"context"
	"time"

	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"go.uber.org/zap"
)

// communicationLoggingDecorator implementa logging para CommunicationClient
type communicationLoggingDecorator struct {
	inner       portsclient.CommunicationClient
	logger      *zap.Logger
	serviceName string
}

// NewCommunicationLoggingDecorator cria um decorator de logging para CommunicationClient
func NewCommunicationLoggingDecorator(
	inner portsclient.CommunicationClient,
	logger *zap.Logger,
	serviceName string,
) portsclient.CommunicationClient {
	return &communicationLoggingDecorator{
		inner:       inner,
		logger:      logger,
		serviceName: serviceName,
	}
}

// SendNotification implementa SendNotification com logging
func (d *communicationLoggingDecorator) SendNotification(ctx context.Context, req portsclient.SendNotificationRequestDTO, xApplication, correlationID string) error {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "SendNotification"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.String("recipient", req.Recipient),
		zap.String("templateType", req.TemplateType),
		zap.String("channel", req.Channel),
	)

	err := d.inner.SendNotification(ctx, req, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "SendNotification"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
			zap.String("recipient", req.Recipient),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Notificação enviada com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "SendNotification"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.String("recipient", req.Recipient),
		zap.Duration("duration", duration),
	)

	return nil
}

// SendMessage implementa SendMessage com logging
func (d *communicationLoggingDecorator) SendMessage(ctx context.Context, req communicationDto.SendMessageRequestDTO, xApplication, correlationID string) (communicationDto.SendMessageResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "SendMessage"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.String("recipient", req.Recipient),
		zap.String("messageType", req.MessageType),
		zap.String("templateType", req.TemplateType),
		zap.String("communicationType", req.CommunicationType),
	)

	response, err := d.inner.SendMessage(ctx, req, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "SendMessage"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
			zap.String("recipient", req.Recipient),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Mensagem enviada com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "SendMessage"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.String("recipient", req.Recipient),
		zap.Bool("success", response.Success),
		zap.Duration("duration", duration),
	)

	return response, nil
}
