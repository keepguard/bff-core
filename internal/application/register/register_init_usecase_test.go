package register

import (
	"context"
	"errors"
	"testing"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestNewRegisterInitUseCase(t *testing.T) {
	// Arrange
	mockAuthClient := &MockAuthClient{}
	mockUserClient := &MockUserClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}
	mockMessagePublisher := &MockMessagePublisher{}
	logger := zap.NewNop()

	// Act
	useCase := NewRegisterInitUseCase(mockAuthClient, mockUserClient, mockCompanyClient, mockCommunicationClient, mockMessagePublisher, logger)

	// Assert
	assert.NotNil(t, useCase)
	assert.IsType(t, &registerInitUseCaseImpl{}, useCase)
}

func TestRegisterInitUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockAuthClient := &MockAuthClient{}
	mockUserClient := &MockUserClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}
	mockMessagePublisher := &MockMessagePublisher{}
	logger, _ := zap.NewDevelopment()

	useCase := &registerInitUseCaseImpl{
		authClient:          mockAuthClient,
		userClient:          mockUserClient,
		companyClient:       mockCompanyClient,
		communicationClient: mockCommunicationClient,
		messagePublisher:    mockMessagePublisher,
		logger:              logger,
	}

	// Mock data
	company := companyDto.MSCompanyResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	registerResponse := userDto.MSUserRegisterInitResponseDTO{
		RegistrationSessionID: "session-123",
		Email:                 "test@example.com",
		Token:                 "token-123",
		ExpiresIn:             1800, // 30 minutos
	}

	// Setup mocks
	mockCompanyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(company, nil)
	mockUserClient.On("InitRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(registerResponse, nil)
	mockMessagePublisher.On("PublishMessage", mock.Anything, mock.Anything).Return(nil)

	command := appdto.RegisterInitCommand{
		Context:                    context.Background(),
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          false,
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
		TenantId:               "test-app",
		CorrelationID:              "corr-123",
	}

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "session-123", result.RegistrationSessionID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, 1800, result.ExpiresIn)

	// Verify all mocks were called
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockMessagePublisher.AssertExpectations(t)
}

func TestRegisterInitUseCase_Execute_CompanyNotFound(t *testing.T) {
	// Arrange
	mockAuthClient := &MockAuthClient{}
	mockUserClient := &MockUserClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}
	mockMessagePublisher := &MockMessagePublisher{}
	logger, _ := zap.NewDevelopment()

	useCase := &registerInitUseCaseImpl{
		authClient:          mockAuthClient,
		userClient:          mockUserClient,
		companyClient:       mockCompanyClient,
		communicationClient: mockCommunicationClient,
		messagePublisher:    mockMessagePublisher,
		logger:              logger,
	}

	// Setup mocks - company not found
	mockCompanyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, errors.New("company not found"))

	command := appdto.RegisterInitCommand{
		Context:                    context.Background(),
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          false,
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
		TenantId:               "test-app",
		CorrelationID:              "corr-123",
	}

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result.RegistrationSessionID)
	assert.Equal(t, "", result.Email)
	assert.Equal(t, 0, result.ExpiresIn)

	// Verify only company client was called
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertNotCalled(t, "InitRegister")
	mockCommunicationClient.AssertNotCalled(t, "SendMessage")
}

func TestRegisterInitUseCase_Execute_InitRegisterFailed(t *testing.T) {
	// Arrange
	mockAuthClient := &MockAuthClient{}
	mockUserClient := &MockUserClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}
	mockMessagePublisher := &MockMessagePublisher{}
	logger, _ := zap.NewDevelopment()

	useCase := &registerInitUseCaseImpl{
		authClient:          mockAuthClient,
		userClient:          mockUserClient,
		companyClient:       mockCompanyClient,
		communicationClient: mockCommunicationClient,
		messagePublisher:    mockMessagePublisher,
		logger:              logger,
	}

	// Mock data
	company := companyDto.MSCompanyResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	// Setup mocks - init register fails
	mockCompanyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(company, nil)
	mockUserClient.On("InitRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(userDto.MSUserRegisterInitResponseDTO{}, errors.New("email already exists"))

	command := appdto.RegisterInitCommand{
		Context:                    context.Background(),
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          false,
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
		TenantId:               "test-app",
		CorrelationID:              "corr-123",
	}

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result.RegistrationSessionID)
	assert.Equal(t, "", result.Email)
	assert.Equal(t, 0, result.ExpiresIn)

	// Verify company and user clients were called
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockCommunicationClient.AssertNotCalled(t, "SendMessage")
}

