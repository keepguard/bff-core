package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/keepguard/bff-core/internal/infrastructure/logger"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// middlewareImpl representa os middlewares HTTP
type middlewareImpl struct {
	logger  *zap.Logger
	metrics *metrics.Metrics
}

// NewMiddleware cria um novo middleware
func NewMiddleware(logger *zap.Logger) Middleware {
	return &middlewareImpl{
		logger: logger,
	}
}

// NewMiddlewareWithMetrics cria um novo middleware com suporte a métricas
func NewMiddlewareWithMetrics(logger *zap.Logger, metrics *metrics.Metrics) Middleware {
	return &middlewareImpl{
		logger:  logger,
		metrics: metrics,
	}
}

// NewMiddlewareWithLogger cria um novo middleware usando a interface logger.Logger
func NewMiddlewareWithLogger(log logger.Logger) Middleware {
	zapLogger, _ := zap.NewDevelopment()
	return &middlewareImpl{
		logger: zapLogger,
	}
}

// RequestIDMiddleware adiciona ID de requisição (hop local, não é o ID de auditoria)
func (m *middlewareImpl) RequestIDMiddleware() echo.MiddlewareFunc {
	return middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return generateRequestID()
		},
	})
}

// CorrelationIDMiddleware garante X-Correlation-ID (UUID). Se o cliente não enviar, gera.
func (m *middlewareImpl) CorrelationIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			EnsureCorrelationID(c)
			return next(c)
		}
	}
}

// LoggingMiddleware registra logs das requisições
func (m *middlewareImpl) LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Extrai informações da requisição
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			correlationID := c.Request().Header.Get("X-Correlation-ID")
			tenantId := c.Request().Header.Get("X-Tenant-Id")
			userID := c.Request().Header.Get("X-User-ID")
			method := c.Request().Method
			path := c.Path()
			uri := c.Request().RequestURI
			userAgent := c.Request().UserAgent()
			remoteIP := c.RealIP()

			// Executa o handler
			err := next(c)

			// Calcula métricas
			duration := time.Since(start)
			status := c.Response().Status
			responseSize := c.Response().Size

			// Determina o nível do log baseado no status
			var logFunc func(string, ...zap.Field)

			switch {
			case status >= 500:
				logFunc = m.logger.Error
			case status >= 400:
				logFunc = m.logger.Warn
			default:
				logFunc = m.logger.Info
			}

			// Log estruturado
			logFunc("HTTP request completed",
				zap.String("requestId", requestID),
				zap.String("correlationId", correlationID),
				zap.String("tenantId", tenantId),
				zap.String("userId", userID),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("uri", uri),
				zap.Int("status", status),
				zap.Int64("latencyMs", duration.Milliseconds()),
				zap.String("duration", duration.String()),
				zap.String("userAgent", userAgent),
				zap.String("ip", remoteIP),
				zap.Int64("responseSize", responseSize),
				zap.String("component", "bff-core"),
				zap.String("service", "bff-core"),
				zap.String("environment", getEnvOrDefault("ENV", "local")),
				zap.String("version", "1.0.0"),
			)

			return err
		}
	}
}

// RecoverMiddleware recupera de panics
func (m *middlewareImpl) RecoveryMiddleware() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			userID := c.Request().Header.Get("X-User-ID")

			m.logger.Error("Panic recovered",
				zap.String("requestId", requestID),
				zap.String("correlationId", c.Request().Header.Get("X-Correlation-ID")),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Path()),
				zap.String("uri", c.Request().RequestURI),
				zap.String("userAgent", c.Request().UserAgent()),
				zap.String("ip", c.RealIP()),
				zap.String("userId", userID),
				zap.Error(err),
				zap.String("stack", string(stack)),
				zap.String("component", "bff-core"),
				zap.String("service", "bff-core"),
				zap.String("environment", getEnvOrDefault("ENV", "local")),
				zap.String("version", "1.0.0"),
			)
			return nil
		},
	})
}

