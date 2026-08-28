package user

import (
	"context"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/resilience"
)

// circuitBreakerDecorator implementa circuit breaker para UserClient
type circuitBreakerDecorator struct {
	inner          portsclient.UserClient
	circuitBreaker *resilience.CircuitBreakerManager
	serviceName    string
}

// NewCircuitBreakerDecorator cria um decorator de circuit breaker para UserClient
func NewCircuitBreakerDecorator(
	inner portsclient.UserClient,
	cbManager *resilience.CircuitBreakerManager,
	serviceName string,
) portsclient.UserClient {
	return &circuitBreakerDecorator{
		inner:          inner,
		circuitBreaker: cbManager,
		serviceName:    serviceName,
	}
}

// CreateUser implementa CreateUser com circuit breaker
func (d *circuitBreakerDecorator) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.CreateUser(ctx, req, tenantId, correlationID)
	})

	if err != nil {
		return userDto.MSUserResponseDTO{}, err
	}

	return result.(userDto.MSUserResponseDTO), nil
}

// GetUserByCodeUser implementa GetUserByCodeUser com circuit breaker
func (d *circuitBreakerDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
	})

	if err != nil {
		return userDto.MSUserResponseDTO{}, err
	}

	return result.(userDto.MSUserResponseDTO), nil
}

// GetByEmail implementa GetByEmail com circuit breaker
func (d *circuitBreakerDecorator) GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.GetByEmail(ctx, email, tenantId, companyId, correlationID)
	})

	if err != nil {
		return authDto.UserByEmailResponseDTO{}, err
	}

	return result.(authDto.UserByEmailResponseDTO), nil
}

// CreateUserNotify implementa CreateUserNotify com circuit breaker
func (d *circuitBreakerDecorator) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.CreateUserNotify(ctx, req, tenantId, correlationID)
	})

	if err != nil {
		return userDto.MSUserNotifyResponseDTO{}, err
	}

	return result.(userDto.MSUserNotifyResponseDTO), nil
}

// InitRegister implementa InitRegister com circuit breaker
func (d *circuitBreakerDecorator) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.InitRegister(ctx, req, tenantId, correlationID)
	})

	if err != nil {
		return userDto.MSUserRegisterInitResponseDTO{}, err
	}

	return result.(userDto.MSUserRegisterInitResponseDTO), nil
}

// ConfirmRegister implementa ConfirmRegister com circuit breaker
func (d *circuitBreakerDecorator) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.ConfirmRegister(ctx, req, tenantId, correlationID)
	})

	if err != nil {
		return userDto.MSUserRegisterConfirmResponseDTO{}, err
	}

	return result.(userDto.MSUserRegisterConfirmResponseDTO), nil
}

// DeleteUser implementa DeleteUser com circuit breaker
func (d *circuitBreakerDecorator) DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.DeleteUser(ctx, userID, tenantId, correlationID)
	})

	return err
}

// ResendRegisterToken implementa ResendRegisterToken com circuit breaker
func (d *circuitBreakerDecorator) ResendRegisterToken(ctx context.Context, req userDto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.ResendRegisterToken(ctx, req, tenantId, correlationID)
	})

	if err != nil {
		return userDto.MSUserRegisterResendResponseDTO{}, err
	}

	return result.(userDto.MSUserRegisterResendResponseDTO), nil
}
