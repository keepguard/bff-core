package decorator

import (
	"context"
	"time"

	outport "github.com/keepguard/bff-core/internal/application/port/out"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

// metricsDecorator implementa MessagePublisher com métricas
type metricsDecorator struct {
	inner   outport.MessagePublisher
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// NewMetricsDecorator cria um novo decorator de métricas
func NewMetricsDecorator(inner outport.MessagePublisher, metrics *metrics.Metrics, logger *zap.Logger) outport.MessagePublisher {
	return &metricsDecorator{
		inner:   inner,
		metrics: metrics,
		logger:  logger,
	}
}

// PublishMessage implementa PublishMessage com métricas
func (d *metricsDecorator) PublishMessage(ctx context.Context, message outport.MessageDTO) error {
	start := time.Now()

	err := d.inner.PublishMessage(ctx, message)
	duration := time.Since(start)

	// Determinar status da publicação
	status := "success"
	if err != nil {
		status = "failure"
	}

	// Registrar métricas
	// Usar valores padrão para exchange e routing key se não estiverem disponíveis
	exchange := "ms-communication-exchange"
	routingKey := "communication.message.send"

	d.metrics.RecordRabbitMQPublish(exchange, routingKey, status, duration)

	if err != nil {
		d.logger.Error("Falha na publicação registrada nas métricas",
			zap.String("status", status),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		d.logger.Debug("Publicação registrada nas métricas",
			zap.String("status", status),
			zap.Duration("duration", duration),
		)
	}

	return err
}

// Close implementa Close
func (d *metricsDecorator) Close() error {
	return d.inner.Close()
}
