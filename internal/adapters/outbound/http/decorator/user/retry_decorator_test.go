package user

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
)

func TestNewRetryDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := DefaultRetryConfig()

	// Act
	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &retryDecorator{}, decorator)
}

func TestDefaultRetryConfig(t *testing.T) {
	// Act
	config := DefaultRetryConfig()

	// Assert
	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 2*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.True(t, config.Jitter)
}

func TestRetryDecorator_CreateUser_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

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

	mockInner.On("CreateUser", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.CreateUser(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_CreateUser_RetryableError_EventuallySuccess(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

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

	retryableError := &appdto.HTTPError{
		Code:    http.StatusServiceUnavailable,
		Message: "Service temporarily unavailable",
	}

	// First two calls fail with retryable error, third succeeds
	mockInner.On("CreateUser", ctx, req, tenantId, correlationID).Return(userDto.MSUserResponseDTO{}, retryableError).Twice()
	mockInner.On("CreateUser", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.CreateUser(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_CreateUser_NonRetryableError_NoRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	ctx := context.Background()
	req := userDto.MSUserCreateRequestDTO{
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}
	tenantId := "test-app"
	correlationID := "corr-123"

	nonRetryableError := &appdto.HTTPError{
		Code:    http.StatusBadRequest,
		Message: "Invalid request",
	}

	// Should only be called once for non-retryable error
	mockInner.On("CreateUser", ctx, req, tenantId, correlationID).Return(userDto.MSUserResponseDTO{}, nonRetryableError).Once()

	// Act
	result, err := decorator.CreateUser(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, nonRetryableError, err)
	assert.Equal(t, userDto.MSUserResponseDTO{}, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_CreateUser_AllAttemptsFail(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	ctx := context.Background()
	req := userDto.MSUserCreateRequestDTO{
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}
	tenantId := "test-app"
	correlationID := "corr-123"

	retryableError := &appdto.HTTPError{
		Code:    http.StatusServiceUnavailable,
		Message: "Service temporarily unavailable",
	}

	// All attempts fail
	mockInner.On("CreateUser", ctx, req, tenantId, correlationID).Return(userDto.MSUserResponseDTO{}, retryableError).Twice()

	// Act
	result, err := decorator.CreateUser(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, retryableError, err)
	assert.Equal(t, userDto.MSUserResponseDTO{}, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_GetUserByCodeUser_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

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

	mockInner.On("GetUserByCodeUser", ctx, codeUser, token, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_GetByEmail_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	ctx := context.Background()
	email := "test@example.com"
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedResponse := authDto.UserByEmailResponseDTO{
		ID:    "123",
		Email: "test@example.com",
	}

	mockInner.On("GetByEmail", ctx, email, tenantId, "company-123", correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.GetByEmail(ctx, email, tenantId, "company-123", correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_CreateUserNotify_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

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

	mockInner.On("CreateUserNotify", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.CreateUserNotify(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_InitRegister_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

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

	mockInner.On("InitRegister", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.InitRegister(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_ConfirmRegister_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

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

	mockInner.On("ConfirmRegister", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.ConfirmRegister(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_DeleteUser_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	ctx := context.Background()
	userID := "user123"
	tenantId := "test-app"
	correlationID := "corr-123"

	mockInner.On("DeleteUser", ctx, userID, tenantId, correlationID).Return(nil).Once()

	// Act
	err := decorator.DeleteUser(ctx, userID, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_DeleteUser_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	ctx := context.Background()
	userID := "user123"
	tenantId := "test-app"
	correlationID := "corr-123"

	expectedError := errors.New("user not found")
	mockInner.On("DeleteUser", ctx, userID, tenantId, correlationID).Return(expectedError).Once()

	// Act
	err := decorator.DeleteUser(ctx, userID, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_ContextCancellation(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	ctx, cancel := context.WithCancel(context.Background())
	req := userDto.MSUserCreateRequestDTO{
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}
	tenantId := "test-app"
	correlationID := "corr-123"

	retryableError := &appdto.HTTPError{
		Code:    http.StatusServiceUnavailable,
		Message: "Service temporarily unavailable",
	}

	mockInner.On("CreateUser", ctx, req, tenantId, correlationID).Return(userDto.MSUserResponseDTO{}, retryableError).Once()

	// Cancel context after first call
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	// Act
	result, err := decorator.CreateUser(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, userDto.MSUserResponseDTO{}, result)
	mockInner.AssertExpectations(t)
}

// TestRetryDecorator_IsRetryableError_4xxErrors_ShouldNotRetry testa que erros 4xx não fazem retry
func TestRetryDecorator_IsRetryableError_4xxErrors_ShouldNotRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

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
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

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

// TestRetryDecorator_IsRetryableError_NetworkErrors_ShouldRetry testa que erros de rede fazem retry
func TestRetryDecorator_IsRetryableError_NetworkErrors_ShouldRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	// Test cases para erros de rede que DEVEM fazer retry
	testCases := []struct {
		name string
		err  error
	}{
		{"NetworkError", errors.New("network error")},
		{"TimeoutError", errors.New("timeout")},
		{"ConnectionError", errors.New("connection refused")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			isRetryable := decorator.isRetryableError(tc.err)

			// Assert
			assert.True(t, isRetryable, "Erro de rede (%s) deveria ser retryable", tc.name)
		})
	}
}

// TestRetryDecorator_ConfirmRegister_4xxError_ShouldNotRetry testa que ConfirmRegister com erro 4xx não faz retry
func TestRetryDecorator_ConfirmRegister_4xxError_ShouldNotRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	ctx := context.Background()
	req := userDto.MSUserRegisterConfirmRequestDTO{
		Email:                 "test@example.com",
		RegistrationSessionID: "session123",
		Token:                 "123456",
	}
	tenantId := "test-app"
	correlationID := "corr-123"

	// Erro 400 (Bad Request) - token inválido
	httpErr := &appdto.HTTPError{
		Code:    http.StatusBadRequest,
		Message: "token inválido",
		Details: "token não encontrado ou expirado",
	}

	mockInner.On("ConfirmRegister", ctx, req, tenantId, correlationID).Return(userDto.MSUserRegisterConfirmResponseDTO{}, httpErr).Once()

	// Act
	result, err := decorator.ConfirmRegister(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, httpErr, err)
	assert.Equal(t, userDto.MSUserRegisterConfirmResponseDTO{}, result)

	// Verificar que foi chamado apenas UMA vez (sem retry)
	mockInner.AssertExpectations(t)
}

// TestRetryDecorator_ConfirmRegister_5xxError_ShouldRetry testa que ConfirmRegister com erro 5xx faz retry
func TestRetryDecorator_ConfirmRegister_5xxError_ShouldRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config).(*retryDecorator)

	ctx := context.Background()
	req := userDto.MSUserRegisterConfirmRequestDTO{
		Email:                 "test@example.com",
		RegistrationSessionID: "session123",
		Token:                 "123456",
	}
	tenantId := "test-app"
	correlationID := "corr-123"

	// Erro 500 (Internal Server Error) - erro de infraestrutura
	httpErr := &appdto.HTTPError{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
		Details: "database connection failed",
	}

	// Mock deve ser chamado 3 vezes (MaxAttempts)
	mockInner.On("ConfirmRegister", ctx, req, tenantId, correlationID).Return(userDto.MSUserRegisterConfirmResponseDTO{}, httpErr).Times(3)

	// Act
	result, err := decorator.ConfirmRegister(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, httpErr, err)
	assert.Equal(t, userDto.MSUserRegisterConfirmResponseDTO{}, result)

	// Verificar que foi chamado 3 vezes (com retry)
	mockInner.AssertExpectations(t)
}
