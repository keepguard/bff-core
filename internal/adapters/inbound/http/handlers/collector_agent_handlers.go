package handlers

import (
	"encoding/json"
	"net/http"
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
	logger          *zap.Logger
}

func NewCollectorAgentHandlers(
	collectorClient client.CollectorClient,
	companyClient client.CompanyClient,
	logger *zap.Logger,
) *CollectorAgentHandlers {
	return &CollectorAgentHandlers{
		collectorClient: collectorClient,
		companyClient:   companyClient,
		logger:          logger,
	}
}

type collectorAgentCreateBody struct {
	Name            string                      `json:"name"`
	Description     string                      `json:"description,omitempty"`
	CollectorType   string                      `json:"collectorType"`
	CollectorConfig json.RawMessage             `json:"collectorConfig"`
	Prompt          string                      `json:"prompt,omitempty"`
	Schedule        appdto.CollectorScheduleDTO `json:"schedule"`
	Enabled         *bool                       `json:"enabled,omitempty"`
}

type collectorAgentUpdateBody struct {
	Name            *string                      `json:"name,omitempty"`
	Description     *string                      `json:"description,omitempty"`
	CollectorConfig json.RawMessage              `json:"collectorConfig,omitempty"`
	Prompt          *string                      `json:"prompt,omitempty"`
	Schedule        *appdto.CollectorScheduleDTO `json:"schedule,omitempty"`
}

func (h *CollectorAgentHandlers) ListCollectorAgentsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	query := map[string]string{}
	for _, key := range []string{"q", "enabled", "collector_type", "page", "size", "sort", "dir"} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	if value := c.QueryParam("collectorType"); value != "" {
		query["collector_type"] = value
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
	schedule := appdto.MapCollectorScheduleDTO(body.Schedule)
	raw, err := h.collectorClient.CreateAgent(c.Request().Context(), companyID, correlationID, appdto.CollectorAgentWriteRaw{
		Name:            body.Name,
		Description:     optionalString(body.Description),
		CollectorType:   body.CollectorType,
		CollectorConfig: body.CollectorConfig,
		Prompt:          optionalString(body.Prompt),
		Schedule:        &schedule,
		Enabled:         &enabled,
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
	}
	if body.Name != nil {
		write.Name = *body.Name
	}
	if body.Schedule != nil {
		schedule := appdto.MapCollectorScheduleDTO(*body.Schedule)
		write.Schedule = &schedule
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
