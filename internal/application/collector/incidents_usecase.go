package collector

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
)

func (s *service) ListIncidents(ctx context.Context, query ListIncidentsQuery) (appdto.PaginatedCollectorIncidents, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return appdto.PaginatedCollectorIncidents{}, err
	}
	raw, err := s.collector.ListIncidents(ctx, companyID, query.CorrelationID, query.Filters)
	if err != nil {
		return appdto.PaginatedCollectorIncidents{}, err
	}
	return appdto.MapPaginatedCollectorIncidents(raw), nil
}

func (s *service) ListAgentIncidents(ctx context.Context, query ListAgentIncidentsQuery) ([]appdto.CollectorIncidentDTO, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return nil, err
	}
	raw, err := s.collector.ListAgentIncidents(ctx, companyID, query.AgentID, query.CorrelationID)
	if err != nil {
		return nil, err
	}
	out := make([]appdto.CollectorIncidentDTO, 0, len(raw))
	for _, item := range raw {
		out = append(out, appdto.MapCollectorIncidentRaw(item))
	}
	return out, nil
}

func (s *service) AcknowledgeIncident(ctx context.Context, cmd MutateIncidentCommand) (appdto.CollectorIncidentDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorIncidentDTO{}, err
	}
	raw, err := s.collector.AcknowledgeIncident(port.WithUserID(ctx, cmd.ActorUserID), companyID, cmd.ID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorIncidentDTO{}, err
	}
	return appdto.MapCollectorIncidentRaw(raw), nil
}

func (s *service) ResolveIncident(ctx context.Context, cmd MutateIncidentCommand) (appdto.CollectorIncidentDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorIncidentDTO{}, err
	}
	raw, err := s.collector.ResolveIncident(port.WithUserID(ctx, cmd.ActorUserID), companyID, cmd.ID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorIncidentDTO{}, err
	}
	return appdto.MapCollectorIncidentRaw(raw), nil
}

func (s *service) GetIncidentSuggestion(ctx context.Context, query GetIncidentSuggestionQuery) (appdto.CollectorIncidentSuggestionDTO, bool, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return appdto.CollectorIncidentSuggestionDTO{}, false, err
	}
	raw, ok, err := s.collector.GetIncidentSuggestion(ctx, companyID, query.ID, query.CorrelationID)
	if err != nil {
		return appdto.CollectorIncidentSuggestionDTO{}, false, err
	}
	if !ok {
		return appdto.CollectorIncidentSuggestionDTO{}, false, nil
	}
	return appdto.MapCollectorIncidentSuggestion(raw), true, nil
}

func (s *service) ApplyIncidentSuccessor(ctx context.Context, cmd ApplyIncidentSuccessorCommand) (appdto.CollectorIncidentDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorIncidentDTO{}, err
	}
	raw, err := s.collector.ApplyIncidentSuccessor(
		port.WithUserID(ctx, cmd.ActorUserID),
		companyID,
		cmd.ID,
		cmd.CorrelationID,
		appdto.CollectorApplySuccessorRaw{Confirmed: cmd.Confirmed},
	)
	if err != nil {
		return appdto.CollectorIncidentDTO{}, err
	}
	return appdto.MapCollectorIncidentRaw(raw), nil
}
