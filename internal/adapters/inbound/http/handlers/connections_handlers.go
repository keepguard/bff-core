package handlers

import (
	"net/http"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/connections"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ConnectionsHandlers expõe o snapshot autenticado da tela Conexões.
type ConnectionsHandlers struct {
	service *connections.Service
	logger  *zap.Logger
}

func NewConnectionsHandlers(service *connections.Service, logger *zap.Logger) *ConnectionsHandlers {
	return &ConnectionsHandlers{service: service, logger: logger}
}

// GetConnectionsHealthHandler devolve o snapshot (cache Redis 60s).
func (h *ConnectionsHandlers) GetConnectionsHealthHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.service == nil {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Verificação de conexões indisponível",
			CorrelationID: correlationID,
		})
	}
	snap, err := h.service.GetSnapshot(c.Request().Context())
	if err != nil {
		h.logger.Error("Erro ao obter snapshot de conexões",
			zap.String("correlationId", correlationID),
			zap.Error(err),
		)
		return c.JSON(http.StatusInternalServerError, pkg.ErrorResponse{
			Error:         "INTERNAL_ERROR",
			Message:       "Não foi possível verificar as conexões",
			CorrelationID: correlationID,
		})
	}
	return c.JSON(http.StatusOK, snap)
}
