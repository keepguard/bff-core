package handlers

import (
	"net/http"
	"time"

	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// UserHandlers handlers autenticados do perfil do usuário logado.
type UserHandlers struct {
	userClient client.UserClient
	logger     *zap.Logger
}

// NewUserHandlers cria UserHandlers.
func NewUserHandlers(userClient client.UserClient, logger *zap.Logger) *UserHandlers {
	return &UserHandlers{
		userClient: userClient,
		logger:     logger,
	}
}

// GetMeHandler retorna o perfil do usuário autenticado (sub do JWT).
func (h *UserHandlers) GetMeHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	tenantId := middlewarePkg.GetTenantId(c)
	token := middlewarePkg.GetTokenFromContext(c)
	codeUser := middlewarePkg.GetUserIDFromContext(c)

	if codeUser == "" {
		h.logger.Warn("JWT sem sub/codeUser", zap.String("correlationId", correlationID))
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Token JWT sem identificador de usuário",
			TraceID: correlationID,
		})
	}

	ctx := client.WithCompanyID(c.Request().Context(), middlewarePkg.GetCompanyIDFromContext(c))
	user, err := h.userClient.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
	if err != nil {
		h.logger.Error("Erro ao buscar perfil do usuário",
			zap.String("correlationId", correlationID),
			zap.String("codeUser", codeUser),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, toMeProfile(user))
}

func toMeProfile(user userDto.MSUserResponseDTO) inboundDto.MeProfileResponseDTO {
	resp := inboundDto.MeProfileResponseDTO{
		Email:           user.Email,
		PhoneE164:       user.PhoneE164,
		PreferredLocale: user.PreferredLocale,
		Timezone:        user.Timezone,
		AvatarURL:       user.AvatarURL,
		DisplayHandle:   user.DisplayHandle,
		Type:            user.Type,
		Status:          user.Status,
	}
	if !user.CreatedAt.IsZero() {
		resp.CreatedAt = user.CreatedAt.Format(time.RFC3339)
	}
	if user.PersonProfile != nil && user.PersonProfile.FullName != "" {
		resp.PersonProfile = &inboundDto.MePersonProfileDTO{FullName: user.PersonProfile.FullName}
	}
	return resp
}
