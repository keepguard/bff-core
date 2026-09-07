package llmgateway

import (
	"context"
	"encoding/json"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.LlmClient
	cfg   observe.Config
}

func New(inner port.LlmClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.LlmClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) ListProviders(ctx context.Context, tenantID, correlationID string) (json.RawMessage, error) {
	return observe.Call(d.cfg, "ListProviders", "GET", "/llm/providers", correlationID, func() (json.RawMessage, error) {
		return d.inner.ListProviders(ctx, tenantID, correlationID)
	})
}

func (d *decorated) CreateProvider(ctx context.Context, tenantID, correlationID string, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "CreateProvider", "POST", "/llm/providers", correlationID, func() (json.RawMessage, error) {
		return d.inner.CreateProvider(ctx, tenantID, correlationID, body)
	})
}

func (d *decorated) UpdateProvider(ctx context.Context, tenantID, correlationID, id string, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "UpdateProvider", "PUT", "/llm/providers/{id}", correlationID, func() (json.RawMessage, error) {
		return d.inner.UpdateProvider(ctx, tenantID, correlationID, id, body)
	})
}

func (d *decorated) SetProviderEnabled(ctx context.Context, tenantID, correlationID, id string, enabled bool) (json.RawMessage, error) {
	return observe.Call(d.cfg, "SetProviderEnabled", "PATCH", "/llm/providers/{id}/enabled", correlationID, func() (json.RawMessage, error) {
		return d.inner.SetProviderEnabled(ctx, tenantID, correlationID, id, enabled)
	})
}

func (d *decorated) Complete(ctx context.Context, tenantID, companyID, correlationID string, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "Complete", "POST", "/llm/complete", correlationID, func() (json.RawMessage, error) {
		return d.inner.Complete(ctx, tenantID, companyID, correlationID, body)
	})
}

func (d *decorated) ListUsage(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedLlmUsageResponse, error) {
	return observe.Call(d.cfg, "ListUsage", "GET", "/llm/usage", correlationID, func() (appdto.PaginatedLlmUsageResponse, error) {
		return d.inner.ListUsage(ctx, tenantID, correlationID, query)
	})
}

func (d *decorated) GetUsage(ctx context.Context, tenantID, correlationID, id string) (appdto.LlmUsageResponse, error) {
	return observe.Call(d.cfg, "GetUsage", "GET", "/llm/usage/{id}", correlationID, func() (appdto.LlmUsageResponse, error) {
		return d.inner.GetUsage(ctx, tenantID, correlationID, id)
	})
}

func (d *decorated) ListAlertRules(ctx context.Context, tenantID, correlationID string) (json.RawMessage, error) {
	return observe.Call(d.cfg, "ListAlertRules", "GET", "/llm/alert-rules", correlationID, func() (json.RawMessage, error) {
		return d.inner.ListAlertRules(ctx, tenantID, correlationID)
	})
}

func (d *decorated) CreateAlertRule(ctx context.Context, tenantID, correlationID string, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "CreateAlertRule", "POST", "/llm/alert-rules", correlationID, func() (json.RawMessage, error) {
		return d.inner.CreateAlertRule(ctx, tenantID, correlationID, body)
	})
}

func (d *decorated) UpdateAlertRule(ctx context.Context, tenantID, correlationID, id string, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "UpdateAlertRule", "PUT", "/llm/alert-rules/{id}", correlationID, func() (json.RawMessage, error) {
		return d.inner.UpdateAlertRule(ctx, tenantID, correlationID, id, body)
	})
}

func (d *decorated) SetAlertRuleEnabled(ctx context.Context, tenantID, correlationID, id string, enabled bool) (json.RawMessage, error) {
	return observe.Call(d.cfg, "SetAlertRuleEnabled", "PATCH", "/llm/alert-rules/{id}/enabled", correlationID, func() (json.RawMessage, error) {
		return d.inner.SetAlertRuleEnabled(ctx, tenantID, correlationID, id, enabled)
	})
}

func (d *decorated) ListAlertFirings(ctx context.Context, tenantID, correlationID string, query map[string]string) (json.RawMessage, error) {
	return observe.Call(d.cfg, "ListAlertFirings", "GET", "/llm/alert-firings", correlationID, func() (json.RawMessage, error) {
		return d.inner.ListAlertFirings(ctx, tenantID, correlationID, query)
	})
}