// CORSMiddleware configura CORS
func (m *middlewareImpl) CORSMiddleware() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestID,
			"X-User-ID",
			"X-Correlation-ID",
			"X-Tenant-Id",
			"X-Client-Id",
			"X-Client-ID",
			"X-Device-Id",
			"X-Device-Name",
			"X-Device-Type",
			"X-Public-IP",
			"X-Public-Location",
			"Idempotency-Key",
		},
		ExposeHeaders: []string{
			echo.HeaderXRequestID,
			"X-User-ID",
			"X-Correlation-ID",
			"X-Tenant-Id",
			"X-Device-Id",
		},
		MaxAge: 86400, // 24 horas
	})
}

// SecurityMiddleware adiciona headers de segurança
func (m *middlewareImpl) SecurityMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Remove header do servidor
			c.Response().Header().Set(echo.HeaderServer, "")

			// Adiciona headers de segurança
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			return next(c)
		}
	}
}

// MetricsMiddleware registra métricas das requisições
func (m *middlewareImpl) MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status

			if m.metrics != nil {
				path := c.Path()
				if path == "" {
					path = c.Request().URL.Path
				}
				m.metrics.RecordHTTPRequest(c.Request().Method, path, status, duration)
			}

			m.logger.Debug("HTTP request metrics",
				zap.String("method", c.Request().Method),
				zap.String("path", c.Path()),
				zap.Int("status", status),
				zap.Duration("duration", duration),
				zap.String("requestId", c.Response().Header().Get(echo.HeaderXRequestID)),
				zap.String("correlationId", c.Request().Header.Get("X-Correlation-ID")),
			)

			return err
		}
	}
}

// TimeoutMiddleware adiciona timeout às requisições
func (m *middlewareImpl) TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: timeout,
	})
}

// generateRequestID gera um ID único para a requisição (hop HTTP)
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.FormatInt(time.Now().Unix(), 36)
}

func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return generateRequestID()
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	hexStr := hex.EncodeToString(b)
	return hexStr[0:8] + "-" + hexStr[8:12] + "-" + hexStr[12:16] + "-" + hexStr[16:20] + "-" + hexStr[20:32]
}

// GetTraceID retorna o request ID do hop (não usar para auditoria).
func GetTraceID(c echo.Context) string {
	return GetCorrelationID(c)
}

// GetUserID extrai o user ID do contexto (se disponível)
func GetUserID(c echo.Context) string {
	return c.Request().Header.Get("X-User-ID")
}

// SetUserID define o user ID no contexto
func SetUserID(c echo.Context, userID string) {
	c.Request().Header.Set("X-User-ID", userID)
	c.Response().Header().Set("X-User-ID", userID)
}

// GetCorrelationID extrai o correlation ID (UUID) do request/response.
func GetCorrelationID(c echo.Context) string {
	return EnsureCorrelationID(c)
}

// EnsureCorrelationID lê ou gera UUID e ecoa no header.
func EnsureCorrelationID(c echo.Context) string {
	correlationID := strings.TrimSpace(c.Request().Header.Get("X-Correlation-ID"))
	if correlationID == "" {
		correlationID = strings.TrimSpace(c.Response().Header().Get("X-Correlation-ID"))
	}
	if correlationID == "" {
		correlationID = generateUUID()
	}
	SetCorrelationID(c, correlationID)
	return correlationID
}

// SetCorrelationID define o correlation ID no contexto
func SetCorrelationID(c echo.Context, correlationID string) {
	c.Request().Header.Set("X-Correlation-ID", correlationID)
	c.Response().Header().Set("X-Correlation-ID", correlationID)
}

// GetTenantId extrai o X-Tenant-Id do contexto
func GetTenantId(c echo.Context) string {
	return c.Request().Header.Get("X-Tenant-Id")
}

// SetTenantId define o X-Tenant-Id no contexto
func SetTenantId(c echo.Context, tenantId string) {
	c.Request().Header.Set("X-Tenant-Id", tenantId)
	c.Response().Header().Set("X-Tenant-Id", tenantId)
}

// getEnvOrDefault retorna valor da variável de ambiente ou valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
