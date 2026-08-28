package user

import (
	"context"
	"net/http"
	"time"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
)

// userMetricsDecorator implementa métricas para UserClient
type userMetricsDecorator struct {
	inner       portsclient.UserClient
	metrics     *metrics.Metrics
	serviceName string
}

// NewUserMetricsDecorator cria um decorator de métricas para UserClient
func NewUserMetricsDecorator(
	inner portsclient.UserClient,
	metrics *metrics.Metrics,
	serviceName string,
) portsclient.UserClient {
	return &userMetricsDecorator{
		inner:       inner,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

// getStatusCodeFromError extrai o status code de um erro
func (d *userMetricsDecorator) getStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.Code
	}

	return http.StatusInternalServerError
}

// CreateUser implementa CreateUser com métricas
func (d *userMetricsDecorator) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.CreateUser(ctx, req, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/users", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/users", errorType)
	}

	return response, err
}

// GetUserByCodeUser implementa GetUserByCodeUser com métricas
func (d *userMetricsDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/users/by-code", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "GET", "/users/by-code", errorType)
	}

	return response, err
}

// GetByEmail implementa GetByEmail com métricas
func (d *userMetricsDecorator) GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.GetByEmail(ctx, email, tenantId, companyId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/users/email", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "GET", "/users/email", errorType)
	}

	return response, err
}

// CreateUserNotify implementa CreateUserNotify com métricas
func (d *userMetricsDecorator) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.CreateUserNotify(ctx, req, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/users/notify", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/users/notify", errorType)
	}

	return response, err
}

// InitRegister implementa InitRegister com métricas
func (d *userMetricsDecorator) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.InitRegister(ctx, req, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/register/init", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/register/init", errorType)
	}

	return response, err
}

// ConfirmRegister implementa ConfirmRegister com métricas
func (d *userMetricsDecorator) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.ConfirmRegister(ctx, req, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/register/confirm", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/register/confirm", errorType)
	}

	return response, err
}

// DeleteUser implementa DeleteUser com métricas
func (d *userMetricsDecorator) DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error {
	start := time.Now()

	err := d.inner.DeleteUser(ctx, userID, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "DELETE", "/users", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "DELETE", "/users", errorType)
	}

	return err
}

// ResendRegisterToken implementa ResendRegisterToken com métricas
func (d *userMetricsDecorator) ResendRegisterToken(ctx context.Context, req userDto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.ResendRegisterToken(ctx, req, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := 200
	if err != nil {
		statusCode = 500
	}

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/register/resend", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/register/resend", errorType)
	}

	return response, err
}
