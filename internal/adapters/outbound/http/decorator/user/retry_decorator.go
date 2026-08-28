package user

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
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

// DefaultRetryConfig retorna configuração padrão para UserClient
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,                      // 3 tentativas para operações de usuário
		InitialDelay: 100 * time.Millisecond, // Delay inicial de 100ms
		MaxDelay:     2 * time.Second,        // Delay máximo de 2s
		Multiplier:   2.0,                    // Backoff exponencial: 100ms → 200ms → 400ms
		Jitter:       true,                   // Adiciona aleatoriedade
	}
}

// retryDecorator implementa retry para UserClient
type retryDecorator struct {
	inner  portsclient.UserClient
	config RetryConfig
}

// NewRetryDecorator cria um decorator de retry para UserClient
func NewRetryDecorator(
	inner portsclient.UserClient,
	config RetryConfig,
) portsclient.UserClient {
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
func (d *retryDecorator) retry(ctx context.Context, operation func() (userDto.MSUserResponseDTO, error)) (userDto.MSUserResponseDTO, error) {
	var lastErr error
	var lastResult userDto.MSUserResponseDTO

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return userDto.MSUserResponseDTO{}, ctx.Err()
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
			return userDto.MSUserResponseDTO{}, ctx.Err()
		}
	}

	return lastResult, lastErr
}

// retryError executa uma função que retorna apenas erro com retry
func (d *retryDecorator) retryError(ctx context.Context, operation func() error) error {
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

// CreateUser implementa CreateUser com retry
func (d *retryDecorator) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	return d.retry(ctx, func() (userDto.MSUserResponseDTO, error) {
		return d.inner.CreateUser(ctx, req, tenantId, correlationID)
	})
}

// GetUserByCodeUser implementa GetUserByCodeUser com retry
func (d *retryDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	return d.retry(ctx, func() (userDto.MSUserResponseDTO, error) {
		return d.inner.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
	})
}

// GetByEmail implementa GetByEmail com retry
func (d *retryDecorator) GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	var lastErr error
	var lastResult authDto.UserByEmailResponseDTO

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return authDto.UserByEmailResponseDTO{}, ctx.Err()
		default:
		}

		result, err := d.inner.GetByEmail(ctx, email, tenantId, companyId, correlationID)

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
			return authDto.UserByEmailResponseDTO{}, ctx.Err()
		}
	}

	return lastResult, lastErr
}

// CreateUserNotify implementa CreateUserNotify com retry
func (d *retryDecorator) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	var lastErr error
	var lastResult userDto.MSUserNotifyResponseDTO

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return userDto.MSUserNotifyResponseDTO{}, ctx.Err()
		default:
		}

		result, err := d.inner.CreateUserNotify(ctx, req, tenantId, correlationID)

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
			return userDto.MSUserNotifyResponseDTO{}, ctx.Err()
		}
	}

	return lastResult, lastErr
}

// InitRegister implementa InitRegister com retry
func (d *retryDecorator) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	var lastErr error
	var lastResult userDto.MSUserRegisterInitResponseDTO

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return userDto.MSUserRegisterInitResponseDTO{}, ctx.Err()
		default:
		}

		result, err := d.inner.InitRegister(ctx, req, tenantId, correlationID)

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
			return userDto.MSUserRegisterInitResponseDTO{}, ctx.Err()
		}
	}

	return lastResult, lastErr
}

// ConfirmRegister implementa ConfirmRegister com retry
func (d *retryDecorator) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	var lastResult userDto.MSUserRegisterConfirmResponseDTO
	var lastErr error

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := d.calculateDelay(attempt - 1)
			select {
			case <-ctx.Done():
				return lastResult, ctx.Err()
			case <-time.After(delay):
			}
		}

		lastResult, lastErr = d.inner.ConfirmRegister(ctx, req, tenantId, correlationID)
		if lastErr == nil {
			return lastResult, nil
		}

		// Verifica se o erro é retryable
		if !d.isRetryableError(lastErr) {
			return lastResult, lastErr
		}
	}

	return lastResult, lastErr
}

// DeleteUser implementa DeleteUser com retry
func (d *retryDecorator) DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error {
	return d.retryError(ctx, func() error {
		return d.inner.DeleteUser(ctx, userID, tenantId, correlationID)
	})
}

// ResendRegisterToken implementa ResendRegisterToken com retry
func (d *retryDecorator) ResendRegisterToken(ctx context.Context, req userDto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error) {
	return d.retryResend(ctx, func() (userDto.MSUserRegisterResendResponseDTO, error) {
		return d.inner.ResendRegisterToken(ctx, req, tenantId, correlationID)
	})
}

// retryResend executa uma função com retry para ResendRegisterToken
func (d *retryDecorator) retryResend(ctx context.Context, operation func() (userDto.MSUserRegisterResendResponseDTO, error)) (userDto.MSUserRegisterResendResponseDTO, error) {
	var lastResult userDto.MSUserRegisterResendResponseDTO
	var lastErr error

	for attempt := 0; attempt <= d.config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := d.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return lastResult, ctx.Err()
			case <-time.After(delay):
			}
		}

		lastResult, lastErr = operation()
		if lastErr == nil {
			return lastResult, nil
		}

		if !d.isRetryableError(lastErr) {
			return lastResult, lastErr
		}
	}

	return lastResult, lastErr
}
