package user

import (
	"context"
	"net/http"
	"testing"

	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserMetricsDecorator_CreateUser_Success(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	serviceName := "ms-user"

	decorator := NewUserMetricsDecorator(mockClient, metricsInstance, serviceName).(*userMetricsDecorator)

	req := userDto.MSUserCreateRequestDTO{
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}

	expectedResponse := userDto.MSUserResponseDTO{
		ID:        "123",
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}

	mockClient.On("CreateUser", mock.Anything, req, "app", "corr-123").Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.CreateUser(context.Background(), req, "app", "corr-123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestUserMetricsDecorator_CreateUser_Error(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	serviceName := "ms-user"

	decorator := NewUserMetricsDecorator(mockClient, metricsInstance, serviceName).(*userMetricsDecorator)

	req := userDto.MSUserCreateRequestDTO{
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}

	expectedError := &appdto.HTTPError{
		Code:    http.StatusInternalServerError,
		Message: "Internal Server Error",
	}

	mockClient.On("CreateUser", mock.Anything, req, "app", "corr-123").Return(userDto.MSUserResponseDTO{}, expectedError).Once()

	// Act
	result, err := decorator.CreateUser(context.Background(), req, "app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, userDto.MSUserResponseDTO{}, result)
	mockClient.AssertExpectations(t)
}

func TestUserMetricsDecorator_GetStatusCodeFromError(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	serviceName := "ms-user"

	decorator := NewUserMetricsDecorator(mockClient, metricsInstance, serviceName).(*userMetricsDecorator)

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

func TestUserMetricsDecorator_ServiceName(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	serviceName := "test-service"

	decorator := NewUserMetricsDecorator(mockClient, metricsInstance, serviceName).(*userMetricsDecorator)

	// Assert
	assert.Equal(t, serviceName, decorator.serviceName)
	assert.Equal(t, mockClient, decorator.inner)
	assert.Equal(t, metricsInstance, decorator.metrics)
}
