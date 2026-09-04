package http

import (
	"context"
	"net/http"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	auditport "github.com/keepguard/bff-core/internal/domain/ports/audit"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/keepguard/bff-core/internal/infrastructure/logger"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/keepguard/bff-core/internal/infrastructure/validation"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

// serverImpl representa o servidor HTTP
type serverImpl struct {
	echo        *echo.Echo
	config      *config.Config
	logger      logger.Logger
	metrics     *metrics.Metrics
	rateLimiter *middlewarePkg.RateLimiterMiddleware
	jwt         *middlewarePkg.JWTMiddleware
}

// NewServer cria um novo servidor HTTP
func NewServer(
	config *config.Config,
	logger logger.Logger,
	metrics *metrics.Metrics,
	rateLimiter *middlewarePkg.RateLimiterMiddleware,
	redisClient *redis.Client,
	companyClient client.CompanyClient,
	auditPublisher auditport.EventPublisher,
) Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middlewares
	zapLogger, _ := zap.NewDevelopment()
	middlewareInstance := middlewarePkg.NewMiddlewareWithMetrics(zapLogger, metrics)
	validator := validation.NewValidator()

	// Configura validador personalizado para o Echo
	e.Validator = middlewarePkg.NewCustomValidator()

	e.Use(middlewareInstance.RequestIDMiddleware())
	e.Use(middlewareInstance.CorrelationIDMiddleware())
	e.Use(middlewarePkg.AuditMiddleware(auditPublisher, "bff-core"))
	e.Use(middlewareInstance.RecoveryMiddleware())
	e.Use(middlewareInstance.LoggingMiddleware())
	e.Use(middlewarePkg.ValidationMiddleware(validator))
	e.Use(middlewareInstance.CORSMiddleware())
	e.Use(middlewareInstance.SecurityMiddleware())
	e.Use(middlewareInstance.MetricsMiddleware())
	e.Use(middlewareInstance.TimeoutMiddleware(config.Server.RequestTimeout))
	if companyClient != nil {
		e.Use(middlewarePkg.CompanyResolveMiddleware(companyClient))
	}

	return &serverImpl{
		echo:        e,
		config:      config,
		logger:      logger,
		metrics:     metrics,
		rateLimiter: rateLimiter,
		jwt:         middlewarePkg.NewJWTMiddlewareWithRedis(config.JWT.Secret, redisClient, zapLogger),
	}
}

// Start inicia o servidor
func (s *serverImpl) Start() error {
	return s.echo.Start(":" + s.config.Server.Port)
}

