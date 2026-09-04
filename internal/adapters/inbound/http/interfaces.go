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
	UpdateOAuthClientHandler(c echo.Context) error
	ListOAuthServiceRolesHandler(c echo.Context) error
	BlockOAuthClientHandler(c echo.Context) error
	UnblockOAuthClientHandler(c echo.Context) error
	DeleteOAuthClientHandler(c echo.Context) error
	ListCollectorAgentsHandler(c echo.Context) error
	GetCollectorAgentHandler(c echo.Context) error
	CreateCollectorAgentHandler(c echo.Context) error
	UpdateCollectorAgentHandler(c echo.Context) error
	EnableCollectorAgentHandler(c echo.Context) error
	DisableCollectorAgentHandler(c echo.Context) error
	TestCollectorAgentHandler(c echo.Context) error
	RunCollectorAgentHandler(c echo.Context) error
	BulkCollectorAgentsHandler(c echo.Context) error
	GetCollectorBulkOperationHandler(c echo.Context) error
	GetCollectorActiveBulkOperationHandler(c echo.Context) error
	ListCollectorAgentExecutionsHandler(c echo.Context) error
	GetCollectorExecutionPayloadsHandler(c echo.Context) error
	ListCollectorDataSourcesHandler(c echo.Context) error
	GetCollectorDataSourceHandler(c echo.Context) error
	CreateCollectorDataSourceHandler(c echo.Context) error
	UpdateCollectorDataSourceHandler(c echo.Context) error
	EnableCollectorDataSourceHandler(c echo.Context) error
	DisableCollectorDataSourceHandler(c echo.Context) error
	PropagateCollectorDataSourceHandler(c echo.Context) error
	DeleteCollectorDataSourceHandler(c echo.Context) error
	DeleteCollectorAgentHandler(c echo.Context) error
	ListCollectorIncidentsHandler(c echo.Context) error
	ListCollectorAgentIncidentsHandler(c echo.Context) error
	AcknowledgeCollectorIncidentHandler(c echo.Context) error
	ResolveCollectorIncidentHandler(c echo.Context) error
	GetCollectorIncidentSuggestionHandler(c echo.Context) error
	ApplyCollectorIncidentSuccessorHandler(c echo.Context) error
	AskKnowledgeHandler(c echo.Context) error
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