func TestRegisterInitUseCase_Execute_CommunicationFailed_StillSuccess(t *testing.T) {
	// Arrange
	mockAuthClient := &MockAuthClient{}
	mockUserClient := &MockUserClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}
	mockMessagePublisher := &MockMessagePublisher{}
	logger, _ := zap.NewDevelopment()

	useCase := &registerInitUseCaseImpl{
		authClient:          mockAuthClient,
		userClient:          mockUserClient,
		companyClient:       mockCompanyClient,
		communicationClient: mockCommunicationClient,
		messagePublisher:    mockMessagePublisher,
		logger:              logger,
	}

	// Mock data
	company := companyDto.MSCompanyResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	registerResponse := userDto.MSUserRegisterInitResponseDTO{
		RegistrationSessionID: "session-123",
		Email:                 "test@example.com",
		Token:                 "token-123",
		ExpiresIn:             1800,
	}

	// Setup mocks - communication fails but registration should still succeed
	mockCompanyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(company, nil)
	mockUserClient.On("InitRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(registerResponse, nil)
	mockMessagePublisher.On("PublishMessage", mock.Anything, mock.Anything).Return(errors.New("email service unavailable"))

	command := appdto.RegisterInitCommand{
		Context:                    context.Background(),
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          false,
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
		TenantId:               "test-app",
		CorrelationID:              "corr-123",
	}

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err) // Should not fail even if email fails
	assert.Equal(t, "session-123", result.RegistrationSessionID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, 1800, result.ExpiresIn)

	// Verify all mocks were called
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockMessagePublisher.AssertExpectations(t)
}

func TestRegisterInitUseCase_Execute_VerifyRequestData(t *testing.T) {
	// Arrange
	mockAuthClient := &MockAuthClient{}
	mockUserClient := &MockUserClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}
	mockMessagePublisher := &MockMessagePublisher{}
	logger, _ := zap.NewDevelopment()

	useCase := &registerInitUseCaseImpl{
		authClient:          mockAuthClient,
		userClient:          mockUserClient,
		companyClient:       mockCompanyClient,
		communicationClient: mockCommunicationClient,
		messagePublisher:    mockMessagePublisher,
		logger:              logger,
	}

	// Mock data
	company := companyDto.MSCompanyResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	registerResponse := userDto.MSUserRegisterInitResponseDTO{
		RegistrationSessionID: "session-123",
		Email:                 "test@example.com",
		Token:                 "token-123",
		ExpiresIn:             1800,
	}

	// Setup mocks with specific assertions
	mockCompanyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(company, nil)

	// Capture the request to verify its content
	var capturedRequest userDto.MSUserRegisterInitRequestDTO
	mockUserClient.On("InitRegister", mock.Anything, mock.MatchedBy(func(req userDto.MSUserRegisterInitRequestDTO) bool {
		capturedRequest = req
		return req.CompanyID == company.ID &&
			req.Email == "test@example.com" &&
			req.NameFull == "Test User" &&
			req.Password == "password123" &&
			req.Phone == "+5511999999999" &&
			req.HasAcceptedTermsAndPrivacy == true &&
			req.AcceptedMarketing == false &&
			req.IPAddress == "192.168.1.1" &&
			req.UserAgent == "Mozilla/5.0" &&
			req.Geolocation == "São Paulo, SP" &&
			req.Type == "PERSON"
	}), "test-app", "corr-123").Return(registerResponse, nil)

	// Capture the message request to verify template variables
	var capturedMessage messaging.MessageDTO
	mockMessagePublisher.On("PublishMessage", mock.Anything, mock.MatchedBy(func(req messaging.MessageDTO) bool {
		capturedMessage = req
		return req.MessageType == "EMAIL" &&
			req.CommunicationType == "EMAIL" &&
			req.TemplateType == "AUTENTICACAO_EMAIL_TOKEN" &&
			req.Recipient == "test@example.com" &&
			req.CodeUser == "session-123" &&
			req.Variables["userName"] == "Test User" &&
			req.Variables["token"] == "token-123" &&
			req.Variables["expiresIn"] == "30" && // 1800/60 = 30 minutes
			req.Variables["appName"] == "Test Company"
	})).Return(nil)

	command := appdto.RegisterInitCommand{
		Context:                    context.Background(),
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          false,
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
		TenantId:               "test-app",
		CorrelationID:              "corr-123",
	}

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "session-123", result.RegistrationSessionID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, 1800, result.ExpiresIn)

	// Verify captured data
	assert.Equal(t, "company-123", capturedRequest.CompanyID)
	assert.Equal(t, "test@example.com", capturedRequest.Email)
	assert.Equal(t, "Test User", capturedRequest.NameFull)
	assert.Equal(t, "password123", capturedRequest.Password)
	assert.Equal(t, "+5511999999999", capturedRequest.Phone)
	assert.Equal(t, true, capturedRequest.HasAcceptedTermsAndPrivacy)
	assert.Equal(t, false, capturedRequest.AcceptedMarketing)
	assert.Equal(t, "192.168.1.1", capturedRequest.IPAddress)
	assert.Equal(t, "Mozilla/5.0", capturedRequest.UserAgent)
	assert.Equal(t, "São Paulo, SP", capturedRequest.Geolocation)
	assert.Equal(t, "PERSON", capturedRequest.Type)

	// Verify message template variables
	assert.Equal(t, "Test User", capturedMessage.Variables["userName"])
	assert.Equal(t, "token-123", capturedMessage.Variables["token"])
	assert.Equal(t, "30", capturedMessage.Variables["expiresIn"])
	assert.Equal(t, "Test Company", capturedMessage.Variables["appName"])

	// Verify all mocks were called
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockMessagePublisher.AssertExpectations(t)
}
