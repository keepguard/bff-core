package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	dto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricsRecorder é um mock para MetricsRecorder
type MockMetricsRecorder struct {
	mock.Mock
}

func (m *MockMetricsRecorder) SetCircuitBreakerState(service string, state int) {
	m.Called(service, state)
}

func TestNewCircuitBreakerManager(t *testing.T) {
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.breakers)
	assert.Equal(t, mockMetrics, manager.metrics)
}

func TestCircuitBreakerManager_GetOrCreate_NewBreaker(t *testing.T) {
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
	}

	breaker := manager.GetOrCreate("test-service", config)

	assert.NotNil(t, breaker)
	assert.Equal(t, "test-service", breaker.Name())
	assert.Contains(t, manager.breakers, "test-service")
}

func TestDefaultIsSuccessful(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error is successful",
			err:      nil,
			expected: true,
		},
		{
			name:     "HTTPError 400 Bad Request is considered success (business validation)",
			err:      &dto.HTTPError{Code: 400, Message: "bad request"},
			expected: true,
		},
		{
			name:     "HTTPError 401 Unauthorized is considered success (invalid credentials)",
			err:      &dto.HTTPError{Code: 401, Message: "unauthorized"},
			expected: true,
		},
		{
			name:     "HTTPError 404 Not Found is considered success",
			err:      &dto.HTTPError{Code: 404, Message: "not found"},
			expected: true,
		},
		{
			name:     "AppError 400 is considered success",
			err:      &pkg.AppError{StatusCode: 400, Message: "user not active"},
			expected: true,
		},
		{
			name:     "HTTPError 500 Internal Server Error is considered failure",
			err:      &dto.HTTPError{Code: 500, Message: "internal error"},
			expected: false,
		},
		{
			name:     "HTTPError 503 Service Unavailable is considered failure",
			err:      &dto.HTTPError{Code: 503, Message: "service unavailable"},
			expected: false,
		},
		{
			name:     "AppError 500 is considered failure",
			err:      &pkg.AppError{StatusCode: 500, Message: "server error"},
			expected: false,
		},
		{
			name:     "Generic error is considered failure",
			err:      errors.New("connection refused"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DefaultIsSuccessful(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCircuitBreakerManager_Execute_Success(t *testing.T) {
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name: "test-service",
	}
	manager.GetOrCreate("test-service", config)

	result, err := manager.Execute(context.Background(), "test-service", func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}
