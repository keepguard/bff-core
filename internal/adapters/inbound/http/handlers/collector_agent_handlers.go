package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CollectorAgentHandlers struct {
	collectorClient client.CollectorClient
	companyClient   client.CompanyClient
	knowledgeClient client.KnowledgeClient
	logger          *zap.Logger
}

func NewCollectorAgentHandlers(
	collectorClient client.CollectorClient,
	companyClient client.CompanyClient,
	knowledgeClient client.KnowledgeClient,
	logger *zap.Logger,
) *CollectorAgentHandlers {
	return &CollectorAgentHandlers{
		collectorClient: collectorClient,
		companyClient:   companyClient,
		knowledgeClient: knowledgeClient,
		logger:          logger,
	}
}

type collectorAgentCreateBody struct {
	Name            string                      `json:"name"`
	Description     string                      `json:"description,omitempty"`
	Context         string                      `json:"context,omitempty"`
	CollectorType   string                      `json:"collectorType"`
	CollectorConfig json.RawMessage             `json:"collectorConfig"`
	Prompt          string                      `json:"prompt,omitempty"`
	Schedule        appdto.CollectorScheduleDTO `json:"schedule"`
	Enabled         *bool                       `json:"enabled,omitempty"`
	DataSourceID    string                      `json:"dataSourceId,omitempty"`
}

type collectorAgentUpdateBody struct {
	Name            *string                      `json:"name,omitempty"`
	Description     *string                      `json:"description,omitempty"`
	Context         *string                      `json:"context,omitempty"`
	CollectorConfig json.RawMessage              `json:"collectorConfig,omitempty"`
	Prompt          *string                      `json:"prompt,omitempty"`
	Schedule        *appdto.CollectorScheduleDTO `json:"schedule,omitempty"`
	DataSourceID    *string                      `json:"dataSourceId,omitempty"`
}

