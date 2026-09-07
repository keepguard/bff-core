package handlers

import (
	"net/http"

	"github.com/keepguard/bff-core/internal/adapters/inbound/http/mapper"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	appuser "github.com/keepguard/bff-core/internal/application/user"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type UserHandlers struct {
	users  appuser.UserPort
	logger *zap.Logger
}

func NewUserHandlers(users appuser.UserPort, logger *zap.Logger) *UserHandlers {
	return &UserHandlers{users: users, logger: logger}
}

func (h *UserHandlers) GetMeHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	tenantId := middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	token := middlewarePkg.GetTokenFromContext(c)
	codeUser := middlewarePkg.GetUserIDFromContext(c)

	if tenantId == "" {
		h.logger.Warn("JWT sem tenant_id", zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "tenant_id do token JWT é obrigatório",
			CorrelationID: correlationID,
		})
	}

	if codeUser == "" {
		h.logger.Warn("JWT sem sub/codeUser", zap.String("correlationId", correlationID))
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "Token JWT sem identificador de usuário",
			CorrelationID: correlationID,
		})
	}

	view, err := h.users.GetMe(c.Request().Context(), appdto.GetMeQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		CodeUser:      codeUser,
	})
	if err != nil {
		h.logger.Error("Erro ao buscar perfil do usuário",
			zap.String("correlationId", correlationID),
			zap.String("codeUser", codeUser),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, mapper.ToMeProfileResponse(view))
}
