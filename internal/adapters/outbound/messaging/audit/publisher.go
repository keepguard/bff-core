package audit

import (
	"context"
	"encoding/json"
	"fmt"

	auditport "github.com/keepguard/bff-core/internal/domain/ports/audit"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

type publisher struct {
	publisher *rabbitmq.Publisher
	conn      *rabbitmq.Conn
	exchange  string
	routing   string
	logger    *zap.Logger
}

func NewPublisher(cfg *config.RabbitMQConfig, logger *zap.Logger) (auditport.EventPublisher, error) {
	if cfg == nil || !cfg.Audit.Enabled || cfg.Audit.Exchange == "" {
		return noopPublisher{}, nil
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	vhost := cfg.VHost
	if vhost == "" {
		vhost = "/"
	}
	if vhost[0] != '/' {
		vhost = "/" + vhost
	}
	conn, err := rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s:%d%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, vhost),
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, fmt.Errorf("audit rabbit conn: %w", err)
	}
	opts := []func(*rabbitmq.PublisherOptions){
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(cfg.Audit.Exchange),
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	}
	if cfg.Audit.Durable || cfg.Durable {
		opts = append(opts, rabbitmq.WithPublisherOptionsExchangeDurable)
	}
	pub, err := rabbitmq.NewPublisher(conn, opts...)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("audit rabbit publisher: %w", err)
	}
	routing := cfg.Audit.RoutingKey
	if routing == "" {
		routing = "audit.event"
	}
	return &publisher{publisher: pub, conn: conn, exchange: cfg.Audit.Exchange, routing: routing, logger: logger}, nil
}

func (p *publisher) Publish(_ context.Context, event auditport.Event) {
	go func() {
		body, err := json.Marshal(event)
		if err != nil {
			p.logger.Warn("falha ao serializar evento de auditoria", zap.Error(err), zap.String("correlationId", event.CorrelationID))
			return
		}
		err = p.publisher.Publish(
			body,
			[]string{p.routing},
			rabbitmq.WithPublishOptionsContentType("application/json"),
			rabbitmq.WithPublishOptionsExchange(p.exchange),
			rabbitmq.WithPublishOptionsPersistentDelivery,
			rabbitmq.WithPublishOptionsHeaders(rabbitmq.Table{
				"X-Correlation-ID": event.CorrelationID,
			}),
		)
		if err != nil {
			p.logger.Warn("falha ao publicar evento de auditoria",
				zap.Error(err),
				zap.String("correlationId", event.CorrelationID),
				zap.String("action", event.Action),
			)
		}
	}()
}

func (p *publisher) Close() error {
	if p.publisher != nil {
		p.publisher.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, auditport.Event) {}
func (noopPublisher) Close() error                             { return nil }
