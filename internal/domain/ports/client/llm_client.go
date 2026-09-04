package client

import (
	"context"
	"encoding/json"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type LlmClient interface {
	ListProviders(ctx context.Context, tenantID, correlationID string) (json.RawMessage, error)
	CreateProvider(ctx context.Context, tenantID, correlationID string, body any) (json.RawMessage, error)
	UpdateProvider(ctx context.Context, tenantID, correlationID, id string, body any) (json.RawMessage, error)
	SetProviderEnabled(ctx context.Context, tenantID, correlationID, id string, enabled bool) (json.RawMessage, error)
	Complete(ctx context.Context, tenantID, companyID, correlationID string, body any) (json.RawMessage, error)
	ListUsage(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedLlmUsageResponse, error)
	GetUsage(ctx context.Context, tenantID, correlationID, id string) (appdto.LlmUsageResponse, error)
	ListAlertRules(ctx context.Context, tenantID, correlationID string) (json.RawMessage, error)
	CreateAlertRule(ctx context.Context, tenantID, correlationID string, body any) (json.RawMessage, error)
	UpdateAlertRule(ctx context.Context, tenantID, correlationID, id string, body any) (json.RawMessage, error)
	SetAlertRuleEnabled(ctx context.Context, tenantID, correlationID, id string, enabled bool) (json.RawMessage, error)
	ListAlertFirings(ctx context.Context, tenantID, correlationID string, query map[string]string) (json.RawMessage, error)
}
