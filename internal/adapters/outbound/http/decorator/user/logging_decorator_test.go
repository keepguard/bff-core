package user

import (
	"context"
	"errors"
	"testing"

	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewUserLoggingDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	logger, _ := zap.NewDevelopment()
	serviceName := "test-service"

	// Act
	decorator := NewUserLoggingDecorator(mockInner, logger, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &userLoggingDecorator{}, decorator)
}

func TestUserLoggingDecorator_CreateUser_Success(t *testing.T) {
	// Arrange
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "test-service"

	mockInner := new(MockUserClient)
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

	mockInner.On("CreateUser", mock.Anything, req, "app", "corr-123").Return(expectedResponse, nil).Once()

	decorator := NewUserLoggingDecorator(mockInner, logger, serviceName)

	// Act
	result, err := decorator.CreateUser(context.Background(), req, "app", "corr-123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)

	// Verifica logs
	logEntries := logs.All()
	assert.Len(t, logEntries, 2) // Request + Response

	// Verifica log de request
	requestLog := logEntries[0]
	assert.Equal(t, "Iniciando requisição", requestLog.Message)

	// Verifica log de response
	responseLog := logEntries[1]
	assert.Equal(t, "Requisição concluída com sucesso", responseLog.Message)
}

func TestUserLoggingDecorator_CreateUser_Error(t *testing.T) {
	// Arrange
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "test-service"

	mockInner := new(MockUserClient)
	req := userDto.MSUserCreateRequestDTO{
		Email:     "test@example.com",
		CompanyID: "company123",
		Type:      "PERSON",
	}
	expectedError := errors.New("user already exists")

	mockInner.On("CreateUser", mock.Anything, req, "app", "corr-123").Return(userDto.MSUserResponseDTO{}, expectedError).Once()

	decorator := NewUserLoggingDecorator(mockInner, logger, serviceName)

	// Act
	result, err := decorator.CreateUser(context.Background(), req, "app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, userDto.MSUserResponseDTO{}, result)
	mockInner.AssertExpectations(t)

	// Verifica logs
	logEntries := logs.All()
	assert.Len(t, logEntries, 2) // Request + Response

	// Verifica log de response com erro
	responseLog := logEntries[1]
	assert.Equal(t, "Erro na requisição", responseLog.Message)
}
