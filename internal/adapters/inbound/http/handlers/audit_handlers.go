package handlers

import (
	"net/http"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/audit"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AuditHandlers struct {
	audits audit.AuditPort
	logger *zap.Logger
}

func NewAuditHandlers(audits audit.AuditPort, logger *zap.Logger) *AuditHandlers {
	return &AuditHandlers{audits: audits, logger: logger}
}

// ListAuditsHandler lista eventos de auditoria.
// Path canônico: GET /api/v1/core/audits.
// GET /api/v1/audits é legado (mesmo handler) e está depreciado; remoção exige front+Traefik.
// @Summary Listar auditoria
// @Description Lista eventos. Prefira GET /api/v1/core/audits. GET /api/v1/audits está depreciado (mesmo handler).
// @Tags audits
// @Produce json
// @Success 200 "Lista paginada de eventos"
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 503 {object} pkg.ErrorResponse
// @Router /api/v1/core/audits [get]
func (h *AuditHandlers) ListAuditsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.audits == nil {
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
		"resourceType", "resourceId", "correlationId", "sourceService", "sort", "dir",
	} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	result, err := h.audits.List(c.Request().Context(), audit.ListAuditsQuery{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Filters:       query,
	})
	if err != nil {
		h.logger.Error("Erro ao listar auditoria",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

// GetAuditHandler obtém um evento de auditoria.
// Path canônico: GET /api/v1/core/audits/{eventId}.
// GET /api/v1/audits/{eventId} é legado (mesmo handler) e está depreciado.
// @Summary Obter evento de auditoria
// @Description Obtém um evento. Prefira GET /api/v1/core/audits/{eventId}. GET /api/v1/audits/{eventId} está depreciado (mesmo handler).
// @Tags audits
// @Produce json
// @Param eventId path string true "ID do evento"
// @Success 200 "Evento de auditoria"
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 503 {object} pkg.ErrorResponse
// @Router /api/v1/core/audits/{eventId} [get]
func (h *AuditHandlers) GetAuditHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.audits == nil {
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
	result, err := h.audits.Get(c.Request().Context(), audit.GetAuditQuery{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		EventID:       eventID,
	})
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
