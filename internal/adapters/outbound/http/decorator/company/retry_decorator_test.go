package company

import (
	"context"
	"net/http"
	"testing"
	"time"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryDecorator_GetByTenantId_Success(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	expectedResponse := companyDto.MSCompanyResponseDTO{
		ID:   "123",
		Code: "TEST123",
		Name: "Test Company",
	}

	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestRetryDecorator_GetByTenantId_RetryableError_Success(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	expectedResponse := companyDto.MSCompanyResponseDTO{
		ID:   "123",
		Code: "TEST123",
		Name: "Test Company",
	}

	retryableError := &appdto.HTTPError{
		Code:    http.StatusServiceUnavailable,
		Message: "Service Unavailable",
	}

	// Primeira chamada falha, segunda sucede
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, retryableError).Once()
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestRetryDecorator_GetByTenantId_NonRetryableError(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	nonRetryableError := &appdto.HTTPError{
		Code:    http.StatusBadRequest,
		Message: "Bad Request",
	}

	// Deve chamar apenas uma vez (erro não retryable)
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, nonRetryableError).Once()

	// Act
	result, err := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, nonRetryableError, err)
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result)
	mockClient.AssertExpectations(t)
}

func TestRetryDecorator_GetByTenantId_MaxAttemptsExceeded(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	retryableError := &appdto.HTTPError{
		Code:    http.StatusServiceUnavailable,
		Message: "Service Unavailable",
	}

	// Deve chamar 2 vezes (MaxAttempts)
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, retryableError).Twice()

	// Act
	result, err := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, retryableError, err)
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result)
	mockClient.AssertExpectations(t)
}

func TestRetryDecorator_GetByTenantId_ContextCancelled(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond, // Delay maior para dar tempo de cancelar
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	retryableError := &appdto.HTTPError{
		Code:    http.StatusServiceUnavailable,
		Message: "Service Unavailable",
	}

	// Primeira chamada falha
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, retryableError).Once()

	// Cria contexto que será cancelado
	ctx, cancel := context.WithCancel(context.Background())

	// Act
	go func() {
		time.Sleep(50 * time.Millisecond) // Cancela após 50ms
		cancel()
	}()

	result, err := decorator.GetByTenantId(ctx, "test-app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result)
	mockClient.AssertExpectations(t)
}

func TestRetryDecorator_IsRetryableError(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := DefaultRetryConfig()
	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	tests := []struct {
		name     string
		error    error
		expected bool
	}{
		{
			name:     "nil error",
			error:    nil,
			expected: false,
		},
		{
			name: "retryable error - 429",
			error: &appdto.HTTPError{
				Code:    http.StatusTooManyRequests,
				Message: "Too Many Requests",
			},
			expected: true,
		},
		{
			name: "retryable error - 408",
			error: &appdto.HTTPError{
				Code:    http.StatusRequestTimeout,
				Message: "Request Timeout",
			},
			expected: true,
		},
		{
			name: "retryable error - 503",
			error: &appdto.HTTPError{
				Code:    http.StatusServiceUnavailable,
				Message: "Service Unavailable",
			},
			expected: true,
		},
		{
			name: "retryable error - 502",
			error: &appdto.HTTPError{
				Code:    http.StatusBadGateway,
				Message: "Bad Gateway",
			},
			expected: true,
		},
		{
			name: "retryable error - 504",
			error: &appdto.HTTPError{
				Code:    http.StatusGatewayTimeout,
				Message: "Gateway Timeout",
			},
			expected: true,
		},
		{
			name: "non-retryable error - 400",
			error: &appdto.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Bad Request",
			},
			expected: false,
		},
		{
			name: "non-retryable error - 401",
			error: &appdto.HTTPError{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized",
			},
			expected: false,
		},
		{
			name:     "generic error (network)",
			error:    assert.AnError,
			expected: true,
		},
	}

	// Act & Assert
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decorator.isRetryableError(tt.error)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryDecorator_CalculateDelay(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1000 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first retry",
			attempt:  0,
			expected: 100 * time.Millisecond,
		},
		{
			name:     "second retry",
			attempt:  1,
			expected: 200 * time.Millisecond,
		},
		{
			name:     "third retry",
			attempt:  2,
			expected: 400 * time.Millisecond,
		},
		{
			name:     "max delay exceeded",
			attempt:  10,
			expected: 1000 * time.Millisecond,
		},
	}

	// Act & Assert
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decorator.calculateDelay(tt.attempt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryDecorator_CalculateDelay_WithJitter(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1000 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	// Act
	delay1 := decorator.calculateDelay(1)
	delay2 := decorator.calculateDelay(1)

	// Assert
	// Com jitter, os delays devem ser diferentes (com alta probabilidade)
	// Mas devem estar dentro de uma faixa razoável
	baseDelay := 200 * time.Millisecond
	jitterRange := time.Duration(float64(baseDelay) * 0.25)

	assert.True(t, delay1 >= baseDelay-jitterRange)
	assert.True(t, delay1 <= baseDelay+jitterRange)
	assert.True(t, delay2 >= baseDelay-jitterRange)
	assert.True(t, delay2 <= baseDelay+jitterRange)
}

func TestDefaultRetryConfig(t *testing.T) {
	// Act
	config := DefaultRetryConfig()

	// Assert
	assert.Equal(t, 2, config.MaxAttempts)
	assert.Equal(t, 50*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 500*time.Millisecond, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.True(t, config.Jitter)
}

// TestRetryDecorator_IsRetryableError_4xxErrors_ShouldNotRetry testa que erros 4xx não fazem retry
func TestRetryDecorator_IsRetryableError_4xxErrors_ShouldNotRetry(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	// Test cases para erros 4xx que NÃO devem fazer retry
	testCases := []struct {
		name     string
		httpCode int
	}{
		{"BadRequest", http.StatusBadRequest},                   // 400
		{"Unauthorized", http.StatusUnauthorized},               // 401
		{"Forbidden", http.StatusForbidden},                     // 403
		{"NotFound", http.StatusNotFound},                       // 404
		{"Conflict", http.StatusConflict},                       // 409
		{"UnprocessableEntity", http.StatusUnprocessableEntity}, // 422
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			httpErr := &appdto.HTTPError{
				Code:    tc.httpCode,
				Message: "test error",
				Details: "test details",
			}

			// Act
			isRetryable := decorator.isRetryableError(httpErr)

			// Assert
			assert.False(t, isRetryable, "Erro %d (%s) não deveria ser retryable", tc.httpCode, tc.name)
		})
	}
}

// TestRetryDecorator_IsRetryableError_5xxErrors_ShouldRetry testa que erros 5xx fazem retry
func TestRetryDecorator_IsRetryableError_5xxErrors_ShouldRetry(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockClient, config).(*retryDecorator)

	// Test cases para erros 5xx que DEVEM fazer retry
	testCases := []struct {
		name     string
		httpCode int
	}{
		{"InternalServerError", http.StatusInternalServerError}, // 500
		{"BadGateway", http.StatusBadGateway},                   // 502
		{"ServiceUnavailable", http.StatusServiceUnavailable},   // 503
		{"GatewayTimeout", http.StatusGatewayTimeout},           // 504
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			httpErr := &appdto.HTTPError{
				Code:    tc.httpCode,
				Message: "test error",
				Details: "test details",
			}

			// Act
			isRetryable := decorator.isRetryableError(httpErr)

			// Assert
			assert.True(t, isRetryable, "Erro %d (%s) deveria ser retryable", tc.httpCode, tc.name)
		})
	}
}
