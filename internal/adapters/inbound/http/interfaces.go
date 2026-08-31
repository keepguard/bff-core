package http

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Handler define a interface para handlers HTTP
type Handler interface {
	InitRegisterHandler(c echo.Context) error
	ConfirmRegisterHandler(c echo.Context) error
	ResendRegisterTokenHandler(c echo.Context) error
	GetPublishedConsentsHandler(c echo.Context) error
	GetLatestByTypeHandler(c echo.Context) error
	GetMeHandler(c echo.Context) error
	AcceptBatchHandler(c echo.Context) error
	GetConnectionsHealthHandler(c echo.Context) error
	ListAuditsHandler(c echo.Context) error
	GetAuditHandler(c echo.Context) error
	ListGuardianIncidentsHandler(c echo.Context) error
	GetGuardianIncidentHandler(c echo.Context) error
	ExecuteGuardianActionHandler(c echo.Context) error
	ListGuardianRecipientsHandler(c echo.Context) error
	UpsertGuardianRecipientHandler(c echo.Context) error
	PatchGuardianRecipientHandler(c echo.Context) error
	ListOAuthClientsHandler(c echo.Context) error
	GetOAuthClientHandler(c echo.Context) error
	CreateOAuthClientHandler(c echo.Context) error
	ListOAuthServiceRolesHandler(c echo.Context) error
	BlockOAuthClientHandler(c echo.Context) error
	UnblockOAuthClientHandler(c echo.Context) error
	DeleteOAuthClientHandler(c echo.Context) error
}

// Middleware define a interface para middlewares HTTP
type Middleware interface {
	RequestIDMiddleware() echo.MiddlewareFunc
	CorrelationIDMiddleware() echo.MiddlewareFunc
	LoggingMiddleware() echo.MiddlewareFunc
	RecoveryMiddleware() echo.MiddlewareFunc
	CORSMiddleware() echo.MiddlewareFunc
	SecurityMiddleware() echo.MiddlewareFunc
	MetricsMiddleware() echo.MiddlewareFunc
	TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc
}

// Server define a interface para o servidor HTTP
type Server interface {
	Start() error
	Stop(ctx context.Context) error
	SetupRoutes(handlers Handler)
}

// Logger define a interface para logging
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}
