package user

import (
	"context"
	"errors"
	"testing"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	"github.com/keepguard/bff-core/internal/infrastructure/resilience"
	"github.com/stretchr/testify/assert"
)

// MockMetricsRecorder para teste
type MockMetricsRecorder struct{}

func (m *MockMetricsRecorder) SetCircuitBreakerState(service string, state int) {}

func TestNewCircuitBreakerDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	// Act
	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &circuitBreakerDecorator{}, decorator)
}

func TestCircuitBreakerDecorator_CreateUser_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := userDto.MSUserCreateRequestDTO{
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}
	xApplication := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserResponseDTO{
		ID:        "123",
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}

	mockInner.On("CreateUser", ctx, req, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.CreateUser(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_CreateUser_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := userDto.MSUserCreateRequestDTO{
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}
	xApplication := "test-app"
	correlationID := "corr-123"

	expectedError := errors.New("service unavailable")

	mockInner.On("CreateUser", ctx, req, xApplication, correlationID).Return(userDto.MSUserResponseDTO{}, expectedError)

	// Act
	result, err := decorator.CreateUser(ctx, req, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, userDto.MSUserResponseDTO{}, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_GetUserByCodeUser_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	codeUser := "code123"
	token := "token123"
	xApplication := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserResponseDTO{
		ID:        "123",
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}

	mockInner.On("GetUserByCodeUser", ctx, codeUser, token, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetUserByCodeUser(ctx, codeUser, token, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_GetByEmail_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	email := "test@example.com"
	xApplication := "test-app"
	correlationID := "corr-123"

	expectedResponse := authDto.UserByEmailResponseDTO{
		ID:    "123",
		Email: "test@example.com",
	}

	mockInner.On("GetByEmail", ctx, email, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetByEmail(ctx, email, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_CreateUserNotify_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := userDto.MSUserNotifyCreateRequestDTO{
		UserID:          "user123",
		EmailEnabled:    true,
		SmsEnabled:      true,
		PushEnabled:     true,
		WhatsAppEnabled: true,
	}
	xApplication := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserNotifyResponseDTO{
		ID:     "notify123",
		UserID: "user123",
	}

	mockInner.On("CreateUserNotify", ctx, req, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.CreateUserNotify(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_InitRegister_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := userDto.MSUserRegisterInitRequestDTO{
		Email:                      "test@example.com",
		NameFull:                   "Test User",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          false,
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
	}
	xApplication := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserRegisterInitResponseDTO{
		RegistrationSessionID: "session123",
		Email:                 "test@example.com",
		Token:                 "token123",
		ExpiresIn:             1800,
	}

	mockInner.On("InitRegister", ctx, req, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.InitRegister(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_ConfirmRegister_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	req := userDto.MSUserRegisterConfirmRequestDTO{
		Email:                 "test@example.com",
		RegistrationSessionID: "session123",
		Token:                 "123456",
	}
	xApplication := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserRegisterConfirmResponseDTO{
		Email:    "test@example.com",
		NameFull: "Test User",
		Phone:    "+5511999999999",
		Type:     "PERSON",
		Message:  "Registration confirmed",
	}

	mockInner.On("ConfirmRegister", ctx, req, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.ConfirmRegister(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_DeleteUser_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	userID := "user123"
	xApplication := "test-app"
	correlationID := "corr-123"

	mockInner.On("DeleteUser", ctx, userID, xApplication, correlationID).Return(nil)

	// Act
	err := decorator.DeleteUser(ctx, userID, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_DeleteUser_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	cbManager := resilience.NewCircuitBreakerManager(&MockMetricsRecorder{})
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	userID := "user123"
	xApplication := "test-app"
	correlationID := "corr-123"

	expectedError := errors.New("user not found")
	mockInner.On("DeleteUser", ctx, userID, xApplication, correlationID).Return(expectedError)

	// Act
	err := decorator.DeleteUser(ctx, userID, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)
}
