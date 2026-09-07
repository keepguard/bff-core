package collector

import (
	"context"
	"net/http"
	"strings"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
)

func (s *service) ListAgents(ctx context.Context, query ListAgentsQuery) (appdto.PaginatedCollectorAgents, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return appdto.PaginatedCollectorAgents{}, err
	}
	raw, err := s.collector.SearchAgents(ctx, companyID, query.CorrelationID, query.Filters)
	if err != nil {
		return appdto.PaginatedCollectorAgents{}, err
	}
	return appdto.MapPaginatedCollectorAgents(raw), nil
}

func (s *service) GetAgent(ctx context.Context, query GetAgentQuery) (appdto.CollectorAgentDetailDTO, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return appdto.CollectorAgentDetailDTO{}, err
	}
	raw, err := s.collector.GetAgent(ctx, companyID, query.ID, query.CorrelationID)
	if err != nil {
		return appdto.CollectorAgentDetailDTO{}, err
	}
	return appdto.MapCollectorAgentRaw(raw), nil
}

func (s *service) CreateAgent(ctx context.Context, cmd CreateAgentCommand) (appdto.CollectorAgentDetailDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorAgentDetailDTO{}, err
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return appdto.CollectorAgentDetailDTO{}, pkg.NewAppError("MISSING_NAME", "name é obrigatório", http.StatusBadRequest)
	}
	if strings.TrimSpace(cmd.CollectorType) == "" {
		return appdto.CollectorAgentDetailDTO{}, pkg.NewAppError("MISSING_COLLECTOR_TYPE", "collectorType é obrigatório", http.StatusBadRequest)
	}
	enabled := false
	if cmd.Enabled != nil {
		enabled = *cmd.Enabled
	}
	contextLabel := strings.TrimSpace(cmd.Context)
	if contextLabel == "" {
		contextLabel = "geral"
	}
	schedule := appdto.MapCollectorScheduleDTO(cmd.Schedule)
	raw, err := s.collector.CreateAgent(ctx, companyID, cmd.CorrelationID, appdto.CollectorAgentWriteRaw{
		Name:            cmd.Name,
		Description:     optionalString(cmd.Description),
		Context:         &contextLabel,
		CollectorType:   cmd.CollectorType,
		CollectorConfig: cmd.CollectorConfig,
		Prompt:          optionalString(cmd.Prompt),
		Schedule:        &schedule,
		Enabled:         &enabled,
		DataSourceID:    optionalString(cmd.DataSourceID),
	})
	if err != nil {
		return appdto.CollectorAgentDetailDTO{}, err
	}
	return appdto.MapCollectorAgentRaw(raw), nil
}

func (s *service) UpdateAgent(ctx context.Context, cmd UpdateAgentCommand) (appdto.CollectorAgentDetailDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorAgentDetailDTO{}, err
	}
	write := appdto.CollectorAgentWriteRaw{
		CollectorConfig: cmd.CollectorConfig,
		Prompt:          cmd.Prompt,
		Description:     cmd.Description,
		Context:         cmd.Context,
	}
	if cmd.Name != nil {
		write.Name = *cmd.Name
	}
	if cmd.Schedule != nil {
		schedule := appdto.MapCollectorScheduleDTO(*cmd.Schedule)
		write.Schedule = &schedule
	}
	if cmd.DataSourceID != nil {
		write.DataSourceID = cmd.DataSourceID
	}
	raw, err := s.collector.UpdateAgent(ctx, companyID, cmd.ID, cmd.CorrelationID, write)
	if err != nil {
		return appdto.CollectorAgentDetailDTO{}, err
	}
	return appdto.MapCollectorAgentRaw(raw), nil
}

func (s *service) EnableAgent(ctx context.Context, cmd AgentIDCommand) (appdto.CollectorAgentDetailDTO, error) {
	return s.toggleAgent(ctx, cmd, true)
}

func (s *service) DisableAgent(ctx context.Context, cmd AgentIDCommand) (appdto.CollectorAgentDetailDTO, error) {
	return s.toggleAgent(ctx, cmd, false)
}

func (s *service) toggleAgent(ctx context.Context, cmd AgentIDCommand, enable bool) (appdto.CollectorAgentDetailDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorAgentDetailDTO{}, err
	}
	var raw appdto.CollectorAgentRaw
	if enable {
		raw, err = s.collector.EnableAgent(ctx, companyID, cmd.ID, cmd.CorrelationID)
	} else {
		raw, err = s.collector.DisableAgent(ctx, companyID, cmd.ID, cmd.CorrelationID)
	}
	if err != nil {
		return appdto.CollectorAgentDetailDTO{}, err
	}
	return appdto.MapCollectorAgentRaw(raw), nil
}

func (s *service) DeleteAgent(ctx context.Context, cmd AgentIDCommand) error {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return err
	}
	return s.collector.DeleteAgent(ctx, companyID, cmd.ID, cmd.CorrelationID)
}

func (s *service) TestAgent(ctx context.Context, cmd AgentIDCommand) (appdto.CollectorAgentTestResultDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorAgentTestResultDTO{}, err
	}
	return s.collector.TestAgent(ctx, companyID, cmd.ID, cmd.CorrelationID)
}

func (s *service) RunAgent(ctx context.Context, cmd AgentIDCommand) (appdto.CollectorAgentRunResultDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorAgentRunResultDTO{}, err
	}
	return s.collector.RunAgent(ctx, companyID, cmd.ID, cmd.CorrelationID)
}

func (s *service) BulkAgents(ctx context.Context, cmd BulkAgentsCommand) (appdto.CollectorBulkResultDTO, int, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorBulkResultDTO{}, 0, err
	}
	result, status, err := s.collector.BulkAgents(ctx, companyID, cmd.CorrelationID, appdto.CollectorBulkWriteRaw{
		Action: cmd.Action,
		IDs:    cmd.IDs,
	})
	if err != nil {
		return appdto.CollectorBulkResultDTO{}, 0, err
	}
	if status == 0 {
		status = http.StatusOK
	}
	return result, status, nil
}

func (s *service) GetBulk(ctx context.Context, query GetBulkQuery) (appdto.CollectorBulkProgressDTO, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return appdto.CollectorBulkProgressDTO{}, err
	}
	return s.collector.GetBulkOperation(ctx, companyID, query.ID, query.CorrelationID)
}

func (s *service) GetActiveBulk(ctx context.Context, query GetActiveBulkQuery) (appdto.CollectorBulkProgressDTO, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return appdto.CollectorBulkProgressDTO{}, err
	}
	return s.collector.GetActiveBulkOperation(ctx, companyID, query.CorrelationID)
}

func (s *service) ListExecutions(ctx context.Context, query ListExecutionsQuery) ([]appdto.CollectorExecutionDTO, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return nil, err
	}
	executions, err := s.collector.ListAgentExecutions(ctx, companyID, query.AgentID, query.CorrelationID, query.Limit)
	if err != nil {
		return nil, err
	}
	return appdto.MapCollectorExecutions(executions), nil
}