func (h *CollectorAgentHandlers) ListCollectorAgentsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	query := map[string]string{}
	for _, key := range []string{"q", "enabled", "collector_type", "data_source_id", "page", "size", "sort", "dir"} {
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
	if query["page"] == "" {
		query["page"] = "0"
	}
	raw, err := h.collectorClient.SearchAgents(c.Request().Context(), companyID, correlationID, query)
	if err != nil {
		h.logger.Error("Erro ao listar agents",
			zap.String("correlationId", correlationID),
			zap.String("companyId", companyID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapPaginatedCollectorAgents(raw))
}

func (h *CollectorAgentHandlers) GetCollectorAgentHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	raw, err := h.collectorClient.GetAgent(c.Request().Context(), companyID, c.Param("id"), correlationID)
	if err != nil {
		h.logger.Error("Erro ao obter agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapCollectorAgentRaw(raw))
}

func (h *CollectorAgentHandlers) CreateCollectorAgentHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	var body collectorAgentCreateBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_BODY",
			Message:       "JSON inválido",
			CorrelationID: correlationID,
		})
	}
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(body.Name) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_NAME",
			Message:       "name é obrigatório",
			CorrelationID: correlationID,
		})
	}
	if strings.TrimSpace(body.CollectorType) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_COLLECTOR_TYPE",
			Message:       "collectorType é obrigatório",
			CorrelationID: correlationID,
		})
	}
	enabled := false
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	contextLabel := strings.TrimSpace(body.Context)
	if contextLabel == "" {
		contextLabel = "geral"
	}
	schedule := appdto.MapCollectorScheduleDTO(body.Schedule)
	raw, err := h.collectorClient.CreateAgent(c.Request().Context(), companyID, correlationID, appdto.CollectorAgentWriteRaw{
		Name:            body.Name,
		Description:     optionalString(body.Description),
		Context:         &contextLabel,
		CollectorType:   body.CollectorType,
		CollectorConfig: body.CollectorConfig,
		Prompt:          optionalString(body.Prompt),
		Schedule:        &schedule,
		Enabled:         &enabled,
		DataSourceID:    optionalString(body.DataSourceID),
	})
	if err != nil {
		h.logger.Error("Erro ao criar agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusCreated, appdto.MapCollectorAgentRaw(raw))
}

func (h *CollectorAgentHandlers) UpdateCollectorAgentHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	var body collectorAgentUpdateBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_BODY",
			Message:       "JSON inválido",
			CorrelationID: correlationID,
		})
	}
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	write := appdto.CollectorAgentWriteRaw{
		CollectorConfig: body.CollectorConfig,
		Prompt:          body.Prompt,
		Description:     body.Description,
		Context:         body.Context,
	}
	if body.Name != nil {
		write.Name = *body.Name
	}
	if body.Schedule != nil {
		schedule := appdto.MapCollectorScheduleDTO(*body.Schedule)
		write.Schedule = &schedule
	}
	if body.DataSourceID != nil {
		write.DataSourceID = body.DataSourceID
	}
	raw, err := h.collectorClient.UpdateAgent(c.Request().Context(), companyID, c.Param("id"), correlationID, write)
	if err != nil {
		h.logger.Error("Erro ao atualizar agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapCollectorAgentRaw(raw))
}

func (h *CollectorAgentHandlers) EnableCollectorAgentHandler(c echo.Context) error {
	return h.toggle(c, true)
}

func (h *CollectorAgentHandlers) DisableCollectorAgentHandler(c echo.Context) error {
	return h.toggle(c, false)
}

func (h *CollectorAgentHandlers) toggle(c echo.Context, enable bool) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	var raw appdto.CollectorAgentRaw
	if enable {
		raw, err = h.collectorClient.EnableAgent(c.Request().Context(), companyID, c.Param("id"), correlationID)
	} else {
		raw, err = h.collectorClient.DisableAgent(c.Request().Context(), companyID, c.Param("id"), correlationID)
	}
	if err != nil {
		h.logger.Error("Erro ao alterar status do agent",
			zap.String("correlationId", correlationID),
			zap.Bool("enable", enable),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapCollectorAgentRaw(raw))
}

func (h *CollectorAgentHandlers) DeleteCollectorAgentHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	if err := h.collectorClient.DeleteAgent(c.Request().Context(), companyID, c.Param("id"), correlationID); err != nil {
		h.logger.Error("Erro ao excluir agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CollectorAgentHandlers) TestCollectorAgentHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	result, err := h.collectorClient.TestAgent(c.Request().Context(), companyID, c.Param("id"), correlationID)
	if err != nil {
		h.logger.Error("Erro ao testar agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *CollectorAgentHandlers) ListCollectorAgentExecutionsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	limit := 50
	if raw := strings.TrimSpace(c.QueryParam("limit")); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	executions, err := h.collectorClient.ListAgentExecutions(
		c.Request().Context(), companyID, c.Param("id"), correlationID, limit,
	)
	if err != nil {
		h.logger.Error("Erro ao listar execuções do agent",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapCollectorExecutions(executions))
}

func (h *CollectorAgentHandlers) GetCollectorExecutionPayloadsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	if h.knowledgeClient == nil {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Serviço de conhecimento indisponível",
			CorrelationID: correlationID,
		})
	}
	executionID := strings.TrimSpace(c.Param("executionId"))
	if executionID == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "BAD_REQUEST",
			Message:       "executionId é obrigatório",
			CorrelationID: correlationID,
		})
	}
	execution, execErr := h.collectorClient.GetExecution(c.Request().Context(), companyID, executionID, correlationID)
	if execErr != nil {
		h.logger.Error("Erro ao buscar execução",
			zap.String("correlationId", correlationID),
			zap.String("executionId", executionID),
			zap.Error(execErr),
		)
		return handleError(c, execErr, correlationID)
	}
	items, loadErr := h.loadExecutionPayloads(c, companyID, correlationID, execution)
	if loadErr != nil {
		h.logger.Error("Erro ao carregar payloads da execução",
			zap.String("correlationId", correlationID),
			zap.String("executionId", executionID),
			zap.Error(loadErr),
		)
		return handleError(c, loadErr, correlationID)
	}
	if items == nil {
		items = []appdto.ExecutionPayloadItemDTO{}
	}
	return c.JSON(http.StatusOK, items)
}

