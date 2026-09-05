package handlers

import (
	"encoding/json"
	"net/http"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type LlmHandlers struct {
	llmClient domainclient.LlmClient
	logger    *zap.Logger
}

func NewLlmHandlers(llmClient domainclient.LlmClient, logger *zap.Logger) *LlmHandlers {
	return &LlmHandlers{llmClient: llmClient, logger: logger}
}

func (h *LlmHandlers) ListLlmProvidersHandler(c echo.Context) error {
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.ListProviders(c.Request().Context(), tenantID, correlationID)
	})
}

func (h *LlmHandlers) CreateLlmProviderHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	return h.proxyRaw(c, http.StatusCreated, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.CreateProvider(c.Request().Context(), tenantID, correlationID, body)
	})
}

func (h *LlmHandlers) UpdateLlmProviderHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.UpdateProvider(c.Request().Context(), tenantID, correlationID, id, body)
	})
}

func (h *LlmHandlers) EnableLlmProviderHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.SetProviderEnabled(c.Request().Context(), tenantID, correlationID, id, true)
	})
}

func (h *LlmHandlers) DisableLlmProviderHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.SetProviderEnabled(c.Request().Context(), tenantID, correlationID, id, false)
	})
}

func (h *LlmHandlers) CompleteLlmHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	correlationID, tenantID, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	companyID := domainclient.CompanyIDFromContext(c.Request().Context())
	result, callErr := h.llmClient.Complete(c.Request().Context(), tenantID, companyID, correlationID, body)
	if callErr != nil {
		h.logger.Error("Erro ao completar LLM", zap.String("correlationId", correlationID), zap.Error(callErr))
		return handleError(c, callErr, correlationID)
	}
	return writeRaw(c, http.StatusOK, result)
}

func (h *LlmHandlers) ListLlmUsageHandler(c echo.Context) error {
	correlationID, tenantID, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	query := map[string]string{}
	for _, key := range []string{
		"page", "size", "from", "to", "companyId", "providerType", "model",
		"feature", "sourceService", "outcome", "sort", "dir",
	} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	result, err := h.llmClient.ListUsage(c.Request().Context(), tenantID, correlationID, query)
	if err != nil {
		h.logger.Error("Erro ao listar uso LLM", zap.String("correlationId", correlationID), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *LlmHandlers) GetLlmUsageHandler(c echo.Context) error {
	correlationID, tenantID, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	id := c.Param("id")
	result, err := h.llmClient.GetUsage(c.Request().Context(), tenantID, correlationID, id)
	if err != nil {
		h.logger.Error("Erro ao buscar uso LLM", zap.String("correlationId", correlationID), zap.String("id", id), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *LlmHandlers) ListLlmAlertRulesHandler(c echo.Context) error {
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.ListAlertRules(c.Request().Context(), tenantID, correlationID)
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
	return h.proxyRaw(c, http.StatusCreated, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.CreateAlertRule(c.Request().Context(), tenantID, correlationID, body)
	})
}

func (h *LlmHandlers) UpdateLlmAlertRuleHandler(c echo.Context) error {
	body, err := bindJSON(c)
	if err != nil {
		return err
	}
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.UpdateAlertRule(c.Request().Context(), tenantID, correlationID, id, body)
	})
}

func (h *LlmHandlers) EnableLlmAlertRuleHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.SetAlertRuleEnabled(c.Request().Context(), tenantID, correlationID, id, true)
	})
}

func (h *LlmHandlers) DisableLlmAlertRuleHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.SetAlertRuleEnabled(c.Request().Context(), tenantID, correlationID, id, false)
	})
}

func (h *LlmHandlers) ListLlmAlertFiringsHandler(c echo.Context) error {
	query := map[string]string{}
	for _, key := range []string{"page", "size"} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	return h.proxyRaw(c, http.StatusOK, func(tenantID, correlationID string) (json.RawMessage, error) {
		return h.llmClient.ListAlertFirings(c.Request().Context(), tenantID, correlationID, query)
	})
}

func (h *LlmHandlers) proxyRaw(c echo.Context, status int, fn func(tenantID, correlationID string) (json.RawMessage, error)) error {
	correlationID, tenantID, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	result, err := fn(tenantID, correlationID)
	if err != nil {
		h.logger.Error("Erro no gateway LLM", zap.String("correlationId", correlationID), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return writeRaw(c, status, result)
}

func (h *LlmHandlers) guard(c echo.Context) (correlationID, tenantID string, err error) {
	correlationID = middlewarePkg.GetCorrelationID(c)
	if h.llmClient == nil {
		return correlationID, "", c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Consulta LLM indisponível",
			CorrelationID: correlationID,
		})
	}
	tenantID = middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	if tenantID == "" {
		return correlationID, "", c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "tenant_id do token JWT é obrigatório",
			CorrelationID: correlationID,
		})
	}
	if token := middlewarePkg.GetTokenFromContext(c); token != "" {
		ctx := domainclient.WithBearerToken(c.Request().Context(), token)
		c.SetRequest(c.Request().WithContext(ctx))
	}
	return correlationID, tenantID, nil
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
