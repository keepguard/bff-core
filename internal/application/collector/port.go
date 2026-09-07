package collector

import (
	"context"
	"strings"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/application/scope"
	"go.uber.org/zap"
)

type CollectorPort interface {
	ListAgents(ctx context.Context, query ListAgentsQuery) (appdto.PaginatedCollectorAgents, error)
	GetAgent(ctx context.Context, query GetAgentQuery) (appdto.CollectorAgentDetailDTO, error)
	CreateAgent(ctx context.Context, cmd CreateAgentCommand) (appdto.CollectorAgentDetailDTO, error)
	UpdateAgent(ctx context.Context, cmd UpdateAgentCommand) (appdto.CollectorAgentDetailDTO, error)
	EnableAgent(ctx context.Context, cmd AgentIDCommand) (appdto.CollectorAgentDetailDTO, error)
	DisableAgent(ctx context.Context, cmd AgentIDCommand) (appdto.CollectorAgentDetailDTO, error)
	DeleteAgent(ctx context.Context, cmd AgentIDCommand) error
	TestAgent(ctx context.Context, cmd AgentIDCommand) (appdto.CollectorAgentTestResultDTO, error)
	RunAgent(ctx context.Context, cmd AgentIDCommand) (appdto.CollectorAgentRunResultDTO, error)
	BulkAgents(ctx context.Context, cmd BulkAgentsCommand) (appdto.CollectorBulkResultDTO, int, error)
	GetBulk(ctx context.Context, query GetBulkQuery) (appdto.CollectorBulkProgressDTO, error)
	GetActiveBulk(ctx context.Context, query GetActiveBulkQuery) (appdto.CollectorBulkProgressDTO, error)
	ListExecutions(ctx context.Context, query ListExecutionsQuery) ([]appdto.CollectorExecutionDTO, error)
	GetExecutionPayloads(ctx context.Context, query GetExecutionPayloadsQuery) ([]appdto.ExecutionPayloadItemDTO, error)
	ListDataSources(ctx context.Context, query ListDataSourcesQuery) ([]appdto.CollectorDataSourceDTO, error)
	GetDataSource(ctx context.Context, query GetDataSourceQuery) (appdto.CollectorDataSourceDTO, error)
	CreateDataSource(ctx context.Context, cmd CreateDataSourceCommand) (appdto.CollectorDataSourceDTO, error)
	UpdateDataSource(ctx context.Context, cmd UpdateDataSourceCommand) (appdto.CollectorDataSourceDTO, error)
	EnableDataSource(ctx context.Context, cmd DataSourceIDCommand) (appdto.CollectorDataSourceDTO, error)
	DisableDataSource(ctx context.Context, cmd DataSourceIDCommand) (appdto.CollectorDataSourceDTO, error)
	DeleteDataSource(ctx context.Context, cmd DataSourceIDCommand) error
	PropagateDataSource(ctx context.Context, cmd PropagateDataSourceCommand) (appdto.PropagateDataSourceDTO, error)
	ListIncidents(ctx context.Context, query ListIncidentsQuery) (appdto.PaginatedCollectorIncidents, error)
	ListAgentIncidents(ctx context.Context, query ListAgentIncidentsQuery) ([]appdto.CollectorIncidentDTO, error)
	AcknowledgeIncident(ctx context.Context, cmd MutateIncidentCommand) (appdto.CollectorIncidentDTO, error)
	ResolveIncident(ctx context.Context, cmd MutateIncidentCommand) (appdto.CollectorIncidentDTO, error)
	GetIncidentSuggestion(ctx context.Context, query GetIncidentSuggestionQuery) (appdto.CollectorIncidentSuggestionDTO, bool, error)
	ApplyIncidentSuccessor(ctx context.Context, cmd ApplyIncidentSuccessorCommand) (appdto.CollectorIncidentDTO, error)
}

type service struct {
	collector port.CollectorClient
	companies port.CompanyClient
	knowledge port.KnowledgeClient
	tokens    port.ServiceTokenClient
	logger    *zap.Logger
}

func NewCollectorPort(
	collector port.CollectorClient,
	companies port.CompanyClient,
	knowledge port.KnowledgeClient,
	tokens port.ServiceTokenClient,
	logger *zap.Logger,
) CollectorPort {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &service{
		collector: collector,
		companies: companies,
		knowledge: knowledge,
		tokens:    tokens,
		logger:    logger,
	}
}

func (s *service) requireCollector(ctx context.Context, companyFromCtx, tenantID, correlationID string) (string, error) {
	if s.collector == nil || s.companies == nil {
		return "", scope.Unavailable("Gestão de agents indisponível")
	}
	return scope.ResolveCompany(ctx, s.companies, companyFromCtx, tenantID, correlationID)
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
