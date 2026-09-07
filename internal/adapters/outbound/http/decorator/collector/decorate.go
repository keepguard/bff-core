package collector

import (
	"context"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.CollectorClient
	cfg   observe.Config
}

func New(inner port.CollectorClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.CollectorClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) ListAgents(ctx context.Context, companyID, correlationID string) ([]appdto.CollectorAgentRaw, error) {
	return observe.Call(d.cfg, "ListAgents", "GET", "/agents", correlationID, func() ([]appdto.CollectorAgentRaw, error) {
		return d.inner.ListAgents(ctx, companyID, correlationID)
	})
}

func (d *decorated) SearchAgents(ctx context.Context, companyID, correlationID string, query map[string]string) (appdto.PaginatedCollectorAgentsRaw, error) {
	return observe.Call(d.cfg, "SearchAgents", "GET", "/agents/search", correlationID, func() (appdto.PaginatedCollectorAgentsRaw, error) {
		return d.inner.SearchAgents(ctx, companyID, correlationID, query)
	})
}

func (d *decorated) GetAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error) {
	return observe.Call(d.cfg, "GetAgent", "GET", "/agents/{id}", correlationID, func() (appdto.CollectorAgentRaw, error) {
		return d.inner.GetAgent(ctx, companyID, agentID, correlationID)
	})
}

func (d *decorated) CreateAgent(ctx context.Context, companyID, correlationID string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error) {
	return observe.Call(d.cfg, "CreateAgent", "POST", "/agents", correlationID, func() (appdto.CollectorAgentRaw, error) {
		return d.inner.CreateAgent(ctx, companyID, correlationID, body)
	})
}

func (d *decorated) UpdateAgent(ctx context.Context, companyID, agentID, correlationID string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error) {
	return observe.Call(d.cfg, "UpdateAgent", "PUT", "/agents/{id}", correlationID, func() (appdto.CollectorAgentRaw, error) {
		return d.inner.UpdateAgent(ctx, companyID, agentID, correlationID, body)
	})
}

func (d *decorated) EnableAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error) {
	return observe.Call(d.cfg, "EnableAgent", "POST", "/agents/{id}/enable", correlationID, func() (appdto.CollectorAgentRaw, error) {
		return d.inner.EnableAgent(ctx, companyID, agentID, correlationID)
	})
}

func (d *decorated) DisableAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error) {
	return observe.Call(d.cfg, "DisableAgent", "POST", "/agents/{id}/disable", correlationID, func() (appdto.CollectorAgentRaw, error) {
		return d.inner.DisableAgent(ctx, companyID, agentID, correlationID)
	})
}

func (d *decorated) TestAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentTestResultDTO, error) {
	return observe.Call(d.cfg, "TestAgent", "POST", "/agents/{id}/test", correlationID, func() (appdto.CollectorAgentTestResultDTO, error) {
		return d.inner.TestAgent(ctx, companyID, agentID, correlationID)
	})
}

func (d *decorated) RunAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRunResultDTO, error) {
	return observe.Call(d.cfg, "RunAgent", "POST", "/agents/{id}/run", correlationID, func() (appdto.CollectorAgentRunResultDTO, error) {
		return d.inner.RunAgent(ctx, companyID, agentID, correlationID)
	})
}

type bulkResult struct {
	Body   appdto.CollectorBulkResultDTO
	Status int
}

func (d *decorated) BulkAgents(ctx context.Context, companyID, correlationID string, body appdto.CollectorBulkWriteRaw) (appdto.CollectorBulkResultDTO, int, error) {
	out, err := observe.Call(d.cfg, "BulkAgents", "POST", "/agents/bulk", correlationID, func() (bulkResult, error) {
		result, status, callErr := d.inner.BulkAgents(ctx, companyID, correlationID, body)
		return bulkResult{Body: result, Status: status}, callErr
	})
	return out.Body, out.Status, err
}

