package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-core/internal/adapters/inbound/http/mapper"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/consent"
	"github.com/keepguard/bff-core/internal/application/consentdocument"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ConsentHandlers struct {
	consents  consent.ConsentPort
	documents consentdocument.ConsentDocumentPort
	logger    *zap.Logger
}

func NewConsentHandlers(consents consent.ConsentPort, documents consentdocument.ConsentDocumentPort, logger *zap.Logger) *ConsentHandlers {
	return &ConsentHandlers{
		consents:  consents,
		documents: documents,
		logger:    logger,
	}
}

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

	var req inboundDto.UserConsentAcceptBatchRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

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
	geolocation := req.Geolocation
	if geolocation == "" {
		if loc := c.Request().Header.Get("X-Public-Location"); loc != "" {
			decoded, err := url.QueryUnescape(loc)
			if err == nil {
				geolocation = decoded
			} else {
				geolocation = loc
			}
		}
	}
	clientIP := firstNonEmpty(c.Request().Header.Get("X-Public-IP"), c.RealIP())
	userAgent := c.Request().UserAgent()

	result, err := h.consents.AcceptBatch(c.Request().Context(), mapper.ToAcceptBatchCommand(
		req, codeUser, token, tenantId, correlationID, clientIP, userAgent, geolocation,
	))
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

func (h *ConsentHandlers) GetPublishedConsentsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		h.logger.Warn("X-Tenant-Id ausente", zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       "Header X-Tenant-Id é obrigatório",
			CorrelationID: correlationID,
		})
	}

	docs, err := h.documents.ListPublished(c.Request().Context(), appdto.ListPublishedConsentsQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
	})
	if err != nil {
		h.logger.Error("Erro ao buscar documentos publicados",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusInternalServerError, pkg.ErrorResponse{
			Error:         "INTERNAL_SERVER_ERROR",
			Message:       "Erro ao buscar documentos de consentimento",
			CorrelationID: correlationID,
		})
	}

	return c.JSON(http.StatusOK, docs)
}

func (h *ConsentHandlers) GetLatestByTypeHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		h.logger.Warn("X-Tenant-Id ausente", zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       "Header X-Tenant-Id é obrigatório",
			CorrelationID: correlationID,
		})
	}

	consentType := c.Param("type")
	if consentType == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_PARAM",
			Message:       "Parâmetro type é obrigatório",
			CorrelationID: correlationID,
		})
	}

	doc, err := h.documents.GetLatestByType(c.Request().Context(), appdto.GetLatestConsentQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		ConsentType:   consentType,
	})
	if err != nil {
		h.logger.Error("Erro ao buscar documento por tipo",
			zap.String("type", consentType),
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusInternalServerError, pkg.ErrorResponse{
			Error:         "INTERNAL_SERVER_ERROR",
			Message:       "Erro ao buscar documento de consentimento",
			CorrelationID: correlationID,
		})
	}

	return c.JSON(http.StatusOK, doc)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
