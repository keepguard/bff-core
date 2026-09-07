package guardian

import (
	"context"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.GuardianClient
	cfg   observe.Config
}

func New(inner port.GuardianClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.GuardianClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) ListIncidents(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedGuardianIncidents, error) {
	return observe.Call(d.cfg, "ListIncidents", "GET", "/incidents", correlationID, func() (appdto.PaginatedGuardianIncidents, error) {
		return d.inner.ListIncidents(ctx, tenantID, correlationID, query)
	})
}

func (d *decorated) GetIncident(ctx context.Context, tenantID, correlationID, id string) (map[string]any, error) {
	return observe.Call(d.cfg, "GetIncident", "GET", "/incidents/{id}", correlationID, func() (map[string]any, error) {
		return d.inner.GetIncident(ctx, tenantID, correlationID, id)
	})
}

func (d *decorated) ExecuteAction(ctx context.Context, tenantID, correlationID, userID, userEmail, userRole, id string, body appdto.GuardianExecuteActionRequest) (map[string]any, error) {
	return observe.Call(d.cfg, "ExecuteAction", "POST", "/incidents/{id}/actions", correlationID, func() (map[string]any, error) {
		return d.inner.ExecuteAction(ctx, tenantID, correlationID, userID, userEmail, userRole, id, body)
	})
}

func (d *decorated) ListRecipients(ctx context.Context, tenantID, correlationID string) ([]map[string]any, error) {
	return observe.Call(d.cfg, "ListRecipients", "GET", "/alert-recipients", correlationID, func() ([]map[string]any, error) {
		return d.inner.ListRecipients(ctx, tenantID, correlationID)
	})
}

func (d *decorated) UpsertRecipient(ctx context.Context, tenantID, correlationID string, body appdto.GuardianRecipientUpsertRequest) (map[string]any, error) {
	return observe.Call(d.cfg, "UpsertRecipient", "PUT", "/alert-recipients", correlationID, func() (map[string]any, error) {
		return d.inner.UpsertRecipient(ctx, tenantID, correlationID, body)
	})
}

func (d *decorated) PatchRecipient(ctx context.Context, tenantID, correlationID, id string, body appdto.GuardianRecipientUpsertRequest) (map[string]any, error) {
	return observe.Call(d.cfg, "PatchRecipient", "PATCH", "/alert-recipients/{id}", correlationID, func() (map[string]any, error) {
		return d.inner.PatchRecipient(ctx, tenantID, correlationID, id, body)
	})
}
