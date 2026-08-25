package communication

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"

	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
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

// DefaultRetryConfig retorna configuração padrão para CommunicationClient
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  4,                      // 4 tentativas (APIs de e-mail são instáveis)
		InitialDelay: 200 * time.Millisecond, // Mais lento (API externa)
		MaxDelay:     10 * time.Second,       // Aguenta esperar mais
		Multiplier:   2.5,                    // Backoff agressivo
		Jitter:       true,
	}
}

// retryDecorator implementa retry para CommunicationClient
type retryDecorator struct {
	inner  portsclient.CommunicationClient
	config RetryConfig
}

// NewRetryDecorator cria um decorator de retry para CommunicationClient
func NewRetryDecorator(
	inner portsclient.CommunicationClient,
	config RetryConfig,
) portsclient.CommunicationClient {
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
			http.StatusGatewayTimeout,     // 504
			http.StatusRequestTimeout,     // 408
			http.StatusTooManyRequests:    // 429
			return true
		}
		// NUNCA retry para 4xx de validação/negócio (400, 401, 403, 404, 422)
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
func (d *retryDecorator) retry(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := operation()

		if err == nil {
			return nil
		}

		lastErr = err

		if !d.isRetryableError(err) || attempt == d.config.MaxAttempts-1 {
			return lastErr
		}

		delay := d.calculateDelay(attempt)

		select {
		case <-time.After(delay):
			// Continua para próxima tentativa
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// SendNotification implementa SendNotification com retry
func (d *retryDecorator) SendNotification(ctx context.Context, req portsclient.SendNotificationRequestDTO, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.SendNotification(ctx, req, tenantId, correlationID)
	})
}

// SendMessage implementa SendMessage com retry
func (d *retryDecorator) SendMessage(ctx context.Context, req communicationDto.SendMessageRequestDTO, tenantId, correlationID string) (communicationDto.SendMessageResponseDTO, error) {
	var response communicationDto.SendMessageResponseDTO
	var lastErr error

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return communicationDto.SendMessageResponseDTO{}, ctx.Err()
		default:
		}

		resp, err := d.inner.SendMessage(ctx, req, tenantId, correlationID)
		response = resp

		if err == nil {
			return response, nil
		}

		lastErr = err

		if !d.isRetryableError(err) || attempt == d.config.MaxAttempts-1 {
			return response, lastErr
		}

		delay := d.calculateDelay(attempt)

		select {
		case <-time.After(delay):
			// Continua para próxima tentativa
		case <-ctx.Done():
			return communicationDto.SendMessageResponseDTO{}, ctx.Err()
		}
	}

	return response, lastErr
}