// Stop para o servidor
func (s *serverImpl) Stop(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

// SetupRoutes configura as rotas da API com proteção de Rate Limit
func (s *serverImpl) SetupRoutes(handlers Handler) {
	// Health check
	s.echo.GET("/health", s.HealthHandler)

	// Swagger
	s.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	// Cria o middleware de endpoint público
	publicEndpoint := middlewarePkg.NewPublicEndpoint()

	rl := s.rateLimiter
	rules := s.config.RateLimit.Rules

	// Register routes
	userGroup := s.echo.Group("/api/v1")

	// ========================================================================
	// ROTAS PÚBLICAS - Registro de usuários e Documentos Legais com Rate Limit
	// ========================================================================
	userGroup.POST("/register/init", handlers.InitRegisterHandler, publicEndpoint.Middleware(), rl.Limit("register_init", rules.RegisterInit))
	userGroup.POST("/register/confirm", handlers.ConfirmRegisterHandler, publicEndpoint.Middleware(), rl.Limit("register_confirm", rules.RegisterConfirm))
	userGroup.POST("/register/resend", handlers.ResendRegisterTokenHandler, publicEndpoint.Middleware(), rl.Limit("register_resend", rules.RegisterResend))
	userGroup.GET("/consents/published", handlers.GetPublishedConsentsHandler, publicEndpoint.Middleware(), rl.Limit("consents", rules.Consents))
	userGroup.GET("/consents/type/:type/latest", handlers.GetLatestByTypeHandler, publicEndpoint.Middleware(), rl.Limit("consents", rules.Consents))

	// ========================================================================
	// ROTAS AUTENTICADAS - Perfil do usuário logado
	// ========================================================================
	userGroup.GET("/users/me", handlers.GetMeHandler, s.jwt.Middleware(), rl.Limit("users_me", rules.UsersMe))
	userGroup.POST("/user-consents/accept-batch", handlers.AcceptBatchHandler, s.jwt.Middleware(), rl.Limit("consents", rules.Consents))
	userGroup.GET("/core/connections/health", handlers.GetConnectionsHealthHandler,
		s.jwt.Middleware(),
		middlewarePkg.RequireAnyRole("ADMIN", "SYSTEM"),
		rl.Limit("connections_health", rules.ConnectionsHealth),
	)
	// /api/v1/core/audits usa o PathPrefix já publicado no Traefik.
	// /api/v1/audits permanece para o IngressRoute atualizado.
	auditRead := []echo.MiddlewareFunc{
		s.jwt.Middleware(),
		middlewarePkg.RequireAuditRead(),
		rl.Limit("audits", rules.Audits),
	}
	userGroup.GET("/core/audits", handlers.ListAuditsHandler, auditRead...)
	userGroup.GET("/core/audits/:eventId", handlers.GetAuditHandler, auditRead...)
	userGroup.GET("/audits", handlers.ListAuditsHandler, auditRead...)
	userGroup.GET("/audits/:eventId", handlers.GetAuditHandler, auditRead...)

	guardianAdmin := []echo.MiddlewareFunc{
		s.jwt.Middleware(),
		middlewarePkg.RequireAnyRole("ADMIN", "SYSTEM"),
		rl.Limit("guardian", rules.Guardian),
	}
	userGroup.GET("/core/guardian/incidents", handlers.ListGuardianIncidentsHandler, guardianAdmin...)
	userGroup.GET("/core/guardian/incidents/:id", handlers.GetGuardianIncidentHandler, guardianAdmin...)
	userGroup.POST("/core/guardian/incidents/:id/actions", handlers.ExecuteGuardianActionHandler, guardianAdmin...)
	userGroup.GET("/core/guardian/alert-recipients", handlers.ListGuardianRecipientsHandler, guardianAdmin...)
	userGroup.PUT("/core/guardian/alert-recipients", handlers.UpsertGuardianRecipientHandler, guardianAdmin...)
	userGroup.PATCH("/core/guardian/alert-recipients/:id", handlers.PatchGuardianRecipientHandler, guardianAdmin...)

	oauthAdmin := []echo.MiddlewareFunc{
		s.jwt.Middleware(),
		middlewarePkg.RequireAnyRole("ADMIN", "SYSTEM"),
		rl.Limit("guardian", rules.Guardian),
	}
	userGroup.GET("/core/oauth/clients", handlers.ListOAuthClientsHandler, oauthAdmin...)
	userGroup.GET("/core/oauth/clients/service-roles", handlers.ListOAuthServiceRolesHandler, oauthAdmin...)
	userGroup.GET("/core/oauth/clients/:id", handlers.GetOAuthClientHandler, oauthAdmin...)
	userGroup.POST("/core/oauth/clients", handlers.CreateOAuthClientHandler, oauthAdmin...)
	userGroup.PUT("/core/oauth/clients/:id", handlers.UpdateOAuthClientHandler, oauthAdmin...)
	userGroup.POST("/core/oauth/clients/:id/block", handlers.BlockOAuthClientHandler, oauthAdmin...)
	userGroup.POST("/core/oauth/clients/:id/unblock", handlers.UnblockOAuthClientHandler, oauthAdmin...)
	userGroup.DELETE("/core/oauth/clients/:id", handlers.DeleteOAuthClientHandler, oauthAdmin...)

	collectorAdmin := []echo.MiddlewareFunc{
		s.jwt.Middleware(),
		middlewarePkg.RequireAnyRole("ADMIN", "SYSTEM"),
		rl.Limit("collector", rules.Collector),
	}
	collectorIncidents := []echo.MiddlewareFunc{
		s.jwt.Middleware(),
		middlewarePkg.RequireAnyRole("ADMIN", "MANAGER"),
		rl.Limit("collector", rules.Collector),
	}
	userGroup.GET("/core/collector/agents", handlers.ListCollectorAgentsHandler, collectorAdmin...)
	userGroup.POST("/core/collector/agents", handlers.CreateCollectorAgentHandler, collectorAdmin...)
	userGroup.POST("/core/collector/agents/bulk", handlers.BulkCollectorAgentsHandler, collectorAdmin...)
	userGroup.GET("/core/collector/agents/bulk-operations/active", handlers.GetCollectorActiveBulkOperationHandler, collectorAdmin...)
	userGroup.GET("/core/collector/agents/bulk-operations/:id", handlers.GetCollectorBulkOperationHandler, collectorAdmin...)
	userGroup.GET("/core/collector/agents/:id", handlers.GetCollectorAgentHandler, collectorAdmin...)
	userGroup.PUT("/core/collector/agents/:id", handlers.UpdateCollectorAgentHandler, collectorAdmin...)
	userGroup.POST("/core/collector/agents/:id/enable", handlers.EnableCollectorAgentHandler, collectorAdmin...)
	userGroup.POST("/core/collector/agents/:id/disable", handlers.DisableCollectorAgentHandler, collectorAdmin...)
	userGroup.POST("/core/collector/agents/:id/test", handlers.TestCollectorAgentHandler, collectorAdmin...)
	userGroup.POST("/core/collector/agents/:id/run", handlers.RunCollectorAgentHandler, collectorAdmin...)
	userGroup.GET("/core/collector/agents/:id/executions", handlers.ListCollectorAgentExecutionsHandler, collectorAdmin...)
	userGroup.GET("/core/collector/executions/:executionId/payloads", handlers.GetCollectorExecutionPayloadsHandler, collectorAdmin...)
	userGroup.GET("/core/collector/data-sources", handlers.ListCollectorDataSourcesHandler, collectorAdmin...)
	userGroup.POST("/core/collector/data-sources", handlers.CreateCollectorDataSourceHandler, collectorAdmin...)
	userGroup.GET("/core/collector/data-sources/:id", handlers.GetCollectorDataSourceHandler, collectorAdmin...)
	userGroup.PUT("/core/collector/data-sources/:id", handlers.UpdateCollectorDataSourceHandler, collectorAdmin...)
	userGroup.POST("/core/collector/data-sources/:id/enable", handlers.EnableCollectorDataSourceHandler, collectorAdmin...)
	userGroup.POST("/core/collector/data-sources/:id/disable", handlers.DisableCollectorDataSourceHandler, collectorAdmin...)
	userGroup.POST("/core/collector/data-sources/:id/propagate", handlers.PropagateCollectorDataSourceHandler, collectorAdmin...)
	userGroup.DELETE("/core/collector/data-sources/:id", handlers.DeleteCollectorDataSourceHandler, collectorAdmin...)
	userGroup.GET("/core/collector/incidents", handlers.ListCollectorIncidentsHandler, collectorIncidents...)
	userGroup.GET("/core/collector/agents/:id/incidents", handlers.ListCollectorAgentIncidentsHandler, collectorIncidents...)
	userGroup.POST("/core/collector/incidents/:id/acknowledge", handlers.AcknowledgeCollectorIncidentHandler, collectorIncidents...)
	userGroup.POST("/core/collector/incidents/:id/resolve", handlers.ResolveCollectorIncidentHandler, collectorIncidents...)
	userGroup.GET("/core/collector/incidents/:id/suggestion", handlers.GetCollectorIncidentSuggestionHandler, collectorIncidents...)
	userGroup.POST("/core/collector/incidents/:id/apply-successor", handlers.ApplyCollectorIncidentSuccessorHandler, collectorIncidents...)
	userGroup.DELETE("/core/collector/agents/:id", handlers.DeleteCollectorAgentHandler, collectorAdmin...)

	knowledgeAdmin := []echo.MiddlewareFunc{
		s.jwt.Middleware(),
		middlewarePkg.RequireKnowledgeRead(),
		rl.Limit("knowledge", rules.Knowledge),
	}
	userGroup.POST("/core/knowledge/ask", handlers.AskKnowledgeHandler, knowledgeAdmin...)

	llmRead := []echo.MiddlewareFunc{
		s.jwt.Middleware(),
		middlewarePkg.RequireLlmRead(),
		rl.Limit("llm", rules.Llm),
	}
	llmWrite := []echo.MiddlewareFunc{
		s.jwt.Middleware(),
		middlewarePkg.RequireLlmWrite(),
		rl.Limit("llm", rules.Llm),
	}
	userGroup.GET("/core/llm/providers", handlers.ListLlmProvidersHandler, llmRead...)
	userGroup.POST("/core/llm/providers", handlers.CreateLlmProviderHandler, llmWrite...)
	userGroup.PATCH("/core/llm/providers/:id", handlers.UpdateLlmProviderHandler, llmWrite...)
	userGroup.POST("/core/llm/providers/:id/enable", handlers.EnableLlmProviderHandler, llmWrite...)
	userGroup.POST("/core/llm/providers/:id/disable", handlers.DisableLlmProviderHandler, llmWrite...)
	userGroup.POST("/core/llm/complete", handlers.CompleteLlmHandler, llmWrite...)
	userGroup.GET("/core/llm/usage", handlers.ListLlmUsageHandler, llmRead...)
	userGroup.GET("/core/llm/usage/:id", handlers.GetLlmUsageHandler, llmRead...)
	userGroup.GET("/core/llm/alert-rules", handlers.ListLlmAlertRulesHandler, llmRead...)
	userGroup.POST("/core/llm/alert-rules", handlers.CreateLlmAlertRuleHandler, llmWrite...)
	userGroup.PATCH("/core/llm/alert-rules/:id", handlers.UpdateLlmAlertRuleHandler, llmWrite...)
	userGroup.POST("/core/llm/alert-rules/:id/enable", handlers.EnableLlmAlertRuleHandler, llmWrite...)
	userGroup.POST("/core/llm/alert-rules/:id/disable", handlers.DisableLlmAlertRuleHandler, llmWrite...)
	userGroup.GET("/core/llm/alert-firings", handlers.ListLlmAlertFiringsHandler, llmRead...)

	s.logger.Info("Rotas configuradas com sucesso com proteção de Rate Limit",
		zap.String("port", s.config.Server.Port),
	)
}

// HealthHandler verifica o status do serviço
// @Summary Health check
// @Description Verifica se o serviço está funcionando corretamente
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Serviço funcionando"
// @Router /health [get]
func (s *serverImpl) HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "bff-core",
		"version": "1.0.0",
	})
}
