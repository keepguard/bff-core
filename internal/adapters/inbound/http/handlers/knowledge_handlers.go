package handlers

import (
	"net/http"
	"strings"
	"time"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type KnowledgeHandlers struct {
	knowledgeClient client.KnowledgeClient
	collectorClient client.CollectorClient
	companyClient   client.CompanyClient
	logger          *zap.Logger
}

func NewKnowledgeHandlers(
	knowledgeClient client.KnowledgeClient,
	collectorClient client.CollectorClient,
	companyClient client.CompanyClient,
	logger *zap.Logger,
) *KnowledgeHandlers {
	return &KnowledgeHandlers{
		knowledgeClient: knowledgeClient,
		collectorClient: collectorClient,
		companyClient:   companyClient,
		logger:          logger,
	}
}

type knowledgeAskBody struct {
	Question string `json:"question"`
	Context  string `json:"context,omitempty"`
}

func (h *KnowledgeHandlers) AskKnowledgeHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.knowledgeClient == nil || h.companyClient == nil {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Knowledge indisponível",
			CorrelationID: correlationID,
		})
	}
	companyID, err := h.resolveCompany(c, correlationID)
	if err != nil {
		return err
	}
	var body knowledgeAskBody
	if bindErr := c.Bind(&body); bindErr != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "BAD_REQUEST",
			Message:       "Payload inválido",
			CorrelationID: correlationID,
		})
	}
	if strings.TrimSpace(body.Question) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "BAD_REQUEST",
			Message:       "question é obrigatório",
			CorrelationID: correlationID,
		})
	}
	hints := h.sourceHints(c, companyID, correlationID, strings.TrimSpace(body.Context))
	result, askErr := h.knowledgeClient.Ask(
		c.Request().Context(),
		companyID,
		bearerFrom(c),
		correlationID,
		appdto.KnowledgeAskRequest{
			Question:    strings.TrimSpace(body.Question),
			Context:     strings.TrimSpace(body.Context),
			SourceHints: hints,
		},
	)
	if askErr != nil {
		h.logger.Error("Erro ao perguntar ao knowledge",
			zap.String("correlationId", correlationID),
			zap.String("companyId", companyID),
			zap.Error(askErr),
		)
		return handleError(c, askErr, correlationID)
	}
	result.Freshness = h.freshness(c, companyID, correlationID, result, hints)
	return c.JSON(http.StatusOK, result)
}

func (h *KnowledgeHandlers) sourceHints(c echo.Context, companyID, correlationID, contextFilter string) []appdto.KnowledgeSourceHint {
	if h.collectorClient == nil {
		return []appdto.KnowledgeSourceHint{}
	}
	raw, err := h.collectorClient.SearchAgents(c.Request().Context(), companyID, correlationID, map[string]string{
		"page": "0",
		"size": "100",
	})
	if err != nil {
		h.logger.Warn("Não foi possível carregar agents para sourceHints",
			zap.String("correlationId", correlationID),
			zap.String("companyId", companyID),
			zap.Error(err),
		)
		return []appdto.KnowledgeSourceHint{}
	}
	hints := make([]appdto.KnowledgeSourceHint, 0, len(raw.Content))
	for _, agent := range raw.Content {
		if contextFilter != "" && strings.TrimSpace(agent.Context) != "" &&
			!strings.EqualFold(strings.TrimSpace(agent.Context), contextFilter) {
			continue
		}
		hints = append(hints, appdto.KnowledgeSourceHint{
			AgentID: agent.ID,
			Name:    agent.Name,
			Context: agent.Context,
			Prompt:  agent.Prompt,
		})
	}
	return hints
}

