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
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserResponseDTO{
		ID:        "123",
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}

	mockInner.On("CreateUser", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.CreateUser(ctx, req, tenantId, correlationID)

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
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedError := errors.New("service unavailable")

	mockInner.On("CreateUser", ctx, req, tenantId, correlationID).Return(userDto.MSUserResponseDTO{}, expectedError)

	// Act
	result, err := decorator.CreateUser(ctx, req, tenantId, correlationID)

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
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserResponseDTO{
		ID:        "123",
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}

	mockInner.On("GetUserByCodeUser", ctx, codeUser, token, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)

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
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedResponse := authDto.UserByEmailResponseDTO{
		ID:    "123",
		Email: "test@example.com",
	}

	mockInner.On("GetByEmail", ctx, email, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetByEmail(ctx, email, tenantId, correlationID)

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
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserNotifyResponseDTO{
		ID:     "notify123",
		UserID: "user123",
	}

	mockInner.On("CreateUserNotify", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.CreateUserNotify(ctx, req, tenantId, correlationID)

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
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserRegisterInitResponseDTO{
		RegistrationSessionID: "session123",
		Email:                 "test@example.com",
		Token:                 "token123",
		ExpiresIn:             1800,
	}

	mockInner.On("InitRegister", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.InitRegister(ctx, req, tenantId, correlationID)

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
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedResponse := userDto.MSUserRegisterConfirmResponseDTO{
		Email:    "test@example.com",
		NameFull: "Test User",
		Phone:    "+5511999999999",
		Type:     "PERSON",
		Message:  "Registration confirmed",
	}

	mockInner.On("ConfirmRegister", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.ConfirmRegister(ctx, req, tenantId, correlationID)

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
	tenantId := "test-app"
	correlationID := "corr-123"

	mockInner.On("DeleteUser", ctx, userID, tenantId, correlationID).Return(nil)

	// Act
	err := decorator.DeleteUser(ctx, userID, tenantId, correlationID)

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
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedError := errors.New("user not found")
	mockInner.On("DeleteUser", ctx, userID, tenantId, correlationID).Return(expectedError)

	// Act
	err := decorator.DeleteUser(ctx, userID, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)
}
