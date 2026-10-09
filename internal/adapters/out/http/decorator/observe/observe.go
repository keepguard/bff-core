package observe

import (
	"math"
	"math/rand"
	"net/http"
	"time"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

func DefaultRetry() RetryConfig {
	return RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

type Config struct {
	Logger      *zap.Logger
	Metrics     *metrics.Metrics
	ServiceName string
	Retry       RetryConfig
}

func Call[T any](cfg Config, operation, method, endpoint, correlationID string, fn func() (T, error)) (T, error) {
	var zero T
	attempts := cfg.Retry.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			time.Sleep(delay(cfg.Retry, attempt-1))
		}
		start := time.Now()
		if cfg.Logger != nil {
			cfg.Logger.Info("Iniciando requisição",
				zap.String("service", cfg.ServiceName),
				zap.String("operation", operation),
				zap.String("correlationID", correlationID),
			)
		}
		result, err := fn()
		duration := time.Since(start)
		status := statusCode(err)
		if cfg.Metrics != nil {
			cfg.Metrics.RecordUpstreamRequest(cfg.ServiceName, method, endpoint, status, duration)
			if err != nil {
				cfg.Metrics.RecordUpstreamError(cfg.ServiceName, method, endpoint, errorType(err))
			}
		}
		if err == nil {
			if cfg.Logger != nil {
				cfg.Logger.Info("Requisição concluída com sucesso",
					zap.String("service", cfg.ServiceName),
					zap.String("operation", operation),
					zap.String("correlationID", correlationID),
					zap.Duration("duration", duration),
				)
			}
			return result, nil
		}
		lastErr = err
		if cfg.Logger != nil {
			cfg.Logger.Error("Erro na requisição",
				zap.String("service", cfg.ServiceName),
				zap.String("operation", operation),
				zap.String("correlationID", correlationID),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		}
		if !retryable(err) {
			return zero, err
		}
	}
	return zero, lastErr
}

func CallErr(cfg Config, operation, method, endpoint, correlationID string, fn func() error) error {
	_, err := Call(cfg, operation, method, endpoint, correlationID, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}

func statusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.Code
	}
	return http.StatusInternalServerError
}

func errorType(err error) string {
	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.Message
	}
	return "unknown"
}

func retryable(err error) bool {
	if err == nil {
		return false
	}
	if httpErr, ok := err.(*appdto.HTTPError); ok {
		switch httpErr.Code {
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
		return false
	}
	return true
}

func delay(cfg RetryConfig, attempt int) time.Duration {
	d := float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt))
	if d > float64(cfg.MaxDelay) {
		d = float64(cfg.MaxDelay)
	}
	wait := time.Duration(d)
	if cfg.Jitter && wait > 0 {
		wait = time.Duration(rand.Int63n(int64(wait) + 1))
	}
	return wait
}
