package handlers

import (
	"net/http"
	"strings"

	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-core/internal/adapters/inbound/http/mapper"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/knowledge"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type KnowledgeHandlers struct {
	knowledge knowledge.KnowledgePort
	logger    *zap.Logger
}

func NewKnowledgeHandlers(knowledgePort knowledge.KnowledgePort, logger *zap.Logger) *KnowledgeHandlers {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &KnowledgeHandlers{knowledge: knowledgePort, logger: logger}
}

func (h *KnowledgeHandlers) AskKnowledgeHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.knowledge == nil {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Serviço de conhecimento indisponível",
			CorrelationID: correlationID,
		})
	}
	var body inboundDto.KnowledgeAskRequestDTO
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
	result, err := h.knowledge.Ask(c.Request().Context(), mapper.ToAskKnowledgeCommand(
		body,
		port.CompanyIDFromContext(c.Request().Context()),
		middlewarePkg.ResolveTenantId(c, middlewarePkg.GetClaimsFromContext(c)),
		correlationID,
	))
	if err != nil {
		h.logger.Error("Erro ao perguntar ao knowledge",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, result)
}
