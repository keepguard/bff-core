package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/keepguard/bff-core/internal/infrastructure/config"
	metricsPkg "github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiterMiddleware gerencia rate limiting distribuído usando Redis
type RateLimiterMiddleware struct {
	redisClient *redis.Client
	config      config.RateLimitConfig
	logger      *zap.Logger
	metrics     *metricsPkg.Metrics
}

// NewRateLimiterMiddleware cria uma nova instância do middleware de rate limit
func NewRateLimiterMiddleware(redisClient *redis.Client, cfg config.RateLimitConfig, logger *zap.Logger, metrics *metricsPkg.Metrics) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		redisClient: redisClient,
		config:      cfg,
		logger:      logger,
		metrics:     metrics,
	}
}

// Limit cria um middleware do Echo aplicando o rate limit configurado para uma ação específica
func (r *RateLimiterMiddleware) Limit(action string, rule config.RateLimitRule) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Se o rate limiting estiver desativado globalmente ou na regra, segue em frente
			if !r.config.Enabled || rule.Limit <= 0 {
				return next(c)
			}

			// Se o Redis não estiver conectado, aplica fail-open defensivo
			if r.redisClient == nil {
				return next(c)
			}

			clientIP := c.RealIP()
			if clientIP == "" {
				clientIP = "unknown"
			}

			keyIdentifier := clientIP
			identifierType := "ip"
			if userID := GetUserIDFromContext(c); userID != "" {
				keyIdentifier = userID
				identifierType = "user"
			} else if userID := GetUserID(c); userID != "" {
				keyIdentifier = userID
				identifierType = "user"
			}

			redisKey := fmt.Sprintf("rl:bff-core:%s:%s", action, keyIdentifier)
			ctx, cancel := context.WithTimeout(c.Request().Context(), 500*time.Millisecond)
			defer cancel()

			window := rule.Window
			if window <= 0 {
				window = 60 * time.Second
			}

			// Incremento atômico
			count, err := r.redisClient.Incr(ctx, redisKey).Result()
			if err != nil {
				r.logger.Warn("Erro ao consultar Redis para RateLimit (fail-open)",
					zap.String("key", redisKey),
					zap.Error(err),
				)
				return next(c)
			}

			if count == 1 {
				r.redisClient.Expire(ctx, redisKey, window)
			}

			remaining := rule.Limit - int(count)
			if remaining < 0 {
				remaining = 0
			}

			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rule.Limit))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			if count > int64(rule.Limit) {
				ttl, _ := r.redisClient.TTL(ctx, redisKey).Result()
				retrySeconds := int(ttl.Seconds())
				if retrySeconds <= 0 {
					retrySeconds = 1
				}

				c.Response().Header().Set("Retry-After", fmt.Sprintf("%d", retrySeconds))

				// Registra métrica dedicada de bloqueio
				if r.metrics != nil {
					r.metrics.RecordRateLimitBlocked(action, identifierType)
				}

				r.logger.Warn("🚨 [BFF-Core Rate Limit Excedido]",
					zap.String("action", action),
					zap.String("identifier", keyIdentifier),
					zap.Int64("count", count),
					zap.Int("limit", rule.Limit),
					zap.Int("retryAfter", retrySeconds),
				)

				return c.JSON(http.StatusTooManyRequests, pkg.ErrorResponse{
					Error:   "RATE_LIMIT_EXCEEDED",
					Message: "Muitas tentativas. Por favor, aguarde antes de tentar novamente.",
					TraceID: GetTraceID(c),
				})
			}

			return next(c)
		}
	}
}
