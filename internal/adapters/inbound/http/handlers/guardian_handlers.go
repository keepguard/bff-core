package handlers

import (
	"net/http"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type GuardianHandlers struct {
	guardianClient client.GuardianClient
	logger         *zap.Logger
}

func NewGuardianHandlers(guardianClient client.GuardianClient, logger *zap.Logger) *GuardianHandlers {
	return &GuardianHandlers{guardianClient: guardianClient, logger: logger}
}

func (h *GuardianHandlers) ListGuardianIncidentsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardianClient == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	query := map[string]string{}
	for _, key := range []string{
		"page", "size", "from", "to", "status", "severity", "serviceName", "namespace",
		"k8sConclusion", "errorReason", "correlationId", "q", "sort", "dir",
	} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	result, err := h.guardianClient.ListIncidents(c.Request().Context(), tenantID, correlationID, query)
	if err != nil {
		h.logger.Error("Erro ao listar incidentes do Guardian", zap.String("correlationId", correlationID), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) GetGuardianIncidentHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardianClient == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	result, err := h.guardianClient.GetIncident(c.Request().Context(), tenantID, correlationID, c.Param("id"))
	if err != nil {
		h.logger.Error("Erro ao buscar incidente do Guardian", zap.String("correlationId", correlationID), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) ExecuteGuardianActionHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardianClient == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	var body appdto.GuardianExecuteActionRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error: "BAD_REQUEST", Message: "Payload inválido", CorrelationID: correlationID,
		})
	}
	claims := middlewarePkg.GetClaimsFromContext(c)
	userID := middlewarePkg.GetUserIDFromContext(c)
	email := ""
	role := ""
	if claims != nil {
		email = claims.Email
		if len(claims.Roles) > 0 {
			role = claims.Roles[0]
		}
	}
	result, err := h.guardianClient.ExecuteAction(c.Request().Context(), tenantID, correlationID, userID, email, role, c.Param("id"), body)
	if err != nil {
		h.logger.Error("Erro ao executar ação do Guardian", zap.String("correlationId", correlationID), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) ListGuardianRecipientsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardianClient == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	result, err := h.guardianClient.ListRecipients(c.Request().Context(), tenantID, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) UpsertGuardianRecipientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardianClient == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	var body appdto.GuardianRecipientUpsertRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error: "BAD_REQUEST", Message: "Payload inválido", CorrelationID: correlationID,
		})
	}
	result, err := h.guardianClient.UpsertRecipient(c.Request().Context(), tenantID, correlationID, body)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) PatchGuardianRecipientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardianClient == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	var body appdto.GuardianRecipientUpsertRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error: "BAD_REQUEST", Message: "Payload inválido", CorrelationID: correlationID,
		})
	}
	result, err := h.guardianClient.PatchRecipient(c.Request().Context(), tenantID, correlationID, c.Param("id"), body)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func requireTenant(c echo.Context, correlationID string) (string, error) {
	tenantID := middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	if tenantID == "" {
		return "", c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error: "MISSING_TENANT", Message: "tenant_id do token JWT é obrigatório", CorrelationID: correlationID,
		})
	}
	return tenantID, nil
}

func unavailable(c echo.Context, correlationID string) error {
	return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
		Error: "SERVICE_UNAVAILABLE", Message: "Guardian indisponível", CorrelationID: correlationID,
	})
}
