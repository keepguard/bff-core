package client

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type OAuthClientClient interface {
	Search(ctx context.Context, companyID, bearerToken, correlationID string, query map[string]string) (appdto.PaginatedOAuthClients, error)
	GetByID(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Create(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.OAuthClientCreateRequest) (appdto.OAuthClientDTO, error)
	Update(ctx context.Context, companyID, bearerToken, correlationID, id string, body appdto.OAuthClientUpdateRequest) (appdto.OAuthClientDTO, error)
	ListServiceRoles(ctx context.Context, companyID, bearerToken, correlationID string) ([]appdto.OAuthServiceRoleDTO, error)
	Block(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Unblock(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Delete(ctx context.Context, companyID, bearerToken, correlationID, id string) error
}

type CollectorClient interface {
	ListAgents(ctx context.Context, companyID, correlationID string) ([]appdto.CollectorAgentRaw, error)
	SearchAgents(ctx context.Context, companyID, correlationID string, query map[string]string) (appdto.PaginatedCollectorAgentsRaw, error)
	GetAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error)
	CreateAgent(ctx context.Context, companyID, correlationID string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error)
	UpdateAgent(ctx context.Context, companyID, agentID, correlationID string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error)
	EnableAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error)
	DisableAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error)
	TestAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentTestResultDTO, error)
	RunAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRunResultDTO, error)
	BulkAgents(ctx context.Context, companyID, correlationID string, body appdto.CollectorBulkWriteRaw) (appdto.CollectorBulkResultDTO, int, error)
	GetBulkOperation(ctx context.Context, companyID, bulkID, correlationID string) (appdto.CollectorBulkProgressDTO, error)
	GetActiveBulkOperation(ctx context.Context, companyID, correlationID string) (appdto.CollectorBulkProgressDTO, error)
	ListAgentExecutions(ctx context.Context, companyID, agentID, correlationID string, limit int) ([]appdto.CollectorExecutionRaw, error)
	GetExecution(ctx context.Context, companyID, executionID, correlationID string) (appdto.CollectorExecutionRaw, error)
	DeleteAgent(ctx context.Context, companyID, agentID, correlationID string) error
	ListDataSources(ctx context.Context, companyID, correlationID string, query map[string]string) ([]appdto.CollectorDataSourceRaw, error)
	GetDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error)
	CreateDataSource(ctx context.Context, companyID, correlationID string, body appdto.CollectorDataSourceWriteRaw) (appdto.CollectorDataSourceRaw, error)
	UpdateDataSource(ctx context.Context, companyID, sourceID, correlationID string, body appdto.CollectorDataSourceWriteRaw) (appdto.CollectorDataSourceRaw, error)
	EnableDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error)
	DisableDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error)
	DeleteDataSource(ctx context.Context, companyID, sourceID, correlationID string) error
	PropagateDataSource(ctx context.Context, companyID, sourceID, correlationID string, body appdto.PropagateDataSourceWriteRaw) (appdto.PropagateDataSourceRaw, error)
	ListIncidents(ctx context.Context, companyID, correlationID string, query map[string]string) (appdto.PaginatedCollectorIncidentsRaw, error)
	ListAgentIncidents(ctx context.Context, companyID, agentID, correlationID string) ([]appdto.CollectorIncidentRaw, error)
	AcknowledgeIncident(ctx context.Context, companyID, incidentID, correlationID string) (appdto.CollectorIncidentRaw, error)
	ResolveIncident(ctx context.Context, companyID, incidentID, correlationID string) (appdto.CollectorIncidentRaw, error)
	GetIncidentSuggestion(ctx context.Context, companyID, incidentID, correlationID string) (appdto.CollectorIncidentSuggestionRaw, bool, error)
	ApplyIncidentSuccessor(ctx context.Context, companyID, incidentID, correlationID string, body appdto.CollectorApplySuccessorRaw) (appdto.CollectorIncidentRaw, error)
}
