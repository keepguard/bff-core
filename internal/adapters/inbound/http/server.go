package http

import (
	"context"
	"net/http"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/keepguard/bff-core/internal/infrastructure/logger"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/keepguard/bff-core/internal/infrastructure/validation"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

// serverImpl representa o servidor HTTP
type serverImpl struct {
	echo    *echo.Echo
	config  *config.Config
	logger  logger.Logger
	metrics *metrics.Metrics
}

// NewServer cria um novo servidor HTTP
func NewServer(
	config *config.Config,
	logger logger.Logger,
	metrics *metrics.Metrics,
) Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middlewares
	middlewareInstance := middlewarePkg.NewMiddlewareWithLogger(logger)
	validator := validation.NewValidator()

	// Configura validador personalizado para o Echo
	e.Validator = middlewarePkg.NewCustomValidator()

	e.Use(middlewareInstance.RequestIDMiddleware())
	e.Use(middlewareInstance.RecoveryMiddleware())
	e.Use(middlewareInstance.LoggingMiddleware())
	e.Use(middlewarePkg.ValidationMiddleware(validator))
	e.Use(middlewareInstance.CORSMiddleware())
	e.Use(middlewareInstance.SecurityMiddleware())
	e.Use(middlewareInstance.MetricsMiddleware())
	e.Use(middlewareInstance.TimeoutMiddleware(config.Server.RequestTimeout))

	return &serverImpl{
		echo:    e,
		config:  config,
		logger:  logger,
		metrics: metrics,
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

// SetupRoutes configura as rotas da API
func (s *serverImpl) SetupRoutes(handlers Handler) {
	// Health check
	s.echo.GET("/health", s.HealthHandler)

	// Swagger
	s.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	// Cria o middleware de endpoint público
	publicEndpoint := middlewarePkg.NewPublicEndpoint()

	// Register routes
	userGroup := s.echo.Group("/api/v1")

	// ========================================================================
	// ROTAS PÚBLICAS - Registro de usuários (não precisam de token)
	// Similar a @PublicEndpoint no Java
	// ========================================================================
	userGroup.POST("/register/init", handlers.InitRegisterHandler, publicEndpoint.Middleware())
	userGroup.POST("/register/confirm", handlers.ConfirmRegisterHandler, publicEndpoint.Middleware())
	userGroup.POST("/register/resend", handlers.ResendRegisterTokenHandler, publicEndpoint.Middleware())

	// ========================================================================
	// ROTAS PROTEGIDAS (Futuro - usar JWTMiddleware com cfg.JWT.Secret)
	// jwtMw := middlewarePkg.NewJWTMiddleware(s.config.JWT.Secret, logger)
	// protectedGroup := s.echo.Group("/api/v1/protected")
	// protectedGroup.Use(jwtMw.Middleware())
	// ========================================================================

	s.logger.Info("Rotas configuradas com sucesso",
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
