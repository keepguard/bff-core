package handlers

import (
	"errors"
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
	ClientID        string `json:"clientId"`
	Description     string `json:"description,omitempty"`
	RoleID          string `json:"roleId"`
	TokenTTLSeconds *int   `json:"tokenTtlSeconds,omitempty"`
}

func (h *OAuthClientHandlers) ListOAuthClientsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
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
			zap.String("companyId", companyID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *OAuthClientHandlers) GetOAuthClientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
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
	agents, agentsLoadErr := h.loadAgents(c, companyID, correlationID)
	resp := appdto.OAuthClientDetailResponse{
		OAuthClientDTO: item,
		Agents:         agents,
	}
	if agentsLoadErr != nil {
		resp.AgentsLoadError = agentsLoadErr.Error()
	}
	return c.JSON(http.StatusOK, resp)
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
	companyID, err := h.resolveCompany(c, correlationID)
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
	if strings.TrimSpace(body.RoleID) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_ROLE_ID",
			Message:       "roleId é obrigatório",
			CorrelationID: correlationID,
		})
	}
	token := middlewarePkg.GetTokenFromContext(c)
	created, err := h.oauthClient.Create(c.Request().Context(), companyID, token, correlationID, appdto.OAuthClientCreateRequest{
		ClientID:        body.ClientID,
		Description:     body.Description,
		RoleID:          body.RoleID,
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

func (h *OAuthClientHandlers) ListOAuthServiceRolesHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	token := middlewarePkg.GetTokenFromContext(c)
	roles, err := h.oauthClient.ListServiceRoles(c.Request().Context(), companyID, token, correlationID)
	if err != nil {
		h.logger.Error("Erro ao listar service roles",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	if roles == nil {
		roles = []appdto.OAuthServiceRoleDTO{}
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
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
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
	if block {
		h.disableCompanyAgents(c, companyID, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *OAuthClientHandlers) DeleteOAuthClientHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	companyID, err := h.resolveCompany(c, correlationID)
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
	h.disableCompanyAgents(c, companyID, correlationID)
	return c.NoContent(http.StatusNoContent)
}

func (h *OAuthClientHandlers) resolveCompany(c echo.Context, correlationID string) (string, error) {
	if h.oauthClient == nil || h.companyClient == nil {
		return "", c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Gestão de OAuth clients indisponível",
			CorrelationID: correlationID,
		})
	}
	if companyID := client.CompanyIDFromContext(c.Request().Context()); companyID != "" {
		return companyID, nil
	}
	tenantID := middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c))
	if strings.TrimSpace(tenantID) == "" {
		return "", c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "tenantId do JWT é obrigatório",
			CorrelationID: correlationID,
		})
	}
	company, err := h.companyClient.GetByTenantId(c.Request().Context(), tenantID, correlationID)
	if err != nil {
		h.logger.Error("Erro ao resolver company pelo tenant do JWT",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantID),
			zap.Error(err),
		)
		return "", handleError(c, err, correlationID)
	}
	if company.ID == "" {
		return "", c.JSON(http.StatusNotFound, pkg.ErrorResponse{
			Error:         "COMPANY_NOT_FOUND",
			Message:       "Empresa não encontrada para o tenant autenticado",
			CorrelationID: correlationID,
		})
	}
	return company.ID, nil
}

const collectorUnavailableMsg = "srv-data-collector indisponível"

func (h *OAuthClientHandlers) loadAgents(c echo.Context, companyID, correlationID string) ([]appdto.CollectorAgentDTO, error) {
	if h.collectorClient == nil {
		return []appdto.CollectorAgentDTO{}, nil
	}
	raw, err := h.collectorClient.ListAgents(c.Request().Context(), companyID, correlationID)
	if err != nil {
		h.logger.Warn("Collector indisponível ao carregar agents",
			zap.String("correlationId", correlationID),
			zap.String("companyId", companyID),
			zap.Error(err),
		)
		return []appdto.CollectorAgentDTO{}, errors.New(collectorUnavailableMsg)
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
	return out, nil
}

func (h *OAuthClientHandlers) disableCompanyAgents(c echo.Context, companyID, correlationID string) {
	if h.collectorClient == nil {
		return
	}
	agents, err := h.loadAgents(c, companyID, correlationID)
	if err != nil {
		h.logger.Warn("Collector indisponível ao desabilitar agents",
			zap.String("correlationId", correlationID),
			zap.String("companyId", companyID),
			zap.Error(err),
		)
		return
	}
	for _, agent := range agents {
		if !agent.Enabled || strings.TrimSpace(agent.ID) == "" {
			continue
		}
		if err := h.collectorClient.DisableAgent(c.Request().Context(), companyID, agent.ID, correlationID); err != nil {
			h.logger.Warn("Falha ao desabilitar agent após alteração do OAuth client",
				zap.String("correlationId", correlationID),
				zap.String("companyId", companyID),
				zap.String("agentId", agent.ID),
				zap.Error(err),
			)
		}
	}
}
