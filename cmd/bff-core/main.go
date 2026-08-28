// @title BFF-CORE API
// @version 1.0.0
// @description Backend for Frontend responsável pelo core do sistema KeepGuard (registro de usuários)
// @host localhost:8382
// @BasePath /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/keepguard/bff-core/docs" // Importa docs para inicializar Swagger
	httpserver "github.com/keepguard/bff-core/internal/adapters/inbound/http"
	handlersPkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/handlers"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	httpclient "github.com/keepguard/bff-core/internal/adapters/outbound/http/client"
	communicationdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/communication"
	companydecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/company"
	userdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/user"
	messagingDecorator "github.com/keepguard/bff-core/internal/adapters/outbound/messaging/decorator"
	rabbitmqPublisher "github.com/keepguard/bff-core/internal/adapters/outbound/messaging/rabbitmq"
	"github.com/keepguard/bff-core/internal/application/connections"
	"github.com/keepguard/bff-core/internal/application/register"
	"github.com/keepguard/bff-core/internal/infrastructure/cache"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/keepguard/bff-core/internal/infrastructure/logger"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/keepguard/bff-core/internal/infrastructure/resilience"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

func main() {
	// Inicializa logger básico para bootstrap
	bootstrapLogger, _ := zap.NewProduction()
	defer bootstrapLogger.Sync()

	// Carrega configuração
	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Fatal("Erro ao carregar configuração",
			zap.Error(err),
			zap.String("component", "bff-core"),
			zap.String("service", "bff-core"),
		)
	}

	// Inicializa logger padrão
	appLogger, err := logger.New("info", "json")
	if err != nil {
		bootstrapLogger.Fatal("Erro ao inicializar logger",
			zap.Error(err),
			zap.String("component", "bff-core"),
			zap.String("service", "bff-core"),
		)
	}
	defer appLogger.Sync()

	// Loga início da aplicação
	appLogger.Info("Iniciando BFF-CORE",
		zap.String("service", "bff-core"),
		zap.String("component", "bff-core"),
		zap.String("environment", cfg.Env),
		zap.String("version", "1.0.0"),
		zap.String("env", cfg.Env),
		zap.Bool("kibana_enabled", os.Getenv("KIBANA_ENABLED") == "true"),
		zap.String("log_level", cfg.Log.Level),
		zap.String("log_format", cfg.Log.Format),
	)

	// Inicializa métricas
	metricsInstance := metrics.New()

	// Inicializa Circuit Breaker Manager
	cbManager := resilience.NewCircuitBreakerManager(metricsInstance)

	// Configura Circuit Breakers
	authCBConfig := resilience.CircuitBreakerConfig{
		Name:        "ms-auth",
		MaxRequests: 3,                // Máximo 3 requests em half-open para testar recuperação
		Interval:    60 * time.Second, // Janela de amostragem
		Timeout:     10 * time.Second, // Tempo em OPEN antes de tentar HALF-OPEN
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.7
		},
	}
	cbManager.GetOrCreate("ms-auth", authCBConfig)

	userCBConfig := resilience.CircuitBreakerConfig{
		Name:        "ms-user",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.7
		},
	}
	cbManager.GetOrCreate("ms-user", userCBConfig)

	companyCBConfig := resilience.CircuitBreakerConfig{
		Name:        "ms-company",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.7
		},
	}
	cbManager.GetOrCreate("ms-company", companyCBConfig)

	// Converte logger para zap.Logger
	zapLoggerImpl, ok := appLogger.(interface{ GetZapLogger() *zap.Logger })
	if !ok {
		bootstrapLogger.Fatal("Logger não suporta GetZapLogger()",
			zap.String("component", "bff-core"),
			zap.String("service", "bff-core"),
		)
	}
	zapLogger := zapLoggerImpl.GetZapLogger()

	redisClient, err := cache.NewRedisClient(cache.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, zapLogger)
	if err != nil {
		zapLogger.Warn("Aviso ao conectar no Redis", zap.Error(err))
	}

	// =============================================================================
	// INICIALIZAÇÃO DO AUTH CLIENT
	// NOTA: Usado apenas para CRIAR usuários no MS-Auth, NÃO para validar JWT
	// JWT é validado localmente em cada BFF com secret compartilhado
	// =============================================================================
	baseAuthClient := httpclient.NewAuthClient(cfg, zapLogger)
	authClient := baseAuthClient

	// =============================================================================
	// INICIALIZAÇÃO DO USER CLIENT COM DECORATORS
	// =============================================================================
	baseUserClient := httpclient.NewUserClient(cfg, zapLogger)

	userMetricsClient := userdecorator.NewUserMetricsDecorator(baseUserClient, metricsInstance, "ms-user")

	userRedisClient := userdecorator.NewRedisUserCacheDecorator(
		userMetricsClient,
		companydecorator.NewRedisStringCache(redisClient),
		metricsInstance,
		zapLogger,
	)

	userRetryConfig := userdecorator.RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}
	userRetryClient := userdecorator.NewRetryDecorator(userRedisClient, userRetryConfig)

	userCBClient := userdecorator.NewCircuitBreakerDecorator(userRetryClient, cbManager, "ms-user")
	userClient := userdecorator.NewUserLoggingDecorator(userCBClient, zapLogger, "ms-user")

	// =============================================================================
	// INICIALIZAÇÃO DO COMPANY CLIENT COM DECORATORS
	// =============================================================================
	baseCompanyClient := httpclient.NewCompanyClient(cfg, zapLogger)

	companyMetricsClient := companydecorator.NewCompanyMetricsDecorator(baseCompanyClient, metricsInstance, "ms-company")

	companyRedisClient := companydecorator.NewRedisCacheDecorator(
		companyMetricsClient,
		companydecorator.NewRedisStringCache(redisClient),
		metricsInstance,
		zapLogger,
	)

	companyRetryConfig := companydecorator.RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}
	companyRetryClient := companydecorator.NewRetryDecorator(companyRedisClient, companyRetryConfig)

	companyCBClient := companydecorator.NewCircuitBreakerDecorator(companyRetryClient, cbManager, "ms-company")
	companyClient := companydecorator.NewCompanyLoggingDecorator(companyCBClient, zapLogger, "ms-company")

	// =============================================================================
	// INICIALIZAÇÃO DO COMMUNICATION CLIENT COM DECORATORS
	// =============================================================================
	baseCommunicationClient := httpclient.NewCommunicationClient(cfg, zapLogger)

	communicationMetricsClient := communicationdecorator.NewCommunicationMetricsDecorator(baseCommunicationClient, metricsInstance, "ms-communication")

	communicationRetryConfig := communicationdecorator.RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.5,
		Jitter:       true,
	}
	communicationRetryClient := communicationdecorator.NewRetryDecorator(communicationMetricsClient, communicationRetryConfig)

	communicationClient := communicationdecorator.NewCommunicationLoggingDecorator(communicationRetryClient, zapLogger, "ms-communication")

	// =============================================================================
	// INICIALIZAÇÃO DO USER CONSENT CLIENT E CONSENT DOCUMENT CLIENT
	// =============================================================================
	userConsentClient := httpclient.NewUserConsentClient(cfg, zapLogger)
	consentDocumentClient := httpclient.NewConsentDocumentClient(cfg, zapLogger)

	// =============================================================================
	// INICIALIZAÇÃO DO MESSAGE PUBLISHER COM DECORATORS
	// =============================================================================
	rabbitPublisher, err := rabbitmqPublisher.NewMessagePublisher(&cfg.RabbitMQ, zapLogger)
	if err != nil {
		appLogger.Fatal("Erro ao inicializar publisher RabbitMQ",
			zap.Error(err),
			zap.String("component", "bff-core"),
			zap.String("service", "bff-core"),
		)
	}

	loggingPublisher := messagingDecorator.NewLoggingDecorator(rabbitPublisher, zapLogger)
	metricsPublisher := messagingDecorator.NewMetricsDecorator(loggingPublisher, metricsInstance, zapLogger)
	messagePublisher := messagingDecorator.NewCircuitBreakerDecorator(
		metricsPublisher,
		cbManager,
		communicationClient,
		zapLogger,
	)

	appLogger.Info("Message Publisher inicializado com sucesso",
		zap.String("component", "bff-core"),
		zap.String("service", "bff-core"),
		zap.String("exchange", cfg.RabbitMQ.Exchange),
		zap.String("routingKey", cfg.RabbitMQ.RoutingKey),
	)

	// Inicializa use cases de registro
	registerInitUseCase := register.NewRegisterInitUseCase(
		authClient,
		userClient,
		companyClient,
		communicationClient,
		messagePublisher,
		zapLogger,
	)

	registerConfirmUseCase := register.NewRegisterConfirmUseCase(
		userClient,
		companyClient,
		authClient,
		userConsentClient,
		communicationClient,
		messagePublisher,
		zapLogger,
	)

	registerResendUseCase := register.NewRegisterResendUseCase(
		userClient,
		companyClient,
		communicationClient,
		messagePublisher,
		zapLogger,
	)

	// Inicializa handlers HTTP
	registerHandlers := handlersPkg.NewRegisterHandlers(
		registerInitUseCase,
		registerConfirmUseCase,
		registerResendUseCase,
		consentDocumentClient,
		zapLogger,
	)
	userHandlers := handlersPkg.NewUserHandlers(userClient, companyClient, zapLogger)
	consentHandlers := handlersPkg.NewConsentHandlers(userConsentClient, zapLogger)

	connectionsService := connections.NewService(cfg.ConnectionsHealth, connections.NewStore(redisClient), zapLogger)
	connectionsHandlers := handlersPkg.NewConnectionsHandlers(connectionsService, zapLogger)
	httpHandlers := handlersPkg.NewCombinedHandlers(registerHandlers, userHandlers, consentHandlers, connectionsHandlers)

	rateLimiterMiddleware := middlewarePkg.NewRateLimiterMiddleware(redisClient, cfg.RateLimit, zapLogger, metricsInstance)

	// Inicializa servidor HTTP com Rate Limiting
	server := httpserver.NewServer(cfg, appLogger, metricsInstance, rateLimiterMiddleware, companyClient)
	server.SetupRoutes(httpHandlers)

	// =============================================================================
	// SERVIDOR DE MÉTRICAS PROMETHEUS
	// =============================================================================
	go func() {
		metricsPort := cfg.Metrics.Port
		if metricsPort == "" {
			metricsPort = "9092"
		}

		metricsMux := http.NewServeMux()
		metricsMux.Handle(cfg.Metrics.ScrapePath, promhttp.Handler())

		appLogger.Info("Servidor de métricas iniciado",
			zap.String("service", "bff-core"),
			zap.String("port", metricsPort),
			zap.String("path", cfg.Metrics.ScrapePath),
		)

		if err := http.ListenAndServe(":"+metricsPort, metricsMux); err != nil {
			appLogger.Error("Erro no servidor de métricas",
				zap.Error(err),
				zap.String("service", "bff-core"),
			)
		}
	}()

	// Inicia servidor em goroutine
	go func() {
		appLogger.Info("Iniciando servidor HTTP",
			zap.String("port", cfg.Server.Port),
			zap.String("env", cfg.Env),
			zap.String("component", "bff-core"),
			zap.String("service", "bff-core"),
			zap.String("environment", cfg.Env),
			zap.String("version", "1.0.0"),
		)

		if err := server.Start(); err != nil {
			appLogger.Fatal("Erro ao iniciar servidor",
				zap.Error(err),
				zap.String("component", "bff-core"),
				zap.String("service", "bff-core"),
				zap.String("environment", cfg.Env),
				zap.String("version", "1.0.0"),
			)
		}
	}()

	// Aguarda sinal para shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...",
		zap.String("component", "bff-core"),
		zap.String("service", "bff-core"),
		zap.String("environment", cfg.Env),
		zap.String("version", "1.0.0"),
	)

	// Shutdown graceful
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fechar publisher RabbitMQ
	if err := messagePublisher.Close(); err != nil {
		appLogger.Error("Erro ao fechar Message Publisher",
			zap.Error(err),
			zap.String("component", "bff-core"),
			zap.String("service", "bff-core"),
		)
	}

	if err := server.Stop(ctx); err != nil {
		appLogger.Error("Erro durante shutdown",
			zap.Error(err),
			zap.String("component", "bff-core"),
			zap.String("service", "bff-core"),
			zap.String("environment", cfg.Env),
			zap.String("version", "1.0.0"),
		)
	}

	appLogger.Info("Server stopped",
		zap.String("component", "bff-core"),
		zap.String("service", "bff-core"),
		zap.String("environment", cfg.Env),
		zap.String("version", "1.0.0"),
	)
}
