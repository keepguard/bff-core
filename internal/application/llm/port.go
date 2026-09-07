package llm

import (
	"context"
	"encoding/json"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/application/scope"
)

// Bodies de provider/alert/complete permanecem opacos (json.RawMessage / any):
// o contrato atual do gateway LLM não é tipado no BFF.

type TenantQuery struct {
	TenantID      string
	CorrelationID string
}

type ProviderIDCommand struct {
	TenantQuery
	ID      string
	Enabled bool
	Body    any
}

type CompleteCommand struct {
	TenantQuery
	CompanyID string
	Body      any
}

type ListUsageQuery struct {
	TenantQuery
	Filters map[string]string
}

type GetUsageQuery struct {
	TenantQuery
	ID string
}

type AlertRuleCommand struct {
	TenantQuery
	ID      string
	Enabled bool
	Body    any
}

type ListAlertFiringsQuery struct {
	TenantQuery
	Filters map[string]string
}

type LlmPort interface {
	ListProviders(ctx context.Context, query TenantQuery) (json.RawMessage, error)
	CreateProvider(ctx context.Context, cmd ProviderIDCommand) (json.RawMessage, error)
	UpdateProvider(ctx context.Context, cmd ProviderIDCommand) (json.RawMessage, error)
	SetProviderEnabled(ctx context.Context, cmd ProviderIDCommand) (json.RawMessage, error)
	Complete(ctx context.Context, cmd CompleteCommand) (json.RawMessage, error)
	ListUsage(ctx context.Context, query ListUsageQuery) (appdto.PaginatedLlmUsageResponse, error)
	GetUsage(ctx context.Context, query GetUsageQuery) (appdto.LlmUsageResponse, error)
	ListAlertRules(ctx context.Context, query TenantQuery) (json.RawMessage, error)
	CreateAlertRule(ctx context.Context, cmd AlertRuleCommand) (json.RawMessage, error)
	UpdateAlertRule(ctx context.Context, cmd AlertRuleCommand) (json.RawMessage, error)
	SetAlertRuleEnabled(ctx context.Context, cmd AlertRuleCommand) (json.RawMessage, error)
	ListAlertFirings(ctx context.Context, query ListAlertFiringsQuery) (json.RawMessage, error)
}

type service struct {
	client port.LlmClient
}

func NewLlmPort(client port.LlmClient) LlmPort {
	if client == nil {
		return nil
	}
	return &service{client: client}
}

func (s *service) require() error {
	if s == nil || s.client == nil {
		return scope.Unavailable("Consulta LLM indisponível")
	}
	return nil
}

func (s *service) ListProviders(ctx context.Context, query TenantQuery) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.ListProviders(ctx, query.TenantID, query.CorrelationID)
}

func (s *service) CreateProvider(ctx context.Context, cmd ProviderIDCommand) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.CreateProvider(ctx, cmd.TenantID, cmd.CorrelationID, cmd.Body)
}

func (s *service) UpdateProvider(ctx context.Context, cmd ProviderIDCommand) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.UpdateProvider(ctx, cmd.TenantID, cmd.CorrelationID, cmd.ID, cmd.Body)
}

func (s *service) SetProviderEnabled(ctx context.Context, cmd ProviderIDCommand) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.SetProviderEnabled(ctx, cmd.TenantID, cmd.CorrelationID, cmd.ID, cmd.Enabled)
}

func (s *service) Complete(ctx context.Context, cmd CompleteCommand) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.Complete(ctx, cmd.TenantID, cmd.CompanyID, cmd.CorrelationID, cmd.Body)
}

func (s *service) ListUsage(ctx context.Context, query ListUsageQuery) (appdto.PaginatedLlmUsageResponse, error) {
	if err := s.require(); err != nil {
		return appdto.PaginatedLlmUsageResponse{}, err
	}
	return s.client.ListUsage(ctx, query.TenantID, query.CorrelationID, query.Filters)
}

func (s *service) GetUsage(ctx context.Context, query GetUsageQuery) (appdto.LlmUsageResponse, error) {
	if err := s.require(); err != nil {
		return appdto.LlmUsageResponse{}, err
	}
	return s.client.GetUsage(ctx, query.TenantID, query.CorrelationID, query.ID)
}

func (s *service) ListAlertRules(ctx context.Context, query TenantQuery) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.ListAlertRules(ctx, query.TenantID, query.CorrelationID)
}

func (s *service) CreateAlertRule(ctx context.Context, cmd AlertRuleCommand) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.CreateAlertRule(ctx, cmd.TenantID, cmd.CorrelationID, cmd.Body)
}

func (s *service) UpdateAlertRule(ctx context.Context, cmd AlertRuleCommand) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.UpdateAlertRule(ctx, cmd.TenantID, cmd.CorrelationID, cmd.ID, cmd.Body)
}

func (s *service) SetAlertRuleEnabled(ctx context.Context, cmd AlertRuleCommand) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.SetAlertRuleEnabled(ctx, cmd.TenantID, cmd.CorrelationID, cmd.ID, cmd.Enabled)
}

func (s *service) ListAlertFirings(ctx context.Context, query ListAlertFiringsQuery) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.ListAlertFirings(ctx, query.TenantID, query.CorrelationID, query.Filters)
}
