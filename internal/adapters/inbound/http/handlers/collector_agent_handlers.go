package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-core/internal/adapters/inbound/http/mapper"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/collector"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CollectorAgentHandlers struct {
	collector collector.CollectorPort
	logger    *zap.Logger
}

func NewCollectorAgentHandlers(port collector.CollectorPort, logger *zap.Logger) *CollectorAgentHandlers {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CollectorAgentHandlers{collector: port, logger: logger}
}

func (h *CollectorAgentHandlers) requestScope(c echo.Context) (correlationID, tenantID, companyFromCtx string) {
	correlationID = middlewarePkg.GetCorrelationID(c)
	tenantID = middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	companyFromCtx = port.CompanyIDFromContext(c.Request().Context())
	return
}

func collectorInvalidBody(c echo.Context, correlationID string) error {
	return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
		Error:         "INVALID_BODY",
		Message:       "JSON inválido",
		CorrelationID: correlationID,
	})
}

func (h *CollectorAgentHandlers) ListCollectorAgentsHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	query := map[string]string{}
	for _, key := range []string{
		"q", "enabled", "collector_type", "data_source_id",
		"last_execution_status", "has_open_incident", "page", "size", "sort", "dir",
	} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	if value := c.QueryParam("collectorType"); value != "" {
		query["collector_type"] = value
	}
	if value := c.QueryParam("dataSourceId"); value != "" {
		query["data_source_id"] = value
	}
	if value := c.QueryParam("lastExecutionStatus"); value != "" {
		query["last_execution_status"] = value
	}
	if value := c.QueryParam("has_open_incident"); value != "" {
		query["has_open_incident"] = value
	}
	if value := c.QueryParam("hasOpenIncident"); value != "" {
		query["has_open_incident"] = value
	}
	if query["page"] == "" {
		query["page"] = "0"
	}
	result, err := h.collector.ListAgents(c.Request().Context(), collector.ListAgentsQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		Filters:        query,
	})
	if err != nil {
		h.logger.Error("Erro ao listar agents",
			zap.String("correlationId", correlationID),
			zap.String("companyId", companyFromCtx),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) GetCollectorAgentHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, err := h.collector.GetAgent(c.Request().Context(), collector.GetAgentQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao obter agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) CreateCollectorAgentHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	var body inboundDto.CollectorAgentCreateRequestDTO
	if err := c.Bind(&body); err != nil {
		return collectorInvalidBody(c, correlationID)
	}
	result, err := h.collector.CreateAgent(c.Request().Context(), mapper.ToCreateCollectorAgentCommand(body, companyFromCtx, tenantID, correlationID))
	if err != nil {
		h.logger.Error("Erro ao criar agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *CollectorAgentHandlers) UpdateCollectorAgentHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	var body inboundDto.CollectorAgentUpdateRequestDTO
	if err := c.Bind(&body); err != nil {
		return collectorInvalidBody(c, correlationID)
	}
	result, err := h.collector.UpdateAgent(c.Request().Context(), mapper.ToUpdateCollectorAgentCommand(body, companyFromCtx, tenantID, correlationID, c.Param("id")))
	if err != nil {
		h.logger.Error("Erro ao atualizar agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) EnableCollectorAgentHandler(c echo.Context) error {
	return h.toggleAgent(c, true)
}

func (h *CollectorAgentHandlers) DisableCollectorAgentHandler(c echo.Context) error {
	return h.toggleAgent(c, false)
}

func (h *CollectorAgentHandlers) toggleAgent(c echo.Context, enable bool) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	cmd := collector.AgentIDCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	}
	var (
		result appdto.CollectorAgentDetailDTO
		err    error
	)
	if enable {
		result, err = h.collector.EnableAgent(c.Request().Context(), cmd)
	} else {
		result, err = h.collector.DisableAgent(c.Request().Context(), cmd)
	}
	if err != nil {
		h.logger.Error("Erro ao alterar status do agent",
			zap.String("correlationId", correlationID),
			zap.Bool("enable", enable),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) DeleteCollectorAgentHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	if err := h.collector.DeleteAgent(c.Request().Context(), collector.AgentIDCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	}); err != nil {
		h.logger.Error("Erro ao excluir agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CollectorAgentHandlers) TestCollectorAgentHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, err := h.collector.TestAgent(c.Request().Context(), collector.AgentIDCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao testar agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) RunCollectorAgentHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, err := h.collector.RunAgent(c.Request().Context(), collector.AgentIDCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao executar agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusAccepted, result)
}

func (h *CollectorAgentHandlers) BulkCollectorAgentsHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	var body inboundDto.CollectorBulkRequestDTO
	if bindErr := c.Bind(&body); bindErr != nil {
		return collectorInvalidBody(c, correlationID)
	}
	result, status, err := h.collector.BulkAgents(c.Request().Context(), mapper.ToBulkCollectorAgentsCommand(body, companyFromCtx, tenantID, correlationID))
	if err != nil {
		h.logger.Error("Erro ao executar lote de agents",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	if status == 0 {
		status = http.StatusOK
	}
	return c.JSON(status, result)
}

func (h *CollectorAgentHandlers) GetCollectorBulkOperationHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, err := h.collector.GetBulk(c.Request().Context(), collector.GetBulkQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao obter lote de agents",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) GetCollectorActiveBulkOperationHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, err := h.collector.GetActiveBulk(c.Request().Context(), collector.GetActiveBulkQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
	})
	if err != nil {
		if httpErr, ok := err.(*appdto.HTTPError); !ok || httpErr.Code != http.StatusNotFound {
			h.logger.Error("Erro ao obter lote ativo de agents",
				zap.String("correlationId", correlationID),
				zap.Error(err),
			)
		}
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) ListCollectorAgentExecutionsHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	limit := 50
	if raw := strings.TrimSpace(c.QueryParam("limit")); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	result, err := h.collector.ListExecutions(c.Request().Context(), collector.ListExecutionsQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		AgentID:        c.Param("id"),
		Limit:          limit,
	})
	if err != nil {
		h.logger.Error("Erro ao listar execuções do agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) GetCollectorExecutionPayloadsHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	items, err := h.collector.GetExecutionPayloads(c.Request().Context(), collector.GetExecutionPayloadsQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ExecutionID:    c.Param("executionId"),
	})
	if err != nil {
		h.logger.Error("Erro ao carregar payloads da execução",
			zap.String("correlationId", correlationID),
			zap.String("executionId", strings.TrimSpace(c.Param("executionId"))),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *CollectorAgentHandlers) ListCollectorDataSourcesHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	query := map[string]string{}
	if include := strings.TrimSpace(c.QueryParam("includeDisabled")); include != "" {
		query["include_disabled"] = include
	}
	if include := strings.TrimSpace(c.QueryParam("include_disabled")); include != "" {
		query["include_disabled"] = include
	}
	result, err := h.collector.ListDataSources(c.Request().Context(), collector.ListDataSourcesQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		Filters:        query,
	})
	if err != nil {
		h.logger.Error("Erro ao listar fontes de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) GetCollectorDataSourceHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, err := h.collector.GetDataSource(c.Request().Context(), collector.GetDataSourceQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao buscar fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) CreateCollectorDataSourceHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	var body inboundDto.CollectorDataSourceCreateRequestDTO
	if err := c.Bind(&body); err != nil {
		return collectorInvalidBody(c, correlationID)
	}
	result, err := h.collector.CreateDataSource(c.Request().Context(), mapper.ToCreateCollectorDataSourceCommand(body, companyFromCtx, tenantID, correlationID))
	if err != nil {
		h.logger.Error("Erro ao criar fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *CollectorAgentHandlers) UpdateCollectorDataSourceHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	var body inboundDto.CollectorDataSourceUpdateRequestDTO
	if err := c.Bind(&body); err != nil {
		return collectorInvalidBody(c, correlationID)
	}
	result, err := h.collector.UpdateDataSource(c.Request().Context(), mapper.ToUpdateCollectorDataSourceCommand(body, companyFromCtx, tenantID, correlationID, c.Param("id")))
	if err != nil {
		h.logger.Error("Erro ao atualizar fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) EnableCollectorDataSourceHandler(c echo.Context) error {
	return h.toggleDataSource(c, true)
}

func (h *CollectorAgentHandlers) DisableCollectorDataSourceHandler(c echo.Context) error {
	return h.toggleDataSource(c, false)
}

func (h *CollectorAgentHandlers) toggleDataSource(c echo.Context, enable bool) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	cmd := collector.DataSourceIDCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	}
	var (
		result appdto.CollectorDataSourceDTO
		err    error
	)
	if enable {
		result, err = h.collector.EnableDataSource(c.Request().Context(), cmd)
	} else {
		result, err = h.collector.DisableDataSource(c.Request().Context(), cmd)
	}
	if err != nil {
		h.logger.Error("Erro ao alterar status da fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) DeleteCollectorDataSourceHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	if err := h.collector.DeleteDataSource(c.Request().Context(), collector.DataSourceIDCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	}); err != nil {
		h.logger.Error("Erro ao excluir fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CollectorAgentHandlers) PropagateCollectorDataSourceHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	var body inboundDto.CollectorPropagateRequestDTO
	if err := c.Bind(&body); err != nil {
		return collectorInvalidBody(c, correlationID)
	}
	result, err := h.collector.PropagateDataSource(c.Request().Context(), mapper.ToPropagateCollectorDataSourceCommand(body, companyFromCtx, tenantID, correlationID, c.Param("id")))
	if err != nil {
		h.logger.Error("Erro ao propagar fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) ListCollectorIncidentsHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	query := map[string]string{}
	for _, key := range []string{"status", "classification", "agent_id", "page", "size"} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	if value := c.QueryParam("agentId"); value != "" {
		query["agent_id"] = value
	}
	result, err := h.collector.ListIncidents(c.Request().Context(), collector.ListIncidentsQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		Filters:        query,
	})
	if err != nil {
		h.logger.Error("Erro ao listar incidentes de coleta",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) ListCollectorAgentIncidentsHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, err := h.collector.ListAgentIncidents(c.Request().Context(), collector.ListAgentIncidentsQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		AgentID:        c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao listar incidentes do agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) AcknowledgeCollectorIncidentHandler(c echo.Context) error {
	return h.mutateCollectorIncident(c, h.collector.AcknowledgeIncident)
}

func (h *CollectorAgentHandlers) ResolveCollectorIncidentHandler(c echo.Context) error {
	return h.mutateCollectorIncident(c, h.collector.ResolveIncident)
}

func (h *CollectorAgentHandlers) mutateCollectorIncident(
	c echo.Context,
	fn func(ctx context.Context, cmd collector.MutateIncidentCommand) (appdto.CollectorIncidentDTO, error),
) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, err := fn(c.Request().Context(), collector.MutateIncidentCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ActorUserID:    middlewarePkg.GetUserIDFromContext(c),
		ID:             c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao atualizar incidente de coleta",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) GetCollectorIncidentSuggestionHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	result, found, err := h.collector.GetIncidentSuggestion(c.Request().Context(), collector.GetIncidentSuggestionQuery{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao obter sugestão de sucessor",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	if !found {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) ApplyCollectorIncidentSuccessorHandler(c echo.Context) error {
	correlationID, tenantID, companyFromCtx := h.requestScope(c)
	var body inboundDto.CollectorApplySuccessorRequestDTO
	if bindErr := c.Bind(&body); bindErr != nil {
		return collectorInvalidBody(c, correlationID)
	}
	result, err := h.collector.ApplyIncidentSuccessor(
		c.Request().Context(),
		mapper.ToApplyCollectorIncidentSuccessorCommand(body, companyFromCtx, tenantID, correlationID, middlewarePkg.GetUserIDFromContext(c), c.Param("id")),
	)
	if err != nil {
		h.logger.Error("Erro ao aplicar sucessor",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}
