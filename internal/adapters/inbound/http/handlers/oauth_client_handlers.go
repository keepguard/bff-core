package handlers

import (
	"net/http"
	"strings"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type OAuthClientHandlers struct {
	oauthClient     client.OAuthClientClient
	companyClient   client.CompanyClient
	collectorClient client.CollectorClient
	logger          *zap.Logger
}

func NewOAuthClientHandlers(
	oauthClient client.OAuthClientClient,
	companyClient client.CompanyClient,
	collectorClient client.CollectorClient,
	logger *zap.Logger,
) *OAuthClientHandlers {
	return &OAuthClientHandlers{
		oauthClient:     oauthClient,
		companyClient:   companyClient,
		collectorClient: collectorClient,
		logger:          logger,
	}
}

type oauthClientCreateBody struct {
	TenantID        string   `json:"tenantId"`
	ClientID        string   `json:"clientId"`
	Description     string   `json:"description,omitempty"`
	Authorities     []string `json:"authorities,omitempty"`
	TokenTTLSeconds *int     `json:"tokenTtlSeconds,omitempty"`
}

func (h *OAuthClientHandlers) ListOAuthClientsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	tenantID := strings.TrimSpace(c.QueryParam("tenantId"))
	companyID, err := h.resolveCompany(c, tenantID, correlationID)
	if err != nil {
		return err
	}
	token := middlewarePkg.GetTokenFromContext(c)
	query := map[string]string{}
	for _, key := range []string{"clientId", "status", "page", "size", "sort", "dir"} {
		if value := c.QueryParam(key); value != "" {
			query[key] = value
		}
	}
	result, err := h.oauthClient.Search(c.Request().Context(), companyID, token, correlationID, query)
	if err != nil {
		h.logger.Error("Erro ao listar OAuth clients",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *OAuthClientHandlers) GetOAuthClientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	tenantID := strings.TrimSpace(c.QueryParam("tenantId"))
	companyID, err := h.resolveCompany(c, tenantID, correlationID)
	if err != nil {
		return err
	}
	token := middlewarePkg.GetTokenFromContext(c)
	id := c.Param("id")
	item, err := h.oauthClient.GetByID(c.Request().Context(), companyID, token, correlationID, id)
	if err != nil {
		h.logger.Error("Erro ao obter OAuth client",
			zap.String("correlationId", correlationID),
			zap.String("id", id),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	agents := h.loadAgents(c, companyID, correlationID)
	return c.JSON(http.StatusOK, appdto.OAuthClientDetailResponse{
		OAuthClientDTO: item,
		TenantID:       tenantID,
		Agents:         agents,
	})
}

func (h *OAuthClientHandlers) CreateOAuthClientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	var body oauthClientCreateBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_BODY",
			Message:       "JSON inválido",
			CorrelationID: correlationID,
		})
	}
	tenantID := strings.TrimSpace(body.TenantID)
	if tenantID == "" {
		tenantID = strings.TrimSpace(c.QueryParam("tenantId"))
	}
	companyID, err := h.resolveCompany(c, tenantID, correlationID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(body.ClientID) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_CLIENT_ID",
			Message:       "clientId é obrigatório",
			CorrelationID: correlationID,
		})
	}
	token := middlewarePkg.GetTokenFromContext(c)
	created, err := h.oauthClient.Create(c.Request().Context(), companyID, token, correlationID, appdto.OAuthClientCreateRequest{
		ClientID:        body.ClientID,
		Description:     body.Description,
		Authorities:     body.Authorities,
		TokenTTLSeconds: body.TokenTTLSeconds,
	})
	if err != nil {
		h.logger.Error("Erro ao criar OAuth client",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *OAuthClientHandlers) BlockOAuthClientHandler(c echo.Context) error {
	return h.mutate(c, true)
}

func (h *OAuthClientHandlers) UnblockOAuthClientHandler(c echo.Context) error {
	return h.mutate(c, false)
}

func (h *OAuthClientHandlers) mutate(c echo.Context, block bool) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	tenantID := strings.TrimSpace(c.QueryParam("tenantId"))
	companyID, err := h.resolveCompany(c, tenantID, correlationID)
	if err != nil {
		return err
	}
	token := middlewarePkg.GetTokenFromContext(c)
	id := c.Param("id")
	var result appdto.OAuthClientDTO
	if block {
		result, err = h.oauthClient.Block(c.Request().Context(), companyID, token, correlationID, id)
	} else {
		result, err = h.oauthClient.Unblock(c.Request().Context(), companyID, token, correlationID, id)
	}
	if err != nil {
		h.logger.Error("Erro ao alterar status do OAuth client",
			zap.String("correlationId", correlationID),
			zap.String("id", id),
			zap.Bool("block", block),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *OAuthClientHandlers) DeleteOAuthClientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	tenantID := strings.TrimSpace(c.QueryParam("tenantId"))
	companyID, err := h.resolveCompany(c, tenantID, correlationID)
	if err != nil {
		return err
	}
	token := middlewarePkg.GetTokenFromContext(c)
	id := c.Param("id")
	if err := h.oauthClient.Delete(c.Request().Context(), companyID, token, correlationID, id); err != nil {
		h.logger.Error("Erro ao excluir OAuth client",
			zap.String("correlationId", correlationID),
			zap.String("id", id),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *OAuthClientHandlers) resolveCompany(c echo.Context, tenantID, correlationID string) (string, error) {
	if h.oauthClient == nil || h.companyClient == nil {
		return "", c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Gestão de OAuth clients indisponível",
			CorrelationID: correlationID,
		})
	}
	if strings.TrimSpace(tenantID) == "" {
		return "", c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "tenantId é obrigatório",
			CorrelationID: correlationID,
		})
	}
	company, err := h.companyClient.GetByTenantId(c.Request().Context(), tenantID, correlationID)
	if err != nil {
		h.logger.Error("Erro ao resolver company pelo tenant",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantID),
			zap.Error(err),
		)
		return "", handleError(c, err, correlationID)
	}
	if company.ID == "" {
		return "", c.JSON(http.StatusNotFound, pkg.ErrorResponse{
			Error:         "COMPANY_NOT_FOUND",
			Message:       "Empresa não encontrada para o tenant informado",
			CorrelationID: correlationID,
		})
	}
	return company.ID, nil
}

func (h *OAuthClientHandlers) loadAgents(c echo.Context, companyID, correlationID string) []appdto.CollectorAgentDTO {
	if h.collectorClient == nil {
		return []appdto.CollectorAgentDTO{}
	}
	raw, err := h.collectorClient.ListAgents(c.Request().Context(), companyID, correlationID)
	if err != nil {
		h.logger.Warn("Collector indisponível ao carregar agents",
			zap.String("correlationId", correlationID),
			zap.String("companyId", companyID),
			zap.Error(err),
		)
		return []appdto.CollectorAgentDTO{}
	}
	out := make([]appdto.CollectorAgentDTO, 0, len(raw))
	for _, item := range raw {
		out = append(out, appdto.CollectorAgentDTO{
			ID:            item.ID,
			Code:          item.Code,
			CompanyID:     item.CompanyID,
			Name:          item.Name,
			Description:   item.Description,
			CollectorType: item.CollectorType,
			Enabled:       item.Enabled,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}
	return out
}
