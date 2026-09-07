package guardian

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
)

type ListIncidentsQuery struct {
	TenantID      string
	CorrelationID string
	Filters       map[string]string
}

type GetIncidentQuery struct {
	TenantID      string
	CorrelationID string
	ID            string
}

type ExecuteActionCommand struct {
	TenantID      string
	CorrelationID string
	UserID        string
	UserEmail     string
	UserRole      string
	ID            string
	Body          appdto.GuardianExecuteActionRequest
}

type ListRecipientsQuery struct {
	TenantID      string
	CorrelationID string
}

type UpsertRecipientCommand struct {
	TenantID      string
	CorrelationID string
	Body          appdto.GuardianRecipientUpsertRequest
}

type PatchRecipientCommand struct {
	TenantID      string
	CorrelationID string
	ID            string
	Body          appdto.GuardianRecipientUpsertRequest
}

type GuardianPort interface {
	ListIncidents(ctx context.Context, query ListIncidentsQuery) (appdto.PaginatedGuardianIncidents, error)
	GetIncident(ctx context.Context, query GetIncidentQuery) (map[string]any, error)
	ExecuteAction(ctx context.Context, cmd ExecuteActionCommand) (map[string]any, error)
	ListRecipients(ctx context.Context, query ListRecipientsQuery) ([]map[string]any, error)
	UpsertRecipient(ctx context.Context, cmd UpsertRecipientCommand) (map[string]any, error)
	PatchRecipient(ctx context.Context, cmd PatchRecipientCommand) (map[string]any, error)
}

type service struct {
	guardian port.GuardianClient
}

func NewGuardianPort(guardian port.GuardianClient) GuardianPort {
	if guardian == nil {
		return nil
	}
	return &service{guardian: guardian}
}

func (s *service) ListIncidents(ctx context.Context, query ListIncidentsQuery) (appdto.PaginatedGuardianIncidents, error) {
	return s.guardian.ListIncidents(ctx, query.TenantID, query.CorrelationID, query.Filters)
}

func (s *service) GetIncident(ctx context.Context, query GetIncidentQuery) (map[string]any, error) {
	return s.guardian.GetIncident(ctx, query.TenantID, query.CorrelationID, query.ID)
}

func (s *service) ExecuteAction(ctx context.Context, cmd ExecuteActionCommand) (map[string]any, error) {
	return s.guardian.ExecuteAction(ctx, cmd.TenantID, cmd.CorrelationID, cmd.UserID, cmd.UserEmail, cmd.UserRole, cmd.ID, cmd.Body)
}

func (s *service) ListRecipients(ctx context.Context, query ListRecipientsQuery) ([]map[string]any, error) {
	return s.guardian.ListRecipients(ctx, query.TenantID, query.CorrelationID)
}

func (s *service) UpsertRecipient(ctx context.Context, cmd UpsertRecipientCommand) (map[string]any, error) {
	return s.guardian.UpsertRecipient(ctx, cmd.TenantID, cmd.CorrelationID, cmd.Body)
}

func (s *service) PatchRecipient(ctx context.Context, cmd PatchRecipientCommand) (map[string]any, error) {
	return s.guardian.PatchRecipient(ctx, cmd.TenantID, cmd.CorrelationID, cmd.ID, cmd.Body)
}