func (d *decorated) GetBulkOperation(ctx context.Context, companyID, bulkID, correlationID string) (appdto.CollectorBulkProgressDTO, error) {
	return observe.Call(d.cfg, "GetBulkOperation", "GET", "/agents/bulk/{id}", correlationID, func() (appdto.CollectorBulkProgressDTO, error) {
		return d.inner.GetBulkOperation(ctx, companyID, bulkID, correlationID)
	})
}

func (d *decorated) GetActiveBulkOperation(ctx context.Context, companyID, correlationID string) (appdto.CollectorBulkProgressDTO, error) {
	return observe.Call(d.cfg, "GetActiveBulkOperation", "GET", "/agents/bulk/active", correlationID, func() (appdto.CollectorBulkProgressDTO, error) {
		return d.inner.GetActiveBulkOperation(ctx, companyID, correlationID)
	})
}

func (d *decorated) ListAgentExecutions(ctx context.Context, companyID, agentID, correlationID string, limit int) ([]appdto.CollectorExecutionRaw, error) {
	return observe.Call(d.cfg, "ListAgentExecutions", "GET", "/agents/{id}/executions", correlationID, func() ([]appdto.CollectorExecutionRaw, error) {
		return d.inner.ListAgentExecutions(ctx, companyID, agentID, correlationID, limit)
	})
}

func (d *decorated) GetExecution(ctx context.Context, companyID, executionID, correlationID string) (appdto.CollectorExecutionRaw, error) {
	return observe.Call(d.cfg, "GetExecution", "GET", "/executions/{id}", correlationID, func() (appdto.CollectorExecutionRaw, error) {
		return d.inner.GetExecution(ctx, companyID, executionID, correlationID)
	})
}

func (d *decorated) DeleteAgent(ctx context.Context, companyID, agentID, correlationID string) error {
	return observe.CallErr(d.cfg, "DeleteAgent", "DELETE", "/agents/{id}", correlationID, func() error {
		return d.inner.DeleteAgent(ctx, companyID, agentID, correlationID)
	})
}

func (d *decorated) ListDataSources(ctx context.Context, companyID, correlationID string, query map[string]string) ([]appdto.CollectorDataSourceRaw, error) {
	return observe.Call(d.cfg, "ListDataSources", "GET", "/data-sources", correlationID, func() ([]appdto.CollectorDataSourceRaw, error) {
		return d.inner.ListDataSources(ctx, companyID, correlationID, query)
	})
}

func (d *decorated) GetDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error) {
	return observe.Call(d.cfg, "GetDataSource", "GET", "/data-sources/{id}", correlationID, func() (appdto.CollectorDataSourceRaw, error) {
		return d.inner.GetDataSource(ctx, companyID, sourceID, correlationID)
	})
}

func (d *decorated) CreateDataSource(ctx context.Context, companyID, correlationID string, body appdto.CollectorDataSourceWriteRaw) (appdto.CollectorDataSourceRaw, error) {
	return observe.Call(d.cfg, "CreateDataSource", "POST", "/data-sources", correlationID, func() (appdto.CollectorDataSourceRaw, error) {
		return d.inner.CreateDataSource(ctx, companyID, correlationID, body)
	})
}

func (d *decorated) UpdateDataSource(ctx context.Context, companyID, sourceID, correlationID string, body appdto.CollectorDataSourceWriteRaw) (appdto.CollectorDataSourceRaw, error) {
	return observe.Call(d.cfg, "UpdateDataSource", "PUT", "/data-sources/{id}", correlationID, func() (appdto.CollectorDataSourceRaw, error) {
		return d.inner.UpdateDataSource(ctx, companyID, sourceID, correlationID, body)
	})
}

func (d *decorated) EnableDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error) {
	return observe.Call(d.cfg, "EnableDataSource", "POST", "/data-sources/{id}/enable", correlationID, func() (appdto.CollectorDataSourceRaw, error) {
		return d.inner.EnableDataSource(ctx, companyID, sourceID, correlationID)
	})
}

