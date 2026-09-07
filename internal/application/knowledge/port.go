package knowledge

import (
	"context"
	"errors"
	"strings"
	"time"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/application/scope"
	"github.com/keepguard/bff-core/internal/pkg"
	"go.uber.org/zap"
)

var errKnowledgeServiceTokenMissing = errors.New("token de serviço knowledge indisponível")

type AskCommand struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	Question       string
	Context        string
}

type KnowledgePort interface {
	Ask(ctx context.Context, cmd AskCommand) (appdto.KnowledgeAskResponse, error)
}

type service struct {
	knowledge port.KnowledgeClient
	collector port.CollectorClient
	companies port.CompanyClient
	tokens    port.ServiceTokenClient
	logger    *zap.Logger
}

func NewKnowledgePort(
	knowledge port.KnowledgeClient,
	collector port.CollectorClient,
	companies port.CompanyClient,
	tokens port.ServiceTokenClient,
	logger *zap.Logger,
) KnowledgePort {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &service{
		knowledge: knowledge,
		collector: collector,
		companies: companies,
		tokens:    tokens,
		logger:    logger,
	}
}

func (s *service) Ask(ctx context.Context, cmd AskCommand) (appdto.KnowledgeAskResponse, error) {
	if s.knowledge == nil || s.companies == nil {
		return appdto.KnowledgeAskResponse{}, scope.Unavailable("Serviço de conhecimento indisponível")
	}
	companyID, err := scope.ResolveCompany(ctx, s.companies, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.KnowledgeAskResponse{}, err
	}
	question := strings.TrimSpace(cmd.Question)
	if question == "" {
		return appdto.KnowledgeAskResponse{}, pkg.NewAppError("BAD_REQUEST", "question é obrigatório", 400)
	}
	contextFilter := strings.TrimSpace(cmd.Context)
	hints := s.sourceHints(ctx, companyID, cmd.CorrelationID, contextFilter)
	bearer, tokErr := knowledgeServiceBearer(ctx, s.tokens, companyID)
	if tokErr != nil {
		s.logger.Error("Erro ao obter token OAuth do BFF para knowledge",
			zap.String("correlationId", cmd.CorrelationID),
			zap.String("companyId", companyID),
			zap.Error(tokErr),
		)
		return appdto.KnowledgeAskResponse{}, scope.Unavailable("Serviço de conhecimento indisponível")
	}
	result, askErr := s.knowledge.Ask(ctx, companyID, bearer, cmd.CorrelationID, appdto.KnowledgeAskRequest{
		Question:    question,
		Context:     contextFilter,
		SourceHints: hints,
	})
	if askErr != nil {
		return appdto.KnowledgeAskResponse{}, askErr
	}
	result.Freshness = s.freshness(ctx, companyID, cmd.CorrelationID, result, hints)
	return result, nil
}

func (s *service) sourceHints(ctx context.Context, companyID, correlationID, contextFilter string) []appdto.KnowledgeSourceHint {
	if s.collector == nil {
		return []appdto.KnowledgeSourceHint{}
	}
	raw, err := s.collector.SearchAgents(ctx, companyID, correlationID, map[string]string{
		"page": "0",
		"size": "100",
	})
	if err != nil {
		s.logger.Warn("Não foi possível carregar agents para sourceHints",
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

func (s *service) freshness(
	ctx context.Context,
	companyID, correlationID string,
	result appdto.KnowledgeAskResponse,
	hints []appdto.KnowledgeSourceHint,
) *appdto.KnowledgeFreshness {
	if s.collector == nil {
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
		execs, err := s.collector.ListAgentExecutions(ctx, companyID, agentID, correlationID, 1)
		if err != nil {
			s.logger.Warn("Não foi possível carregar última execução do collector",
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

func knowledgeServiceBearer(ctx context.Context, tokens port.ServiceTokenClient, companyID string) (string, error) {
	if tokens == nil {
		return "", errKnowledgeServiceTokenMissing
	}
	return tokens.GetToken(ctx, companyID)
}
