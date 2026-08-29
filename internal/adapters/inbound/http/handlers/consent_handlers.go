package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	userConsentDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user_consent"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ConsentHandlers handlers autenticados de consentimento LGPD.
type ConsentHandlers struct {
	userConsentClient client.UserConsentClient
	logger            *zap.Logger
}

// NewConsentHandlers cria ConsentHandlers.
func NewConsentHandlers(userConsentClient client.UserConsentClient, logger *zap.Logger) *ConsentHandlers {
	return &ConsentHandlers{
		userConsentClient: userConsentClient,
		logger:            logger,
	}
}

// AcceptBatchHandler registra o aceite seletivo em lote (modal de termos).
func (h *ConsentHandlers) AcceptBatchHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	tenantId := middlewarePkg.GetTenantId(c)
	token := middlewarePkg.GetTokenFromContext(c)
	codeUser := middlewarePkg.GetUserIDFromContext(c)

	if codeUser == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "Token JWT sem identificador de usuário",
			CorrelationID: correlationID,
		})
	}

	var req userConsentDto.UserConsentAcceptBatchRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	req.UserID = codeUser
	if strings.TrimSpace(req.Email) == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "E-mail é obrigatório",
			CorrelationID: correlationID,
		})
	}
	if len(req.Consents) == 0 {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Informe ao menos um consentimento",
			CorrelationID: correlationID,
		})
	}
	if req.AcceptedAt.IsZero() {
		req.AcceptedAt = time.Now().UTC()
	}
	if req.Geolocation == "" {
		if loc := c.Request().Header.Get("X-Public-Location"); loc != "" {
			decoded, err := url.QueryUnescape(loc)
			if err == nil {
				req.Geolocation = decoded
			} else {
				req.Geolocation = loc
			}
		}
	}
	req.ClientIP = firstNonEmpty(c.Request().Header.Get("X-Public-IP"), c.RealIP())
	req.UserAgent = c.Request().UserAgent()

	result, err := h.userConsentClient.AcceptBatch(c.Request().Context(), req, token, tenantId, correlationID)
	if err != nil {
		h.logger.Error("Erro ao registrar aceite em lote",
			zap.String("correlationId", correlationID),
			zap.String("codeUser", codeUser),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusCreated, result)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
