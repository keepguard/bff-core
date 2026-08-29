package handlers

import (
	"net/http"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AuditHandlers struct {
	auditClient client.AuditClient
	logger      *zap.Logger
}

func NewAuditHandlers(auditClient client.AuditClient, logger *zap.Logger) *AuditHandlers {
	return &AuditHandlers{auditClient: auditClient, logger: logger}
}

func (h *AuditHandlers) ListAuditsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.auditClient == nil {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Consulta de auditoria indisponível",
			CorrelationID: correlationID,
		})
	}
	tenantID := middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	if tenantID == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "tenant_id do token JWT é obrigatório",
			CorrelationID: correlationID,
		})
	}
	query := map[string]string{}
	for _, key := range []string{
		"page", "size", "from", "to", "actorCodeUser", "action", "outcome",
		"resourceType", "resourceId", "correlationId", "sourceService",
	} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	result, err := h.auditClient.List(c.Request().Context(), tenantID, correlationID, query)
	if err != nil {
		h.logger.Error("Erro ao listar auditoria",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *AuditHandlers) GetAuditHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.auditClient == nil {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Consulta de auditoria indisponível",
			CorrelationID: correlationID,
		})
	}
	tenantID := middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	if tenantID == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "tenant_id do token JWT é obrigatório",
			CorrelationID: correlationID,
		})
	}
	eventID := c.Param("eventId")
	result, err := h.auditClient.GetByID(c.Request().Context(), tenantID, correlationID, eventID)
	if err != nil {
		h.logger.Error("Erro ao buscar evento de auditoria",
			zap.String("correlationId", correlationID),
			zap.String("eventId", eventID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}
