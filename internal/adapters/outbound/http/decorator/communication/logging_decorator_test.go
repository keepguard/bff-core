package communication

import (
	"context"
	"errors"
	"testing"

	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewCommunicationLoggingDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	logger, _ := zap.NewDevelopment()
	serviceName := "test-service"

	// Act
	decorator := NewCommunicationLoggingDecorator(mockInner, logger, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &communicationLoggingDecorator{}, decorator)
}

func TestCommunicationLoggingDecorator_SendNotification_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-communication"

	decorator := NewCommunicationLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	req := portsclient.SendNotificationRequestDTO{
		UserID:       "user-123",
		TemplateType: "WELCOME_EMAIL",
		Channel:      "email",
		Recipient:    "user@example.com",
		Data:         map[string]string{"name": "João"},
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("SendNotification", ctx, req, xApplication, correlationID).Return(nil)

	// Act
	err := decorator.SendNotification(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)

	// Verifica logs
	logEntries := logs.All()
	assert.Len(t, logEntries, 2) // Request + Response
	assert.Equal(t, "Iniciando requisição", logEntries[0].Message)
	assert.Equal(t, "Notificação enviada com sucesso", logEntries[1].Message)
}

func TestCommunicationLoggingDecorator_SendNotification_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-communication"

	decorator := NewCommunicationLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	req := portsclient.SendNotificationRequestDTO{
		UserID:       "user-123",
		TemplateType: "INVALID_TEMPLATE",
		Channel:      "email",
		Recipient:    "invalid@example.com",
		Data:         map[string]string{},
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("template not found")

	mockInner.On("SendNotification", ctx, req, xApplication, correlationID).Return(expectedError)

	// Act
	err := decorator.SendNotification(ctx, req, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)

	// Verifica logs
	logEntries := logs.All()
	assert.Len(t, logEntries, 2) // Request + Response
	assert.Equal(t, "Iniciando requisição", logEntries[0].Message)
	assert.Equal(t, "Erro na requisição", logEntries[1].Message)
}
