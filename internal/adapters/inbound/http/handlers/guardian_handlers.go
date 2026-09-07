package handlers

import (
	"net/http"

	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-core/internal/adapters/inbound/http/mapper"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/guardian"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type GuardianHandlers struct {
	guardian guardian.GuardianPort
	logger   *zap.Logger
}

func NewGuardianHandlers(guardianPort guardian.GuardianPort, logger *zap.Logger) *GuardianHandlers {
	return &GuardianHandlers{guardian: guardianPort, logger: logger}
}

func (h *GuardianHandlers) ListGuardianIncidentsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardian == nil {
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
	result, err := h.guardian.ListIncidents(c.Request().Context(), guardian.ListIncidentsQuery{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Filters:       query,
	})
	if err != nil {
		h.logger.Error("Erro ao listar incidentes do Guardian", zap.String("correlationId", correlationID), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) GetGuardianIncidentHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardian == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	result, err := h.guardian.GetIncident(c.Request().Context(), guardian.GetIncidentQuery{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		ID:            c.Param("id"),
	})
	if err != nil {
		h.logger.Error("Erro ao buscar incidente do Guardian", zap.String("correlationId", correlationID), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) ExecuteGuardianActionHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardian == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	var body inboundDto.GuardianExecuteActionRequestDTO
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
	result, err := h.guardian.ExecuteAction(c.Request().Context(), guardian.ExecuteActionCommand{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		UserID:        userID,
		UserEmail:     email,
		UserRole:      role,
		ID:            c.Param("id"),
		Body:          mapper.ToGuardianExecuteActionRequest(body),
	})
	if err != nil {
		h.logger.Error("Erro ao executar ação do Guardian", zap.String("correlationId", correlationID), zap.Error(err))
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) ListGuardianRecipientsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardian == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	result, err := h.guardian.ListRecipients(c.Request().Context(), guardian.ListRecipientsQuery{
		TenantID:      tenantID,
		CorrelationID: correlationID,
	})
	if err != nil {
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) UpsertGuardianRecipientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardian == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	var body inboundDto.GuardianRecipientUpsertRequestDTO
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error: "BAD_REQUEST", Message: "Payload inválido", CorrelationID: correlationID,
		})
	}
	result, err := h.guardian.UpsertRecipient(c.Request().Context(), guardian.UpsertRecipientCommand{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Body:          mapper.ToGuardianRecipientUpsertRequest(body),
	})
	if err != nil {
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *GuardianHandlers) PatchGuardianRecipientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.guardian == nil {
		return unavailable(c, correlationID)
	}
	tenantID, errResp := requireTenant(c, correlationID)
	if errResp != nil {
		return errResp
	}
	var body inboundDto.GuardianRecipientUpsertRequestDTO
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error: "BAD_REQUEST", Message: "Payload inválido", CorrelationID: correlationID,
		})
	}
	result, err := h.guardian.PatchRecipient(c.Request().Context(), guardian.PatchRecipientCommand{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		ID:            c.Param("id"),
		Body:          mapper.ToGuardianRecipientUpsertRequest(body),
	})
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
