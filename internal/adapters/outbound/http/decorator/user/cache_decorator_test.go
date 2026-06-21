package user

import (
	"context"
	"testing"
	"time"

	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserCacheDecorator_GetUserByCodeUser_CacheHit(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	expectedResponse := userDto.MSUserResponseDTO{
		ID:       "123",
		CodeUser: "user-123",
		Email:    "test@example.com",
	}

	// Primeira chamada - deve ir para o mock
	mockClient.On("GetUserByCodeUser", mock.Anything, "user-123", "token", "app", "corr-123").Return(expectedResponse, nil).Once()

	// Act & Assert
	// Primeira chamada - cache miss
	result1, err1 := decorator.GetUserByCodeUser(context.Background(), "user-123", "token", "app", "corr-123")
	assert.NoError(t, err1)
	assert.Equal(t, expectedResponse, result1)

	// Segunda chamada - cache hit (não deve chamar o mock novamente)
	result2, err2 := decorator.GetUserByCodeUser(context.Background(), "user-123", "token", "app", "corr-123")
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result2)

	// Verifica que o mock foi chamado apenas uma vez
	mockClient.AssertExpectations(t)
}

func TestUserCacheDecorator_GetUserByCodeUser_CacheExpired(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             100 * time.Millisecond, // TTL muito curto
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	expectedResponse := userDto.MSUserResponseDTO{
		ID:       "123",
		CodeUser: "user-123",
		Email:    "test@example.com",
	}

	// Deve chamar o mock duas vezes (cache miss + cache expired)
	mockClient.On("GetUserByCodeUser", mock.Anything, "user-123", "token", "app", "corr-123").Return(expectedResponse, nil).Twice()

	// Act & Assert
	// Primeira chamada - cache miss
	result1, err1 := decorator.GetUserByCodeUser(context.Background(), "user-123", "token", "app", "corr-123")
	assert.NoError(t, err1)
	assert.Equal(t, expectedResponse, result1)

	// Aguarda o cache expirar
	time.Sleep(150 * time.Millisecond)

	// Segunda chamada - cache expired, deve chamar o mock novamente
	result2, err2 := decorator.GetUserByCodeUser(context.Background(), "user-123", "token", "app", "corr-123")
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result2)

	// Verifica que o mock foi chamado duas vezes
	mockClient.AssertExpectations(t)
}

func TestUserCacheDecorator_GetUserByCodeUser_Error(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	expectedError := assert.AnError

	// Mock retorna erro
	mockClient.On("GetUserByCodeUser", mock.Anything, "user-123", "token", "app", "corr-123").Return(userDto.MSUserResponseDTO{}, expectedError).Once()

	// Act
	result, err := decorator.GetUserByCodeUser(context.Background(), "user-123", "token", "app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, userDto.MSUserResponseDTO{}, result)

	mockClient.AssertExpectations(t)
}

func TestUserCacheDecorator_CreateUser_NoCache(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

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

	// CreateUser não usa cache, deve chamar o mock
	mockClient.On("CreateUser", mock.Anything, req, "app", "corr-123").Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.CreateUser(context.Background(), req, "app", "corr-123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestUserCacheDecorator_CreateUserNotify_NoCache(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	req := userDto.MSUserNotifyCreateRequestDTO{
		UserID: "123",
	}

	expectedResponse := userDto.MSUserNotifyResponseDTO{
		ID:     "notify-123",
		UserID: "123",
	}

	// CreateUserNotify não usa cache, deve chamar o mock
	mockClient.On("CreateUserNotify", mock.Anything, req, "app", "corr-123").Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.CreateUserNotify(context.Background(), req, "app", "corr-123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestUserCacheDecorator_InitRegister_NoCache(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	req := userDto.MSUserRegisterInitRequestDTO{
		Email: "test@example.com",
	}

	expectedResponse := userDto.MSUserRegisterInitResponseDTO{
		RegistrationSessionID: "session-123",
		Email:                 "test@example.com",
		ExpiresIn:             300,
	}

	// InitRegister não usa cache, deve chamar o mock
	mockClient.On("InitRegister", mock.Anything, req, "app", "corr-123").Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.InitRegister(context.Background(), req, "app", "corr-123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockClient.AssertExpectations(t)
}

func TestUserCacheDecorator_ConfirmRegister_NoCache(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	req := userDto.MSUserRegisterConfirmRequestDTO{
		Email:                 "test@example.com",
		RegistrationSessionID: "session-123",
		Token:                 "123456",
	}

	// ConfirmRegister não usa cache, deve chamar o mock
	expectedResp := userDto.MSUserRegisterConfirmResponseDTO{Message: "Success"}
	mockClient.On("ConfirmRegister", mock.Anything, req, "app", "corr-123").Return(expectedResp, nil).Once()

	// Act
	resp, err := decorator.ConfirmRegister(context.Background(), req, "app", "corr-123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestUserCacheDecorator_Stop(t *testing.T) {
	// Arrange
	mockClient := &MockUserClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 100 * time.Millisecond,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	// Act
	decorator.Stop()

	// Aguarda um pouco para garantir que o cleanup foi parado
	time.Sleep(200 * time.Millisecond)

	// Assert - não deve haver panics ou erros
	assert.NotNil(t, decorator)
}
