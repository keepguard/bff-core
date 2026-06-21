package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/keepguard/bff-core/internal/domain/ports/messaging"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

// messagePublisher implementa MessagePublisher usando RabbitMQ
type messagePublisher struct {
	publisher *rabbitmq.Publisher
	config    *config.RabbitMQConfig
	logger    *zap.Logger
}

// NewMessagePublisher cria uma nova instância do MessagePublisher
func NewMessagePublisher(cfg *config.RabbitMQConfig, logger *zap.Logger) (messaging.MessagePublisher, error) {
	// Configurar conexão RabbitMQ
	conn, err := rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s:%d%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost),
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com RabbitMQ: %w", err)
	}

	// Criar options do publisher baseado na configuração
	publisherOptions := []func(*rabbitmq.PublisherOptions){
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(cfg.Exchange),
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	}

	// Adicionar durable se configurado
	if cfg.Durable {
		publisherOptions = append(publisherOptions, rabbitmq.WithPublisherOptionsExchangeDurable)
	}

	// Adicionar auto_delete se configurado
	if cfg.AutoDelete {
		publisherOptions = append(publisherOptions, rabbitmq.WithPublisherOptionsExchangeAutoDelete)
	}

	// Criar publisher com as opções configuradas
	publisher, err := rabbitmq.NewPublisher(conn, publisherOptions...)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar publisher RabbitMQ: %w", err)
	}

	return &messagePublisher{
		publisher: publisher,
		config:    cfg,
		logger:    logger,
	}, nil
}

// PublishMessage publica uma mensagem na fila RabbitMQ
func (p *messagePublisher) PublishMessage(ctx context.Context, message messaging.MessageDTO) error {
	// Serializar mensagem para JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	// Publicar mensagem com confirmação
	err = p.publisher.Publish(
		messageBytes,
		[]string{p.config.RoutingKey},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(p.config.Exchange),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsPersistentDelivery,
	)
	if err != nil {
		return fmt.Errorf("erro ao publicar mensagem no RabbitMQ: %w", err)
	}

	p.logger.Debug("Mensagem publicada com sucesso",
		zap.String("exchange", p.config.Exchange),
		zap.String("routingKey", p.config.RoutingKey),
		zap.String("correlationID", message.XCorrelationID),
		zap.String("recipient", message.Recipient),
		zap.String("templateType", message.TemplateType),
	)

	return nil
}

// Close fecha a conexão com RabbitMQ
func (p *messagePublisher) Close() error {
	if p.publisher != nil {
		p.publisher.Close()
		p.logger.Info("Publisher RabbitMQ fechado com sucesso")
	}
	return nil
}
