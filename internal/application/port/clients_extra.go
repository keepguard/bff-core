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

type ServiceTokenClient interface {
	GetToken(ctx context.Context, companyID string) (string, error)
}
