package handlers

import (
	"encoding/json"
	"net/http"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/llm"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Bodies de provider/alert/complete são json opaco: o mapper não tipa o contrato do gateway.

type LlmHandlers struct {
	llm    llm.LlmPort
	logger *zap.Logger
}

func NewLlmHandlers(llmPort llm.LlmPort, logger *zap.Logger) *LlmHandlers {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LlmHandlers{llm: llmPort, logger: logger}
}

func (h *LlmHandlers) ListLlmProvidersHandler(c echo.Context) error {
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.ListProviders(ctx.Request().Context(), query)
	})
}

func (h *LlmHandlers) CreateLlmProviderHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	return h.proxyRaw(c, http.StatusCreated, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.CreateProvider(ctx.Request().Context(), llm.ProviderIDCommand{TenantQuery: query, Body: body})
	})
}

func (h *LlmHandlers) UpdateLlmProviderHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.UpdateProvider(ctx.Request().Context(), llm.ProviderIDCommand{TenantQuery: query, ID: id, Body: body})
	})
}

func (h *LlmHandlers) EnableLlmProviderHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.SetProviderEnabled(ctx.Request().Context(), llm.ProviderIDCommand{TenantQuery: query, ID: id, Enabled: true})
	})
}

func (h *LlmHandlers) DisableLlmProviderHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.SetProviderEnabled(ctx.Request().Context(), llm.ProviderIDCommand{TenantQuery: query, ID: id, Enabled: false})
	})
}

func (h *LlmHandlers) CompleteLlmHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	query, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	result, callErr := h.llm.Complete(c.Request().Context(), llm.CompleteCommand{
		TenantQuery: query,
		CompanyID:   port.CompanyIDFromContext(c.Request().Context()),
		Body:        body,
	})
	if callErr != nil {
		h.logger.Error("Erro ao completar LLM", zap.String("correlationId", query.CorrelationID), zap.Error(callErr))
		return handleError(c, callErr, query.CorrelationID)
	}
	return writeRaw(c, http.StatusOK, result)
}

func (h *LlmHandlers) ListLlmUsageHandler(c echo.Context) error {
	query, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	filters := map[string]string{}
	for _, key := range []string{
		"page", "size", "from", "to", "companyId", "providerType", "model",
		"feature", "sourceService", "outcome", "sort", "dir",
	} {
		if value := c.QueryParam(key); value != "" {
			filters[key] = value
		}
	}
	result, err := h.llm.ListUsage(c.Request().Context(), llm.ListUsageQuery{TenantQuery: query, Filters: filters})
	if err != nil {
		h.logger.Error("Erro ao listar uso LLM", zap.String("correlationId", query.CorrelationID), zap.Error(err))
		return handleError(c, err, query.CorrelationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *LlmHandlers) GetLlmUsageHandler(c echo.Context) error {
	query, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	id := c.Param("id")
	result, err := h.llm.GetUsage(c.Request().Context(), llm.GetUsageQuery{TenantQuery: query, ID: id})
	if err != nil {
		h.logger.Error("Erro ao buscar uso LLM", zap.String("correlationId", query.CorrelationID), zap.String("id", id), zap.Error(err))
		return handleError(c, err, query.CorrelationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *LlmHandlers) ListLlmAlertRulesHandler(c echo.Context) error {
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.ListAlertRules(ctx.Request().Context(), query)
	})
}

func (h *LlmHandlers) CreateLlmAlertRuleHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	if claims := middlewarePkg.GetClaimsFromContext(c); claims != nil {
		if _, ok := body["createdBy"]; !ok && claims.CodeUser != "" {
			body["createdBy"] = claims.CodeUser
		}
	}
	return h.proxyRaw(c, http.StatusCreated, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.CreateAlertRule(ctx.Request().Context(), llm.AlertRuleCommand{TenantQuery: query, Body: body})
	})
}

func (h *LlmHandlers) UpdateLlmAlertRuleHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.UpdateAlertRule(ctx.Request().Context(), llm.AlertRuleCommand{TenantQuery: query, ID: id, Body: body})
	})
}

func (h *LlmHandlers) EnableLlmAlertRuleHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.SetAlertRuleEnabled(ctx.Request().Context(), llm.AlertRuleCommand{TenantQuery: query, ID: id, Enabled: true})
	})
}

func (h *LlmHandlers) DisableLlmAlertRuleHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.SetAlertRuleEnabled(ctx.Request().Context(), llm.AlertRuleCommand{TenantQuery: query, ID: id, Enabled: false})
	})
}

func (h *LlmHandlers) ListLlmAlertFiringsHandler(c echo.Context) error {
	filters := map[string]string{}
	for _, key := range []string{"page", "size"} {
		if value := c.QueryParam(key); value != "" {
			filters[key] = value
		}
	}
	return h.proxyRaw(c, http.StatusOK, func(ctx echo.Context, query llm.TenantQuery) (json.RawMessage, error) {
		return h.llm.ListAlertFirings(ctx.Request().Context(), llm.ListAlertFiringsQuery{TenantQuery: query, Filters: filters})
	})
}

func (h *LlmHandlers) proxyRaw(c echo.Context, status int, fn func(echo.Context, llm.TenantQuery) (json.RawMessage, error)) error {
	query, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	result, err := fn(c, query)
	if err != nil {
		h.logger.Error("Erro no gateway LLM", zap.String("correlationId", query.CorrelationID), zap.Error(err))
		return handleError(c, err, query.CorrelationID)
	}
	return writeRaw(c, status, result)
}

func (h *LlmHandlers) guard(c echo.Context) (llm.TenantQuery, error) {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.llm == nil {
		return llm.TenantQuery{}, c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Consulta LLM indisponível",
			CorrelationID: correlationID,
		})
	}
	tenantID := middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	if tenantID == "" {
		return llm.TenantQuery{}, c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "tenant_id do token JWT é obrigatório",
			CorrelationID: correlationID,
		})
	}
	if token := middlewarePkg.GetTokenFromContext(c); token != "" {
		ctx := port.WithBearerToken(c.Request().Context(), token)
		c.SetRequest(c.Request().WithContext(ctx))
	}
	return llm.TenantQuery{TenantID: tenantID, CorrelationID: correlationID}, nil
}

func bindJSON(c echo.Context) (map[string]any, error) {
	correlationID := middlewarePkg.GetCorrelationID(c)
	body := map[string]any{}
	if err := c.Bind(&body); err != nil {
		return nil, c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "BAD_REQUEST",
			Message:       "Payload inválido",
			CorrelationID: correlationID,
		})
	}
	if body == nil {
		body = map[string]any{}
	}
	return body, nil
}

func writeRaw(c echo.Context, status int, raw json.RawMessage) error {
	if len(raw) == 0 {
		raw = json.RawMessage("null")
	}
	return c.Blob(status, "application/json", raw)
}
