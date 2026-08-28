package company

import (
	"context"
	"testing"
	"time"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	"github.com/keepguard/bff-core/internal/infrastructure/resilience"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCircuitBreakerDecorator_GetByTenantId_Success(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	cbManager := resilience.NewCircuitBreakerManager(metricsInstance)

	// Configura circuit breaker
	config := resilience.CircuitBreakerConfig{
		Name:        "test-company",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= 2 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5
		},
	}
	cbManager.GetOrCreate("test-company", config)

	decorator := NewCircuitBreakerDecorator(mockClient, cbManager, "test-company").(*circuitBreakerDecorator)

	expectedResponse := companyDto.MSCompanyResponseDTO{
		ID:   "123",
		CodeCompany: "TEST123",
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

func TestCircuitBreakerDecorator_GetByTenantId_Error(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	cbManager := resilience.NewCircuitBreakerManager(metricsInstance)

	// Configura circuit breaker
	config := resilience.CircuitBreakerConfig{
		Name:        "test-company",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= 2 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5
		},
	}
	cbManager.GetOrCreate("test-company", config)

	decorator := NewCircuitBreakerDecorator(mockClient, cbManager, "test-company").(*circuitBreakerDecorator)

	expectedError := assert.AnError

	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, expectedError).Once()

	// Act
	result, err := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result)
	mockClient.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_GetByTenantId_CircuitOpen(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	cbManager := resilience.NewCircuitBreakerManager(metricsInstance)

	// Configura circuit breaker com threshold baixo para facilitar abertura
	config := resilience.CircuitBreakerConfig{
		Name:        "test-company",
		MaxRequests: 1,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= 1 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5
		},
	}
	cbManager.GetOrCreate("test-company", config)

	decorator := NewCircuitBreakerDecorator(mockClient, cbManager, "test-company").(*circuitBreakerDecorator)

	// Primeira chamada falha para abrir o circuit
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, assert.AnError).Once()

	// Act
	// Primeira chamada - deve falhar e abrir o circuit
	result1, err1 := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")
	assert.Error(t, err1)

	// Segunda chamada - circuit deve estar aberto
	result2, err2 := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")
	assert.Error(t, err2)

	// Assert
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result1)
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result2)
	mockClient.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_ServiceName(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	cbManager := resilience.NewCircuitBreakerManager(metricsInstance)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockClient, cbManager, serviceName).(*circuitBreakerDecorator)

	// Assert
	assert.Equal(t, serviceName, decorator.serviceName)
	assert.Equal(t, mockClient, decorator.inner)
	assert.Equal(t, cbManager, decorator.circuitBreaker)
}
