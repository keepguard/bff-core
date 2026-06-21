package communication

import (
	"context"
	"net/http"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// getTestMetrics retorna uma instância de métricas para testes
func getTestMetrics() *metrics.Metrics {
	if sharedMetrics == nil {
		sharedMetrics = metrics.New()
	}
	return sharedMetrics
}

var (
	// sharedMetrics é uma instância compartilhada de métricas para os testes
	sharedMetrics *metrics.Metrics
)

func TestNewCommunicationMetricsDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	metricsInstance := getTestMetrics()
	serviceName := "test-service"

	// Act
	decorator := NewCommunicationMetricsDecorator(mockInner, metricsInstance, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &communicationMetricsDecorator{}, decorator)
}

func TestCommunicationMetricsDecorator_SendNotification_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	metricsInstance := getTestMetrics()
	serviceName := "ms-communication"

	decorator := NewCommunicationMetricsDecorator(mockInner, metricsInstance, serviceName)

	req := portsclient.SendNotificationRequestDTO{
		UserID:       "user-123",
		TemplateType: "WELCOME_EMAIL",
		Channel:      "email",
		Recipient:    "user@example.com",
		Data:         map[string]string{"name": "João"},
	}

	mockInner.On("SendNotification", mock.Anything, req, "app", "corr-123").Return(nil).Once()

	// Act
	err := decorator.SendNotification(context.Background(), req, "app", "corr-123")

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestCommunicationMetricsDecorator_SendNotification_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	metricsInstance := getTestMetrics()
	serviceName := "ms-communication"

	decorator := NewCommunicationMetricsDecorator(mockInner, metricsInstance, serviceName)

	req := portsclient.SendNotificationRequestDTO{
		UserID:       "user-123",
		TemplateType: "INVALID_TEMPLATE",
		Channel:      "email",
		Recipient:    "invalid@example.com",
		Data:         map[string]string{},
	}

	expectedError := &appdto.HTTPError{
		Code:    http.StatusInternalServerError,
		Message: "Internal Server Error",
	}

	mockInner.On("SendNotification", mock.Anything, req, "app", "corr-123").Return(expectedError).Once()

	// Act
	err := decorator.SendNotification(context.Background(), req, "app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)
}

func TestCommunicationMetricsDecorator_GetStatusCodeFromError(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	metricsInstance := getTestMetrics()
	serviceName := "ms-communication"

	decorator := NewCommunicationMetricsDecorator(mockInner, metricsInstance, serviceName).(*communicationMetricsDecorator)

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

func TestCommunicationMetricsDecorator_ServiceName(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	metricsInstance := getTestMetrics()
	serviceName := "test-service"

	decorator := NewCommunicationMetricsDecorator(mockInner, metricsInstance, serviceName).(*communicationMetricsDecorator)

	// Assert
	assert.Equal(t, serviceName, decorator.serviceName)
	assert.Equal(t, mockInner, decorator.inner)
	assert.Equal(t, metricsInstance, decorator.metrics)
}