func (d *decorated) DisableDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error) {
	return observe.Call(d.cfg, "DisableDataSource", "POST", "/data-sources/{id}/disable", correlationID, func() (appdto.CollectorDataSourceRaw, error) {
		return d.inner.DisableDataSource(ctx, companyID, sourceID, correlationID)
	})
}

func (d *decorated) DeleteDataSource(ctx context.Context, companyID, sourceID, correlationID string) error {
	return observe.CallErr(d.cfg, "DeleteDataSource", "DELETE", "/data-sources/{id}", correlationID, func() error {
		return d.inner.DeleteDataSource(ctx, companyID, sourceID, correlationID)
	})
}

func (d *decorated) PropagateDataSource(ctx context.Context, companyID, sourceID, correlationID string, body appdto.PropagateDataSourceWriteRaw) (appdto.PropagateDataSourceRaw, error) {
	return observe.Call(d.cfg, "PropagateDataSource", "POST", "/data-sources/{id}/propagate", correlationID, func() (appdto.PropagateDataSourceRaw, error) {
		return d.inner.PropagateDataSource(ctx, companyID, sourceID, correlationID, body)
	})
}

func (d *decorated) ListIncidents(ctx context.Context, companyID, correlationID string, query map[string]string) (appdto.PaginatedCollectorIncidentsRaw, error) {
	return observe.Call(d.cfg, "ListIncidents", "GET", "/incidents", correlationID, func() (appdto.PaginatedCollectorIncidentsRaw, error) {
		return d.inner.ListIncidents(ctx, companyID, correlationID, query)
	})
}

func (d *decorated) ListAgentIncidents(ctx context.Context, companyID, agentID, correlationID string) ([]appdto.CollectorIncidentRaw, error) {
	return observe.Call(d.cfg, "ListAgentIncidents", "GET", "/agents/{id}/incidents", correlationID, func() ([]appdto.CollectorIncidentRaw, error) {
		return d.inner.ListAgentIncidents(ctx, companyID, agentID, correlationID)
	})
}

func (d *decorated) AcknowledgeIncident(ctx context.Context, companyID, incidentID, correlationID string) (appdto.CollectorIncidentRaw, error) {
	return observe.Call(d.cfg, "AcknowledgeIncident", "POST", "/incidents/{id}/ack", correlationID, func() (appdto.CollectorIncidentRaw, error) {
		return d.inner.AcknowledgeIncident(ctx, companyID, incidentID, correlationID)
	})
}

func (d *decorated) ResolveIncident(ctx context.Context, companyID, incidentID, correlationID string) (appdto.CollectorIncidentRaw, error) {
	return observe.Call(d.cfg, "ResolveIncident", "POST", "/incidents/{id}/resolve", correlationID, func() (appdto.CollectorIncidentRaw, error) {
		return d.inner.ResolveIncident(ctx, companyID, incidentID, correlationID)
	})
}

type suggestionResult struct {
	Body    appdto.CollectorIncidentSuggestionRaw
	Present bool
}

func (d *decorated) GetIncidentSuggestion(ctx context.Context, companyID, incidentID, correlationID string) (appdto.CollectorIncidentSuggestionRaw, bool, error) {
	out, err := observe.Call(d.cfg, "GetIncidentSuggestion", "GET", "/incidents/{id}/suggestion", correlationID, func() (suggestionResult, error) {
		body, present, callErr := d.inner.GetIncidentSuggestion(ctx, companyID, incidentID, correlationID)
		return suggestionResult{Body: body, Present: present}, callErr
	})
	return out.Body, out.Present, err
}

func (d *decorated) ApplyIncidentSuccessor(ctx context.Context, companyID, incidentID, correlationID string, body appdto.CollectorApplySuccessorRaw) (appdto.CollectorIncidentRaw, error) {
	return observe.Call(d.cfg, "ApplyIncidentSuccessor", "POST", "/incidents/{id}/successor", correlationID, func() (appdto.CollectorIncidentRaw, error) {
		return d.inner.ApplyIncidentSuccessor(ctx, companyID, incidentID, correlationID, body)
	})
}
