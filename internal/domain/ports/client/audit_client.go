package client

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type AuditClient interface {
	List(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedAuditResponse, error)
	GetByID(ctx context.Context, tenantID, correlationID, eventID string) (appdto.AuditDetailResponse, error)
}
