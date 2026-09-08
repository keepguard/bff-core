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
	auditdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/audit"
	billingdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/billing"
	collectordecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/collector"
	communicationdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/communication"
	companydecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/company"
	consentdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/consent"
	consentdocdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/consentdocument"
	guardiandecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/guardian"
	knowledgedecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/knowledge"
	llmdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/llmgateway"
	oauthdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/oauthclient"
	tokendecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/servicetoken"
	userdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/user"
	auditPublisher "github.com/keepguard/bff-core/internal/adapters/outbound/messaging/audit"
	messagingDecorator "github.com/keepguard/bff-core/internal/adapters/outbound/messaging/decorator"
	rabbitmqPublisher "github.com/keepguard/bff-core/internal/adapters/outbound/messaging/rabbitmq"
	"github.com/keepguard/bff-core/internal/application/audit"
	appbilling "github.com/keepguard/bff-core/internal/application/billing"
	"github.com/keepguard/bff-core/internal/application/collector"
	"github.com/keepguard/bff-core/internal/application/connections"
	"github.com/keepguard/bff-core/internal/application/consent"
	"github.com/keepguard/bff-core/internal/application/consentdocument"
	"github.com/keepguard/bff-core/internal/application/guardian"
	appknowledge "github.com/keepguard/bff-core/internal/application/knowledge"
	appllm "github.com/keepguard/bff-core/internal/application/llm"
	appoauth "github.com/keepguard/bff-core/internal/application/oauth"
	"github.com/keepguard/bff-core/internal/application/register"
	appuser "github.com/keepguard/bff-core/internal/application/user"
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
	userConsentClient := consentdecorator.New(httpclient.NewUserConsentClient(cfg, zapLogger), zapLogger, metricsInstance, "ms-user-consents")
	consentDocumentClient := consentdocdecorator.New(httpclient.NewConsentDocumentClient(cfg, zapLogger), zapLogger, metricsInstance, "ms-user-consents")

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

	auditEventPublisher, err := auditPublisher.NewPublisher(&cfg.RabbitMQ, zapLogger)
	if err != nil {
		appLogger.Warn("Auditoria RabbitMQ indisponível; eventos não serão publicados",
			zap.Error(err),
			zap.String("component", "bff-core"),
		)
		auditEventPublisher = nil
	}

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
		zapLogger,
	)
	userHandlers := handlersPkg.NewUserHandlers(appuser.NewUserPort(userClient, companyClient), zapLogger)
	consentHandlers := handlersPkg.NewConsentHandlers(
		consent.NewConsentPort(userConsentClient),
		consentdocument.NewConsentDocumentPort(consentDocumentClient),
		zapLogger,
	)

	connectionsService := connections.NewService(cfg.ConnectionsHealth, connections.NewStore(redisClient), zapLogger)
	connectionsHandlers := handlersPkg.NewConnectionsHandlers(connectionsService, zapLogger)
	auditClient := auditdecorator.New(httpclient.NewAuditClient(cfg, zapLogger), zapLogger, metricsInstance, "srv-audit")
	auditHandlers := handlersPkg.NewAuditHandlers(audit.NewAuditPort(auditClient), zapLogger)
	guardianClient := guardiandecorator.New(httpclient.NewGuardianClient(cfg, zapLogger), zapLogger, metricsInstance, "ms-ai-guardian")
	guardianHandlers := handlersPkg.NewGuardianHandlers(guardian.NewGuardianPort(guardianClient), zapLogger)
	oauthClientHTTP := oauthdecorator.New(httpclient.NewOAuthClientHTTP(cfg, zapLogger), zapLogger, metricsInstance, "ms-auth")
	collectorClient := collectordecorator.New(httpclient.NewCollectorClient(cfg, zapLogger), zapLogger, metricsInstance, "srv-collector")
	knowledgeClient := knowledgedecorator.New(httpclient.NewKnowledgeClient(cfg, zapLogger), zapLogger, metricsInstance, "srv-knowledge")
	serviceTokenClient := tokendecorator.New(httpclient.NewBffOAuthTokenClient(cfg, zapLogger), zapLogger, metricsInstance, "ms-auth")
	oauthClientHandlers := handlersPkg.NewOAuthClientHandlers(
		appoauth.NewOAuthPort(oauthClientHTTP, companyClient, collectorClient, zapLogger),
		zapLogger,
	)
	collectorPort := collector.NewCollectorPort(collectorClient, companyClient, knowledgeClient, serviceTokenClient, zapLogger)
	collectorAgentHandlers := handlersPkg.NewCollectorAgentHandlers(collectorPort, zapLogger)
	knowledgeHandlers := handlersPkg.NewKnowledgeHandlers(
		appknowledge.NewKnowledgePort(knowledgeClient, collectorClient, companyClient, serviceTokenClient, zapLogger),
		zapLogger,
	)
	llmClient := llmdecorator.New(httpclient.NewLlmClient(cfg, zapLogger), zapLogger, metricsInstance, "srv-llm-gateway")
	llmHandlers := handlersPkg.NewLlmHandlers(appllm.NewLlmPort(llmClient), zapLogger)
	billingClient := billingdecorator.New(httpclient.NewBillingClient(cfg, zapLogger), zapLogger, metricsInstance, "ms-billing")
	billingHandlers := handlersPkg.NewBillingHandlers(appbilling.NewBillingPort(billingClient), zapLogger)
	httpHandlers := handlersPkg.NewCombinedHandlers(registerHandlers, userHandlers, consentHandlers, connectionsHandlers, auditHandlers, guardianHandlers, oauthClientHandlers, collectorAgentHandlers, knowledgeHandlers, llmHandlers, billingHandlers)

	rateLimiterMiddleware := middlewarePkg.NewRateLimiterMiddleware(redisClient, cfg.RateLimit, zapLogger, metricsInstance)

	// Inicializa servidor HTTP com Rate Limiting e Validação de Sessão/Blacklist via Redis
	server := httpserver.NewServer(cfg, appLogger, metricsInstance, rateLimiterMiddleware, redisClient, companyClient, auditEventPublisher)
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

	if auditEventPublisher != nil {
		if err := auditEventPublisher.Close(); err != nil {
			appLogger.Error("Erro ao fechar Audit Publisher", zap.Error(err))
		}
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
