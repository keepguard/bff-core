package client

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type GuardianClient interface {
	ListIncidents(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedGuardianIncidents, error)
	GetIncident(ctx context.Context, tenantID, correlationID, id string) (map[string]any, error)
	ExecuteAction(ctx context.Context, tenantID, correlationID, userID, userEmail, userRole, id string, body appdto.GuardianExecuteActionRequest) (map[string]any, error)
	ListRecipients(ctx context.Context, tenantID, correlationID string) ([]map[string]any, error)
	UpsertRecipient(ctx context.Context, tenantID, correlationID string, body appdto.GuardianRecipientUpsertRequest) (map[string]any, error)
	PatchRecipient(ctx context.Context, tenantID, correlationID, id string, body appdto.GuardianRecipientUpsertRequest) (map[string]any, error)
}
