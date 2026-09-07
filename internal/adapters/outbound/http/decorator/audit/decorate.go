package audit

import (
	"context"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.AuditClient
	cfg   observe.Config
}

func New(inner port.AuditClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.AuditClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) List(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedAuditResponse, error) {
	return observe.Call(d.cfg, "List", "GET", "/audits", correlationID, func() (appdto.PaginatedAuditResponse, error) {
		return d.inner.List(ctx, tenantID, correlationID, query)
	})
}

func (d *decorated) GetByID(ctx context.Context, tenantID, correlationID, eventID string) (appdto.AuditDetailResponse, error) {
	return observe.Call(d.cfg, "GetByID", "GET", "/audits/{id}", correlationID, func() (appdto.AuditDetailResponse, error) {
		return d.inner.GetByID(ctx, tenantID, correlationID, eventID)
	})
}
