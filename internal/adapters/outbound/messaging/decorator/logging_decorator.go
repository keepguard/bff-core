package decorator

import (
	"context"
	"time"

	"github.com/keepguard/bff-core/internal/domain/ports/messaging"
	"go.uber.org/zap"
)

// loggingDecorator implementa MessagePublisher com logging
type loggingDecorator struct {
	inner  messaging.MessagePublisher
	logger *zap.Logger
}

// NewLoggingDecorator cria um novo decorator de logging
func NewLoggingDecorator(inner messaging.MessagePublisher, logger *zap.Logger) messaging.MessagePublisher {
	return &loggingDecorator{
		inner:  inner,
		logger: logger,
	}
}

// PublishMessage implementa PublishMessage com logging
func (d *loggingDecorator) PublishMessage(ctx context.Context, message messaging.MessageDTO) error {
	start := time.Now()

	d.logger.Info("Iniciando publicação de mensagem",
		zap.String("operation", "PublishMessage"),
		zap.String("correlationID", message.XCorrelationID),
		zap.String("tenantId", message.TenantId),
		zap.String("recipient", message.Recipient),
		zap.String("messageType", message.MessageType),
		zap.String("templateType", message.TemplateType),
		zap.String("communicationType", message.CommunicationType),
		zap.String("codeUser", message.CodeUser),
	)

	err := d.inner.PublishMessage(ctx, message)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na publicação de mensagem",
			zap.String("operation", "PublishMessage"),
			zap.String("correlationID", message.XCorrelationID),
			zap.String("tenantId", message.TenantId),
			zap.String("recipient", message.Recipient),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Mensagem publicada com sucesso",
		zap.String("operation", "PublishMessage"),
		zap.String("correlationID", message.XCorrelationID),
		zap.String("tenantId", message.TenantId),
		zap.String("recipient", message.Recipient),
		zap.Duration("duration", duration),
	)

	return nil
}

// Close implementa Close com logging
func (d *loggingDecorator) Close() error {
	d.logger.Info("Fechando publisher de mensagens")
	return d.inner.Close()
}
