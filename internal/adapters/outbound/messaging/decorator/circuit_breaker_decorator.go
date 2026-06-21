package decorator

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/domain/ports/messaging"
	"github.com/keepguard/bff-core/internal/infrastructure/resilience"
	"go.uber.org/zap"
)

// circuitBreakerDecorator implementa MessagePublisher com circuit breaker e fallback HTTP
type circuitBreakerDecorator struct {
	inner               messaging.MessagePublisher
	circuitBreaker      *gobreaker.CircuitBreaker
	communicationClient client.CommunicationClient
	logger              *zap.Logger
}

// NewCircuitBreakerDecorator cria um novo decorator de circuit breaker
func NewCircuitBreakerDecorator(
	inner messaging.MessagePublisher,
	cbManager *resilience.CircuitBreakerManager,
	communicationClient client.CommunicationClient,
	logger *zap.Logger,
) messaging.MessagePublisher {
	
	// Configurar circuit breaker específico para RabbitMQ
	cbConfig := resilience.CircuitBreakerConfig{
		Name:        "rabbitmq-message-publisher",
		MaxRequests: 3,                // Máximo 3 requests em half-open
		Interval:    10 * time.Second, // Janela de tempo para contar falhas
		Timeout:     30 * time.Second, // Tempo para tentar novamente
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Abre o circuit se 50% das requests falharam em 2+ tentativas
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 2 && failureRatio >= 0.5
		},
	}

	circuitBreaker := cbManager.GetOrCreate("rabbitmq-message-publisher", cbConfig)

	return &circuitBreakerDecorator{
		inner:               inner,
		circuitBreaker:      circuitBreaker,
		communicationClient: communicationClient,
		logger:              logger,
	}
}

// PublishMessage implementa PublishMessage com circuit breaker e fallback
func (d *circuitBreakerDecorator) PublishMessage(ctx context.Context, message messaging.MessageDTO) error {
	// Tentar publicar via RabbitMQ primeiro
	result, err := d.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, d.inner.PublishMessage(ctx, message)
	})

	if err != nil {
		// Circuit breaker está aberto ou houve erro - usar fallback HTTP
		d.logger.Warn("Circuit breaker aberto ou erro RabbitMQ, usando fallback HTTP",
			zap.String("correlationID", message.XCorrelationID),
			zap.String("recipient", message.Recipient),
			zap.Error(err),
		)

		return d.publishViaHTTPFallback(ctx, message)
	}

	// Sucesso via RabbitMQ
	if result != nil {
		d.logger.Debug("Mensagem publicada via RabbitMQ com sucesso",
			zap.String("correlationID", message.XCorrelationID),
			zap.String("recipient", message.Recipient),
		)
	}

	return nil
}

// publishViaHTTPFallback publica mensagem via HTTP como fallback
func (d *circuitBreakerDecorator) publishViaHTTPFallback(ctx context.Context, message messaging.MessageDTO) error {
	// Converter MessageDTO para SendMessageRequestDTO
	request := communicationDto.SendMessageRequestDTO{
		MessageType:       message.MessageType,
		CommunicationType: message.CommunicationType,
		TemplateType:      message.TemplateType,
		Recipient:         message.Recipient,
		Subject:           message.Subject,
		Content:           message.Content,
		CodeUser:          message.CodeUser,
		Variables:         message.Variables,
	}

	// Chamar HTTP client
	_, err := d.communicationClient.SendMessage(ctx, request, message.XApplication, message.XCorrelationID)
	if err != nil {
		d.logger.Error("Falha no fallback HTTP",
			zap.String("correlationID", message.XCorrelationID),
			zap.String("recipient", message.Recipient),
			zap.Error(err),
		)
		return fmt.Errorf("falha no fallback HTTP: %w", err)
	}

	d.logger.Info("Mensagem enviada com sucesso via fallback HTTP",
		zap.String("correlationID", message.XCorrelationID),
		zap.String("recipient", message.Recipient),
	)

	return nil
}

// Close implementa Close
func (d *circuitBreakerDecorator) Close() error {
	return d.inner.Close()
}
