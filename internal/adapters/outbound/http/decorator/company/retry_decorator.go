package company

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
)

// RetryConfig configuração para retry
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// DefaultRetryConfig retorna configuração padrão para CompanyClient
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  2,                     // Apenas 2 tentativas (com cache, raramente usado)
		InitialDelay: 50 * time.Millisecond, // Muito rápido (chamado em toda req)
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// retryDecorator implementa retry para CompanyClient
type retryDecorator struct {
	inner  portsclient.CompanyClient
	config RetryConfig
}

// NewRetryDecorator cria um decorator de retry para CompanyClient
func NewRetryDecorator(
	inner portsclient.CompanyClient,
	config RetryConfig,
) portsclient.CompanyClient {
	return &retryDecorator{
		inner:  inner,
		config: config,
	}
}

// isRetryableError verifica se um erro deve ser retentado
func (d *retryDecorator) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		// Retry APENAS para erros 5xx (infraestrutura)
		switch httpErr.Code {
		case http.StatusInternalServerError, // 500
			http.StatusBadGateway,         // 502
			http.StatusServiceUnavailable, // 503
			http.StatusGatewayTimeout:     // 504
			return true
		}
		// NUNCA retry para 4xx (erros de negócio/validação)
		return false
	}

	// Erros de rede/timeout são retryable (infraestrutura)
	return true
}

// calculateDelay calcula o delay com backoff exponencial
func (d *retryDecorator) calculateDelay(attempt int) time.Duration {
	delay := float64(d.config.InitialDelay) * math.Pow(d.config.Multiplier, float64(attempt))

	if delay > float64(d.config.MaxDelay) {
		delay = float64(d.config.MaxDelay)
	}

	if d.config.Jitter {
		jitterRange := delay * 0.25
		jitter := (rand.Float64() * 2 * jitterRange) - jitterRange
		delay += jitter

		if delay < 0 {
			delay = float64(d.config.InitialDelay)
		}
	}

	return time.Duration(delay)
}

// retry executa uma função com retry
func (d *retryDecorator) retry(ctx context.Context, operation func() (companyDto.MSCompanyResponseDTO, error)) (companyDto.MSCompanyResponseDTO, error) {
	var lastErr error
	var lastResult companyDto.MSCompanyResponseDTO

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return companyDto.MSCompanyResponseDTO{}, ctx.Err()
		default:
		}

		result, err := operation()

		if err == nil {
			return result, nil
		}

		lastErr = err
		lastResult = result

		if !d.isRetryableError(err) || attempt == d.config.MaxAttempts-1 {
			return lastResult, lastErr
		}

		delay := d.calculateDelay(attempt)

		select {
		case <-time.After(delay):
			// Continua para próxima tentativa
		case <-ctx.Done():
			return companyDto.MSCompanyResponseDTO{}, ctx.Err()
		}
	}

	return lastResult, lastErr
}

// GetByXApplication implementa GetByXApplication com retry
func (d *retryDecorator) GetByXApplication(ctx context.Context, xApplication, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	return d.retry(ctx, func() (companyDto.MSCompanyResponseDTO, error) {
		return d.inner.GetByXApplication(ctx, xApplication, correlationID)
	})
}
