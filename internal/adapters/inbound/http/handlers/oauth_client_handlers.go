package handlers

import (
	"net/http"

	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-core/internal/adapters/inbound/http/mapper"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/oauth"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type OAuthClientHandlers struct {
	oauth  oauth.OAuthPort
	logger *zap.Logger
}

func NewOAuthClientHandlers(oauthPort oauth.OAuthPort, logger *zap.Logger) *OAuthClientHandlers {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &OAuthClientHandlers{oauth: oauthPort, logger: logger}
}

func (h *OAuthClientHandlers) scope(c echo.Context) oauth.CompanyScope {
	return oauth.CompanyScope{
		CompanyFromCtx: port.CompanyIDFromContext(c.Request().Context()),
		TenantID:       middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c)),
		CorrelationID:  middlewarePkg.GetCorrelationID(c),
		BearerToken:    middlewarePkg.GetTokenFromContext(c),
	}
}

func (h *OAuthClientHandlers) ListOAuthClientsHandler(c echo.Context) error {
	scope := h.scope(c)
	query := map[string]string{}
	for _, key := range []string{"clientId", "status", "page", "size", "sort", "dir"} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	result, err := h.oauth.Search(c.Request().Context(), oauth.ListClientsQuery{CompanyScope: scope, Filters: query})
	if err != nil {
		h.logger.Error("Erro ao listar OAuth clients",
			zap.String("correlationId", scope.CorrelationID),
			zap.Error(err),
		)
		return handleError(c, err, scope.CorrelationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *OAuthClientHandlers) GetOAuthClientHandler(c echo.Context) error {
	scope := h.scope(c)
	id := c.Param("id")
	item, err := h.oauth.GetByID(c.Request().Context(), oauth.GetClientQuery{CompanyScope: scope, ID: id})
	if err != nil {
		h.logger.Error("Erro ao obter OAuth client",
			zap.String("correlationId", scope.CorrelationID),
			zap.String("id", id),
			zap.Error(err),
		)
		return handleError(c, err, scope.CorrelationID)
	}
	return c.JSON(http.StatusOK, item)
}

func (h *OAuthClientHandlers) CreateOAuthClientHandler(c echo.Context) error {
	scope := h.scope(c)
	var body inboundDto.OAuthClientCreateRequestDTO
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_BODY",
			Message:       "JSON inválido",
			CorrelationID: scope.CorrelationID,
		})
	}
	created, err := h.oauth.Create(c.Request().Context(), mapper.ToOAuthCreateCommand(
		body, scope.CompanyFromCtx, scope.TenantID, scope.CorrelationID, scope.BearerToken,
	))
	if err != nil {
		h.logger.Error("Erro ao criar OAuth client",
			zap.String("correlationId", scope.CorrelationID),
			zap.Error(err),
		)
		return handleError(c, err, scope.CorrelationID)
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *OAuthClientHandlers) UpdateOAuthClientHandler(c echo.Context) error {
	scope := h.scope(c)
	var body inboundDto.OAuthClientUpdateRequestDTO
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_BODY",
			Message:       "JSON inválido",
			CorrelationID: scope.CorrelationID,
		})
	}
	id := c.Param("id")
	updated, err := h.oauth.Update(c.Request().Context(), mapper.ToOAuthUpdateCommand(
		body, scope.CompanyFromCtx, scope.TenantID, scope.CorrelationID, scope.BearerToken, id,
	))
	if err != nil {
		h.logger.Error("Erro ao atualizar OAuth client",
			zap.String("correlationId", scope.CorrelationID),
			zap.String("id", id),
			zap.Error(err),
		)
		return handleError(c, err, scope.CorrelationID)
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *OAuthClientHandlers) ListOAuthServiceRolesHandler(c echo.Context) error {
	scope := h.scope(c)
	roles, err := h.oauth.ListServiceRoles(c.Request().Context(), scope)
	if err != nil {
		h.logger.Error("Erro ao listar service roles",
			zap.String("correlationId", scope.CorrelationID),
			zap.Error(err),
		)
		return handleError(c, err, scope.CorrelationID)
	}
	return c.JSON(http.StatusOK, roles)
}

func (h *OAuthClientHandlers) BlockOAuthClientHandler(c echo.Context) error {
	return h.mutate(c, true)
}

func (h *OAuthClientHandlers) UnblockOAuthClientHandler(c echo.Context) error {
	return h.mutate(c, false)
}

func (h *OAuthClientHandlers) mutate(c echo.Context, block bool) error {
	scope := h.scope(c)
	id := c.Param("id")
	cmd := oauth.ClientIDCommand{CompanyScope: scope, ID: id}
	var result interface{}
	var err error
	if block {
		result, err = h.oauth.Block(c.Request().Context(), cmd)
	} else {
		result, err = h.oauth.Unblock(c.Request().Context(), cmd)
	}
	if err != nil {
		h.logger.Error("Erro ao alterar status do OAuth client",
			zap.String("correlationId", scope.CorrelationID),
			zap.String("id", id),
			zap.Bool("block", block),
			zap.Error(err),
		)
		return handleError(c, err, scope.CorrelationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *OAuthClientHandlers) DeleteOAuthClientHandler(c echo.Context) error {
	scope := h.scope(c)
	id := c.Param("id")
	if err := h.oauth.Delete(c.Request().Context(), oauth.ClientIDCommand{CompanyScope: scope, ID: id}); err != nil {
		h.logger.Error("Erro ao excluir OAuth client",
			zap.String("correlationId", scope.CorrelationID),
			zap.String("id", id),
			zap.Error(err),
		)
		return handleError(c, err, scope.CorrelationID)
	}
	return c.NoContent(http.StatusNoContent)
}