func (h *CollectorAgentHandlers) loadExecutionPayloads(
	c echo.Context, companyID, correlationID string, execution appdto.CollectorExecutionRaw,
) ([]appdto.ExecutionPayloadItemDTO, error) {
	bearer := bearerFrom(c)
	ctx := c.Request().Context()
	refs := parsePayloadRefs(execution.Metadata)
	if len(refs) > 0 {
		items := make([]appdto.ExecutionPayloadItemDTO, 0, len(refs))
		for _, ref := range refs {
			item, err := h.loadPayloadRef(ctx, companyID, bearer, correlationID, ref)
			if err != nil {
				if isNotFound(err) {
					continue
				}
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	}
	if strings.TrimSpace(execution.AgentID) == "" || strings.TrimSpace(execution.StartedAt) == "" {
		return []appdto.ExecutionPayloadItemDTO{}, nil
	}
	results, err := h.knowledgeClient.GetCollectionResults(
		ctx, companyID, bearer, correlationID, execution.AgentID, execution.StartedAt, 60,
	)
	if err != nil {
		return nil, err
	}
	items := make([]appdto.ExecutionPayloadItemDTO, 0, len(results.Snapshots)+len(results.Documents))
	for _, snapshot := range results.Snapshots {
		items = append(items, snapshotToPayloadItem(snapshot))
	}
	for _, document := range results.Documents {
		items = append(items, documentToPayloadItem(document))
	}
	return items, nil
}

type payloadRef struct {
	Kind string
	ID   string
}

func parsePayloadRefs(metadata map[string]any) []payloadRef {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata["payload_refs"]
	if !ok || raw == nil {
		return nil
	}
	var rows []any
	switch typed := raw.(type) {
	case []any:
		rows = typed
	case []map[string]any:
		for _, item := range typed {
			rows = append(rows, item)
		}
	default:
		return nil
	}
	refs := make([]payloadRef, 0, len(rows))
	for _, row := range rows {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		kind, _ := item["kind"].(string)
		id, _ := item["id"].(string)
		kind = strings.TrimSpace(strings.ToLower(kind))
		id = strings.TrimSpace(id)
		if (kind != "snapshot" && kind != "document") || id == "" {
			continue
		}
		refs = append(refs, payloadRef{Kind: kind, ID: id})
	}
	return refs
}

func (h *CollectorAgentHandlers) loadPayloadRef(
	ctx context.Context, companyID, bearer, correlationID string, ref payloadRef,
) (appdto.ExecutionPayloadItemDTO, error) {
	switch ref.Kind {
	case "snapshot":
		snapshot, err := h.knowledgeClient.GetSnapshot(ctx, companyID, bearer, correlationID, ref.ID)
		if err != nil {
			return appdto.ExecutionPayloadItemDTO{}, err
		}
		return snapshotToPayloadItem(snapshot), nil
	case "document":
		document, err := h.knowledgeClient.GetDocumentPreview(ctx, companyID, bearer, correlationID, ref.ID)
		if err != nil {
			return appdto.ExecutionPayloadItemDTO{}, err
		}
		return documentToPayloadItem(document), nil
	default:
		return appdto.ExecutionPayloadItemDTO{}, nil
	}
}

func snapshotToPayloadItem(snapshot appdto.KnowledgeSnapshotDTO) appdto.ExecutionPayloadItemDTO {
	return appdto.ExecutionPayloadItemDTO{
		Kind:        "snapshot",
		ID:          snapshot.ID,
		ContentType: "application/json",
		Payload:     snapshot.Payload,
		Metadata: map[string]any{
			"collectorType": snapshot.CollectorType,
			"entityHint":    snapshot.EntityHint,
			"collectedAt":   snapshot.CollectedAt,
			"schema":        snapshot.Schema,
			"sourceUrl":     snapshot.SourceURL,
		},
	}
}

func documentToPayloadItem(document appdto.KnowledgeDocumentPreviewDTO) appdto.ExecutionPayloadItemDTO {
	return appdto.ExecutionPayloadItemDTO{
		Kind:        "document",
		ID:          document.ID,
		ContentType: document.ContentType,
		FileName:    document.FileName,
		PreviewText: document.PreviewText,
		Metadata: map[string]any{
			"entityHint":       document.EntityHint,
			"collectedAt":      document.CollectedAt,
			"status":           document.Status,
			"previewAvailable": document.PreviewAvailable,
			"message":          document.Message,
		},
	}
}

func isNotFound(err error) bool {
	httpErr, ok := err.(*appdto.HTTPError)
	return ok && httpErr.Code == http.StatusNotFound
}

func (h *CollectorAgentHandlers) ListCollectorDataSourcesHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	query := map[string]string{}
	if include := strings.TrimSpace(c.QueryParam("includeDisabled")); include != "" {
		query["include_disabled"] = include
	}
	if include := strings.TrimSpace(c.QueryParam("include_disabled")); include != "" {
		query["include_disabled"] = include
	}
	raw, err := h.collectorClient.ListDataSources(c.Request().Context(), companyID, correlationID, query)
	if err != nil {
		h.logger.Error("Erro ao listar fontes de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapCollectorDataSources(raw))
}

func (h *CollectorAgentHandlers) GetCollectorDataSourceHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	raw, err := h.collectorClient.GetDataSource(c.Request().Context(), companyID, c.Param("id"), correlationID)
	if err != nil {
		h.logger.Error("Erro ao buscar fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapCollectorDataSourceRaw(raw))
}

type collectorDataSourceCreateBody struct {
	Name                string                      `json:"name"`
	Slug                string                      `json:"slug"`
	Description         string                      `json:"description,omitempty"`
	WebsiteURL          string                      `json:"websiteUrl,omitempty"`
	CollectorType       string                      `json:"collectorType"`
	NameTemplate        string                      `json:"nameTemplate,omitempty"`
	DescriptionTemplate string                      `json:"descriptionTemplate,omitempty"`
	PromptTemplate      string                      `json:"promptTemplate,omitempty"`
	DefaultContext      string                      `json:"defaultContext,omitempty"`
	DefaultSchedule     appdto.CollectorScheduleDTO `json:"defaultSchedule"`
	ConfigTemplate      json.RawMessage             `json:"configTemplate"`
	Variables           json.RawMessage             `json:"variables"`
	Notes               string                      `json:"notes,omitempty"`
	Enabled             *bool                       `json:"enabled,omitempty"`
	RateLimit           json.RawMessage             `json:"rateLimit,omitempty"`
}

type collectorDataSourceUpdateBody struct {
	Name                *string                      `json:"name,omitempty"`
	Slug                *string                      `json:"slug,omitempty"`
	Description         *string                      `json:"description,omitempty"`
	WebsiteURL          *string                      `json:"websiteUrl,omitempty"`
	NameTemplate        *string                      `json:"nameTemplate,omitempty"`
	DescriptionTemplate *string                      `json:"descriptionTemplate,omitempty"`
	PromptTemplate      *string                      `json:"promptTemplate,omitempty"`
	DefaultContext      *string                      `json:"defaultContext,omitempty"`
	DefaultSchedule     *appdto.CollectorScheduleDTO `json:"defaultSchedule,omitempty"`
	ConfigTemplate      json.RawMessage              `json:"configTemplate,omitempty"`
	Variables           json.RawMessage              `json:"variables,omitempty"`
	Notes               *string                      `json:"notes,omitempty"`
	RateLimit           json.RawMessage              `json:"rateLimit,omitempty"`
}

func (h *CollectorAgentHandlers) CreateCollectorDataSourceHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	var body collectorDataSourceCreateBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_BODY",
			Message:       "JSON inválido",
			CorrelationID: correlationID,
		})
	}
	if strings.TrimSpace(body.Name) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_NAME",
			Message:       "name é obrigatório",
			CorrelationID: correlationID,
		})
	}
	if strings.TrimSpace(body.Slug) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_SLUG",
			Message:       "slug é obrigatório",
			CorrelationID: correlationID,
		})
	}
	if strings.TrimSpace(body.CollectorType) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_COLLECTOR_TYPE",
			Message:       "collectorType é obrigatório",
			CorrelationID: correlationID,
		})
	}
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	schedule, err := json.Marshal(appdto.MapCollectorScheduleDTO(body.DefaultSchedule))
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_SCHEDULE",
			Message:       "defaultSchedule inválido",
			CorrelationID: correlationID,
		})
	}
	raw, err := h.collectorClient.CreateDataSource(c.Request().Context(), companyID, correlationID, appdto.CollectorDataSourceWriteRaw{
		Name:                body.Name,
		Slug:                body.Slug,
		Description:         optionalString(body.Description),
		WebsiteURL:          optionalString(body.WebsiteURL),
		CollectorType:       body.CollectorType,
		NameTemplate:        optionalString(body.NameTemplate),
		DescriptionTemplate: optionalString(body.DescriptionTemplate),
		PromptTemplate:      optionalString(body.PromptTemplate),
		DefaultContext:      optionalString(body.DefaultContext),
		DefaultSchedule:     schedule,
		ConfigTemplate:      body.ConfigTemplate,
		Variables:           body.Variables,
		Notes:               optionalString(body.Notes),
		Enabled:             body.Enabled,
		RateLimit:           body.RateLimit,
	})
	if err != nil {
		h.logger.Error("Erro ao criar fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusCreated, appdto.MapCollectorDataSourceRaw(raw))
}

func (h *CollectorAgentHandlers) UpdateCollectorDataSourceHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	var body collectorDataSourceUpdateBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_BODY",
			Message:       "JSON inválido",
			CorrelationID: correlationID,
		})
	}
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	write := appdto.CollectorDataSourceWriteRaw{
		Description:         body.Description,
		WebsiteURL:          body.WebsiteURL,
		NameTemplate:        body.NameTemplate,
		DescriptionTemplate: body.DescriptionTemplate,
		PromptTemplate:      body.PromptTemplate,
		DefaultContext:      body.DefaultContext,
		ConfigTemplate:      body.ConfigTemplate,
		Variables:           body.Variables,
		Notes:               body.Notes,
		RateLimit:           body.RateLimit,
	}
	if body.Name != nil {
		write.Name = *body.Name
	}
	if body.Slug != nil {
		write.Slug = *body.Slug
	}
	if body.DefaultSchedule != nil {
		schedule, marshalErr := json.Marshal(appdto.MapCollectorScheduleDTO(*body.DefaultSchedule))
		if marshalErr != nil {
			return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
				Error:         "INVALID_SCHEDULE",
				Message:       "defaultSchedule inválido",
				CorrelationID: correlationID,
			})
		}
		write.DefaultSchedule = schedule
	}
	raw, err := h.collectorClient.UpdateDataSource(c.Request().Context(), companyID, c.Param("id"), correlationID, write)
	if err != nil {
		h.logger.Error("Erro ao atualizar fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapCollectorDataSourceRaw(raw))
}

func (h *CollectorAgentHandlers) EnableCollectorDataSourceHandler(c echo.Context) error {
	return h.toggleDataSource(c, true)
}

func (h *CollectorAgentHandlers) DisableCollectorDataSourceHandler(c echo.Context) error {
	return h.toggleDataSource(c, false)
}

func (h *CollectorAgentHandlers) toggleDataSource(c echo.Context, enable bool) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	var raw appdto.CollectorDataSourceRaw
	if enable {
		raw, err = h.collectorClient.EnableDataSource(c.Request().Context(), companyID, c.Param("id"), correlationID)
	} else {
		raw, err = h.collectorClient.DisableDataSource(c.Request().Context(), companyID, c.Param("id"), correlationID)
	}
	if err != nil {
		h.logger.Error("Erro ao alterar status da fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapCollectorDataSourceRaw(raw))
}

func (h *CollectorAgentHandlers) DeleteCollectorDataSourceHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	if err := h.collectorClient.DeleteDataSource(c.Request().Context(), companyID, c.Param("id"), correlationID); err != nil {
		h.logger.Error("Erro ao excluir fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.NoContent(http.StatusNoContent)
}

type collectorPropagateBody struct {
	Fields []string `json:"fields"`
	DryRun bool     `json:"dryRun"`
	Limit  int      `json:"limit,omitempty"`
}

func (h *CollectorAgentHandlers) PropagateCollectorDataSourceHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	var body collectorPropagateBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_BODY",
			Message:       "JSON inválido",
			CorrelationID: correlationID,
		})
	}
	if len(body.Fields) == 0 {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_FIELDS",
			Message:       "fields é obrigatório",
			CorrelationID: correlationID,
		})
	}
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	raw, err := h.collectorClient.PropagateDataSource(c.Request().Context(), companyID, c.Param("id"), correlationID, appdto.PropagateDataSourceWriteRaw{
		Fields: body.Fields,
		DryRun: body.DryRun,
		Limit:  body.Limit,
	})
	if err != nil {
		h.logger.Error("Erro ao propagar fonte de dados",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, appdto.MapPropagateDataSourceRaw(raw))
}

func (h *CollectorAgentHandlers) resolveCompany(c echo.Context, correlationID string) (string, error) {
	if h.collectorClient == nil || h.companyClient == nil {
		return "", c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Gestão de agents indisponível",
			CorrelationID: correlationID,
		})
	}
	if companyID := client.CompanyIDFromContext(c.Request().Context()); companyID != "" {
		return companyID, nil
	}
	tenantID := middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	if strings.TrimSpace(tenantID) == "" {
		return "", c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "tenantId do JWT é obrigatório",
			CorrelationID: correlationID,
		})
	}
	company, err := h.companyClient.GetByTenantId(c.Request().Context(), tenantID, correlationID)
	if err != nil {
		h.logger.Error("Erro ao resolver company pelo tenant do JWT",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantID),
			zap.Error(err),
		)
		return "", handleError(c, err, correlationID)
	}
	if company.ID == "" {
		return "", c.JSON(http.StatusNotFound, pkg.ErrorResponse{
			Error:         "COMPANY_NOT_FOUND",
			Message:       "Empresa não encontrada para o tenant autenticado",
			CorrelationID: correlationID,
		})
	}
	return company.ID, nil
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
