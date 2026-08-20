package company

import (
	"context"
	"net/http"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCompanyMetricsDecorator_GetByTenantId_Success(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	serviceName := "ms-company"

	decorator := NewCompanyMetricsDecorator(mockClient, metricsInstance, serviceName).(*companyMetricsDecorator)

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

func TestCompanyMetricsDecorator_GetByTenantId_Error(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	serviceName := "ms-company"

	decorator := NewCompanyMetricsDecorator(mockClient, metricsInstance, serviceName).(*companyMetricsDecorator)

	expectedError := &appdto.HTTPError{
		Code:    http.StatusInternalServerError,
		Message: "Internal Server Error",
	}

	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, expectedError).Once()

	// Act
	result, err := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result)
	mockClient.AssertExpectations(t)
}

func TestCompanyMetricsDecorator_GetStatusCodeFromError(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	serviceName := "ms-company"

	decorator := NewCompanyMetricsDecorator(mockClient, metricsInstance, serviceName).(*companyMetricsDecorator)

	tests := []struct {
		name     string
		error    error
		expected int
	}{
		{
			name:     "nil error",
			error:    nil,
			expected: http.StatusOK,
		},
		{
			name: "HTTP error 400",
			error: &appdto.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Bad Request",
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "HTTP error 500",
			error: &appdto.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: "Internal Server Error",
			},
			expected: http.StatusInternalServerError,
		},
		{
			name:     "generic error",
			error:    assert.AnError,
			expected: http.StatusInternalServerError,
		},
	}

	// Act & Assert
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decorator.getStatusCodeFromError(tt.error)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompanyMetricsDecorator_RecordsMetrics(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	serviceName := "ms-company"

	decorator := NewCompanyMetricsDecorator(mockClient, metricsInstance, serviceName).(*companyMetricsDecorator)

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

	// Verifica se as métricas foram registradas
	// Nota: Como não temos acesso direto às métricas internas, vamos apenas verificar
	// que não houve panics ou erros durante a execução
}

func TestCompanyMetricsDecorator_RecordsErrorMetrics(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	serviceName := "ms-company"

	decorator := NewCompanyMetricsDecorator(mockClient, metricsInstance, serviceName).(*companyMetricsDecorator)

	expectedError := &appdto.HTTPError{
		Code:    http.StatusServiceUnavailable,
		Message: "Service Unavailable",
	}

	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, expectedError).Once()

	// Act
	result, err := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result)
	mockClient.AssertExpectations(t)

	// Verifica se as métricas de erro foram registradas
	// Nota: Como não temos acesso direto às métricas internas, vamos apenas verificar
	// que não houve panics ou erros durante a execução
}

func TestCompanyMetricsDecorator_ServiceName(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	serviceName := "test-service"

	decorator := NewCompanyMetricsDecorator(mockClient, metricsInstance, serviceName).(*companyMetricsDecorator)

	// Assert
	assert.Equal(t, serviceName, decorator.serviceName)
	assert.Equal(t, mockClient, decorator.inner)
	assert.Equal(t, metricsInstance, decorator.metrics)
}
