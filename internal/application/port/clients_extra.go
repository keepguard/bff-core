package port

import (
	"context"
	"encoding/json"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type AuditClient interface {
	List(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedAuditResponse, error)
	GetByID(ctx context.Context, tenantID, correlationID, eventID string) (appdto.AuditDetailResponse, error)
}

type GuardianClient interface {
	ListIncidents(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedGuardianIncidents, error)
	GetIncident(ctx context.Context, tenantID, correlationID, id string) (map[string]any, error)
	ExecuteAction(ctx context.Context, tenantID, correlationID, userID, userEmail, userRole, id string, body appdto.GuardianExecuteActionRequest) (map[string]any, error)
	ListRecipients(ctx context.Context, tenantID, correlationID string) ([]map[string]any, error)
	UpsertRecipient(ctx context.Context, tenantID, correlationID string, body appdto.GuardianRecipientUpsertRequest) (map[string]any, error)
	PatchRecipient(ctx context.Context, tenantID, correlationID, id string, body appdto.GuardianRecipientUpsertRequest) (map[string]any, error)
}

type KnowledgeClient interface {
	Ask(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.KnowledgeAskRequest) (appdto.KnowledgeAskResponse, error)
	GetSnapshot(ctx context.Context, companyID, bearerToken, correlationID, snapshotID string) (appdto.KnowledgeSnapshotDTO, error)
	GetDocumentPreview(ctx context.Context, companyID, bearerToken, correlationID, documentID string) (appdto.KnowledgeDocumentPreviewDTO, error)
	GetCollectionResults(ctx context.Context, companyID, bearerToken, correlationID, agentID, collectedAt string, windowSeconds int) (appdto.KnowledgeCollectionResultsDTO, error)
}

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

type BillingScope struct {
	CompanyID     string
	UserID        string
	CorrelationID string
	Admin         bool
}

type BillingClient interface {
	GetEntitlement(ctx context.Context, scope BillingScope) (json.RawMessage, error)
	ListPlans(ctx context.Context, scope BillingScope) (json.RawMessage, error)
	SavePlan(ctx context.Context, scope BillingScope, body any) (json.RawMessage, error)
	PatchPlan(ctx context.Context, scope BillingScope, code string, body any) (json.RawMessage, error)
	GetGatewayAccount(ctx context.Context, scope BillingScope) (json.RawMessage, error)
	PutGatewayAccount(ctx context.Context, scope BillingScope, body any) (json.RawMessage, error)
	GetSubscription(ctx context.Context, scope BillingScope) (json.RawMessage, error)
	CreateSubscription(ctx context.Context, scope BillingScope, body any) (json.RawMessage, int, error)
	CancelSubscription(ctx context.Context, scope BillingScope, id string) (json.RawMessage, error)
	ListInvoices(ctx context.Context, scope BillingScope) (json.RawMessage, error)
	GetInvoice(ctx context.Context, scope BillingScope, id string) (json.RawMessage, error)
	ForwardAsaasWebhook(ctx context.Context, accessToken string, body []byte) (json.RawMessage, int, error)
}

type ServiceTokenClient interface {
	GetToken(ctx context.Context, companyID string) (string, error)
}
