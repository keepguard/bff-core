package audit

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
)

type ListAuditsQuery struct {
	TenantID      string
	CorrelationID string
	Filters       map[string]string
}

type GetAuditQuery struct {
	TenantID      string
	CorrelationID string
	EventID       string
}

type AuditPort interface {
	List(ctx context.Context, query ListAuditsQuery) (appdto.PaginatedAuditResponse, error)
	Get(ctx context.Context, query GetAuditQuery) (appdto.AuditDetailResponse, error)
}

type service struct {
	audits port.AuditClient
}

func NewAuditPort(audits port.AuditClient) AuditPort {
	if audits == nil {
		return nil
	}
	return &service{audits: audits}
}

func (s *service) List(ctx context.Context, query ListAuditsQuery) (appdto.PaginatedAuditResponse, error) {
	return s.audits.List(ctx, query.TenantID, query.CorrelationID, query.Filters)
}

func (s *service) Get(ctx context.Context, query GetAuditQuery) (appdto.AuditDetailResponse, error) {
	return s.audits.GetByID(ctx, query.TenantID, query.CorrelationID, query.EventID)
}