func (h *KnowledgeHandlers) resolveCompany(c echo.Context, correlationID string) (string, error) {
	if h.companyClient == nil {
		return "", c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Knowledge indisponível",
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

func (h *KnowledgeHandlers) freshness(
	c echo.Context,
	companyID, correlationID string,
	result appdto.KnowledgeAskResponse,
	hints []appdto.KnowledgeSourceHint,
) *appdto.KnowledgeFreshness {
	if h.collectorClient == nil {
		return nil
	}
	agentIDs := sourceAgentIDs(result.Sources, hints)
	if len(agentIDs) == 0 {
		return nil
	}
	names := agentNameByID(result.Sources, hints)
	var picked *appdto.KnowledgeFreshness
	lookedUp := false
	for _, agentID := range agentIDs {
		execs, err := h.collectorClient.ListAgentExecutions(
			c.Request().Context(), companyID, agentID, correlationID, 1,
		)
		if err != nil {
			h.logger.Warn("Não foi possível carregar última execução do collector",
				zap.String("correlationId", correlationID),
				zap.String("companyId", companyID),
				zap.String("agentId", agentID),
				zap.Error(err),
			)
			continue
		}
		if len(execs) == 0 {
			continue
		}
		lookedUp = true
		candidate := freshnessFromExecution(execs[0], agentID, names[agentID])
		if candidate == nil {
			continue
		}
		if preferFreshness(candidate, picked) {
			picked = candidate
		}
	}
	if !lookedUp {
		return nil
	}
	return picked
}

func sourceAgentIDs(sources []appdto.KnowledgeAskSource, hints []appdto.KnowledgeSourceHint) []string {
	ids := make([]string, 0, 5)
	seen := map[string]struct{}{}
	appendID := func(raw string) {
		id := strings.TrimSpace(raw)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		if len(ids) >= 5 {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	for _, source := range sources {
		appendID(source.SourceAgentID)
	}
	if len(ids) == 0 {
		for _, hint := range hints {
			appendID(hint.AgentID)
		}
	}
	return ids
}

func agentNameByID(sources []appdto.KnowledgeAskSource, hints []appdto.KnowledgeSourceHint) map[string]string {
	names := map[string]string{}
	for _, hint := range hints {
		id := strings.TrimSpace(hint.AgentID)
		if id == "" {
			continue
		}
		names[id] = strings.TrimSpace(hint.Name)
	}
	for _, source := range sources {
		id := strings.TrimSpace(source.SourceAgentID)
		if id == "" {
			continue
		}
		if name := strings.TrimSpace(source.AgentName); name != "" {
			names[id] = name
		}
	}
	return names
}

func freshnessFromExecution(exec appdto.CollectorExecutionRaw, agentID, agentName string) *appdto.KnowledgeFreshness {
	at := strings.TrimSpace(exec.StartedAt)
	if exec.FinishedAt != nil && strings.TrimSpace(*exec.FinishedAt) != "" {
		at = strings.TrimSpace(*exec.FinishedAt)
	}
	status := strings.ToUpper(strings.TrimSpace(exec.Status))
	return &appdto.KnowledgeFreshness{
		LastCollectionAt: at,
		AgeMinutes:       ageMinutesFrom(at),
		Status:           status,
		Failed:           status == "FAILED" || status == "PARTIAL",
		ErrorMessage:     exec.ErrorMessage,
		AgentID:          agentID,
		AgentName:        agentName,
	}
}

func preferFreshness(candidate, current *appdto.KnowledgeFreshness) bool {
	if candidate == nil {
		return false
	}
	if current == nil {
		return true
	}
	if candidate.Failed != current.Failed {
		return candidate.Failed
	}
	candidateAt, candidateErr := parseFreshnessTime(candidate.LastCollectionAt)
	currentAt, currentErr := parseFreshnessTime(current.LastCollectionAt)
	if candidateErr != nil || currentErr != nil {
		return candidate.AgeMinutes < current.AgeMinutes
	}
	return candidateAt.After(currentAt)
}

func ageMinutesFrom(iso string) int {
	parsed, err := parseFreshnessTime(iso)
	if err != nil {
		return 0
	}
	minutes := int(time.Since(parsed).Minutes())
	if minutes < 0 {
		return 0
	}
	return minutes
}

func parseFreshnessTime(iso string) (time.Time, error) {
	raw := strings.TrimSpace(iso)
	if parsed, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, raw)
}

func bearerFrom(c echo.Context) string {
	header := strings.TrimSpace(c.Request().Header.Get(echo.HeaderAuthorization))
	if header != "" {
		return header
	}
	if token, ok := c.Get("token").(string); ok && strings.TrimSpace(token) != "" {
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			return token
		}
		return "Bearer " + token
	}
	return ""
}
