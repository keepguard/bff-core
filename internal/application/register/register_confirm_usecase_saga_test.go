package register

import (
	"context"
	"testing"
	"time"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	userConsentDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user_consent"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/domain/saga"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock clients para teste
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	args := m.Called(ctx, codeUser, token, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserResponseDTO), args.Error(1)
}

func (m *MockUserClient) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserResponseDTO), args.Error(1)
}

func (m *MockUserClient) GetByEmail(ctx context.Context, email, tenantId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	args := m.Called(ctx, email, tenantId, correlationID)
	return args.Get(0).(authDto.UserByEmailResponseDTO), args.Error(1)
}

func (m *MockUserClient) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserNotifyResponseDTO), args.Error(1)
}

func (m *MockUserClient) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserRegisterInitResponseDTO), args.Error(1)
}

func (m *MockUserClient) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserRegisterConfirmResponseDTO), args.Error(1)
}

func (m *MockUserClient) DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error {
	args := m.Called(ctx, userID, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockUserClient) ResendRegisterToken(ctx context.Context, req userDto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserRegisterResendResponseDTO), args.Error(1)
}

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) CreateUser(ctx context.Context, req authDto.AuthUserCreateRequestDTO, tenantId, correlationID string) (authDto.AuthUserCreateResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(authDto.AuthUserCreateResponseDTO), args.Error(1)
}

func (m *MockAuthClient) RegisterLogin(ctx context.Context, req authDto.AuthRegisterLoginRequestDTO, tenantId, correlationID, clientId string) (authDto.AuthLoginResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(authDto.AuthLoginResponseDTO), args.Error(1)
}

func (m *MockAuthClient) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	args := m.Called(ctx, token, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockAuthClient) GenerateResetToken(ctx context.Context, req authDto.GenerateResetTokenMSRequestDTO, tenantId, correlationID string) (authDto.GenerateResetTokenMSResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(authDto.GenerateResetTokenMSResponseDTO), args.Error(1)
}

func (m *MockAuthClient) HardDeleteUser(ctx context.Context, idUserExternal, tenantId, correlationID string) error {
	args := m.Called(ctx, idUserExternal, tenantId, correlationID)
	return args.Error(0)
}

type MockUserConsentClient struct {
	mock.Mock
}

func (m *MockUserConsentClient) Accept(ctx context.Context, req userConsentDto.UserConsentAcceptRequestDTO, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	args := m.Called(ctx, req, token, tenantId, correlationID)
	return args.Get(0).(userConsentDto.UserConsentResponseDTO), args.Error(1)
}

func (m *MockUserConsentClient) FindByID(ctx context.Context, id, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	args := m.Called(ctx, id, token, tenantId, correlationID)
	return args.Get(0).(userConsentDto.UserConsentResponseDTO), args.Error(1)
}

func (m *MockUserConsentClient) FindByUserID(ctx context.Context, userID, token, tenantId, correlationID string) ([]userConsentDto.UserConsentResponseDTO, error) {
	args := m.Called(ctx, userID, token, tenantId, correlationID)
	return args.Get(0).([]userConsentDto.UserConsentResponseDTO), args.Error(1)
}

func (m *MockUserConsentClient) FindByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) ([]userConsentDto.UserConsentResponseDTO, error) {
	args := m.Called(ctx, userID, consentDocumentID, token, tenantId, correlationID)
	return args.Get(0).([]userConsentDto.UserConsentResponseDTO), args.Error(1)
}

func (m *MockUserConsentClient) FindLatestByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	args := m.Called(ctx, userID, consentDocumentID, token, tenantId, correlationID)
	return args.Get(0).(userConsentDto.UserConsentResponseDTO), args.Error(1)
}

func (m *MockUserConsentClient) HasAccepted(ctx context.Context, userID, consentDocumentID string, version int, token, tenantId, correlationID string) (bool, error) {
	args := m.Called(ctx, userID, consentDocumentID, version, token, tenantId, correlationID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserConsentClient) AcceptAll(ctx context.Context, req userConsentDto.UserConsentAcceptAllRequestDTO, tenantId, correlationID string) (userConsentDto.UserConsentAcceptAllResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userConsentDto.UserConsentAcceptAllResponseDTO), args.Error(1)
}

func (m *MockUserConsentClient) DeleteAllByUserId(ctx context.Context, userID, tenantId, correlationID string) error {
	args := m.Called(ctx, userID, tenantId, correlationID)
	return args.Error(0)
}

type MockCompanyClient struct {
	mock.Mock
}

func (m *MockCompanyClient) GetByTenantId(ctx context.Context, tenantId, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	args := m.Called(ctx, tenantId, correlationID)
	return args.Get(0).(companyDto.MSCompanyResponseDTO), args.Error(1)
}

type MockCommunicationClient struct {
	mock.Mock
}

func (m *MockCommunicationClient) SendMessage(ctx context.Context, req communicationDto.SendMessageRequestDTO, tenantId, correlationID string) (communicationDto.SendMessageResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(communicationDto.SendMessageResponseDTO), args.Error(1)
}

func (m *MockCommunicationClient) SendNotification(ctx context.Context, req client.SendNotificationRequestDTO, tenantId, correlationID string) error {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Error(0)
}

func TestRegisterConfirmUseCase_SAGASuccess(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	sagaExecutor := saga.NewInMemorySagaExecutor(logger)

	mockUserClient := &MockUserClient{}
	mockAuthClient := &MockAuthClient{}
	mockUserConsentClient := &MockUserConsentClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}
	mockMessagePublisher := &MockMessagePublisher{}

	useCase := &registerConfirmUseCaseImpl{
		logger:              logger,
		userClient:          mockUserClient,
		authClient:          mockAuthClient,
		userConsentClient:   mockUserConsentClient,
		companyClient:       mockCompanyClient,
		communicationClient: mockCommunicationClient,
		messagePublisher:    mockMessagePublisher,
		sagaExecutor:        sagaExecutor,
	}

	// Mock data
	company := companyDto.MSCompanyResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	confirmResponse := userDto.MSUserRegisterConfirmResponseDTO{
		Email:    "test@example.com",
		NameFull: "Test User",
		Phone:    "+5511999999999",
		Type:     "PERSON",
		Message:  "Registration confirmed",
	}

	userResponse := userDto.MSUserResponseDTO{
		ID:        "user-123",
		Email:     "test@example.com",
		CompanyID: "company-123",
		CodeUser:  "code-123",
	}

	notifyResponse := userDto.MSUserNotifyResponseDTO{
		ID:     "notify-123",
		UserID: "user-123",
	}

	authUserResponse := authDto.AuthUserCreateResponseDTO{
		ID:            "auth-123",
		CodeUser:      "code-123",
		Username:      "test@example.com",
		Email:         "test@example.com",
		CompanyID:     "company-123",
		Type:          "PERSON",
		Status:        "ACTIVE",
		EmailVerified: true,
		Roles:         []string{"USER"},
	}

	consentResponse := userConsentDto.UserConsentAcceptAllResponseDTO{
		TotalAccepted: 3,
		AcceptedConsents: []userConsentDto.AcceptedConsentItemDTO{
			{
				ID:                "consent-1",
				UserID:            "user-123",
				Email:             "test@example.com",
				ConsentDocumentID: "doc-1",
				Version:           1,
			},
		},
	}

	loginResponse := authDto.AuthLoginResponseDTO{
		Token:     "jwt-token-123",
		ExpiresIn: 3600,
	}

	// Setup mocks
	mockCompanyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(company, nil)
	mockUserClient.On("ConfirmRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(confirmResponse, nil)
	mockUserClient.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").Return(userResponse, nil)
	mockUserClient.On("CreateUserNotify", mock.Anything, mock.Anything, "test-app", "corr-123").Return(notifyResponse, nil)
	mockAuthClient.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").Return(authUserResponse, nil)
	mockUserConsentClient.On("AcceptAll", mock.Anything, mock.Anything, "test-app", "corr-123").Return(consentResponse, nil)
	mockAuthClient.On("RegisterLogin", mock.Anything, mock.Anything, "test-app", "corr-123", mock.Anything).Return(loginResponse, nil)
	mockMessagePublisher.On("PublishMessage", mock.Anything, mock.Anything).Return(nil)

	command := appdto.RegisterConfirmCommand{
		Context:               context.Background(),
		Email:                 "test@example.com",
		RegistrationSessionID: "session-123",
		Token:                 "token-123",
		TenantId:          "test-app",
		CorrelationID:         "corr-123",
	}

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "jwt-token-123", result.Token)
	assert.Equal(t, int64(3600), result.TokenExpiresIn)

	// Verify all mocks were called
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
	mockUserConsentClient.AssertExpectations(t)
	mockMessagePublisher.AssertExpectations(t)
}

func TestRegisterConfirmUseCase_SAGAFailureWithCompensation(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	sagaExecutor := saga.NewInMemorySagaExecutor(logger)

	mockUserClient := &MockUserClient{}
	mockAuthClient := &MockAuthClient{}
	mockUserConsentClient := &MockUserConsentClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}

	useCase := &registerConfirmUseCaseImpl{
		logger:              logger,
		userClient:          mockUserClient,
		authClient:          mockAuthClient,
		userConsentClient:   mockUserConsentClient,
		companyClient:       mockCompanyClient,
		communicationClient: mockCommunicationClient,
		sagaExecutor:        sagaExecutor,
	}

	// Mock data
	company := companyDto.MSCompanyResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	confirmResponse := userDto.MSUserRegisterConfirmResponseDTO{
		Email:    "test@example.com",
		NameFull: "Test User",
		Phone:    "+5511999999999",
		Type:     "PERSON",
		Message:  "Registration confirmed",
	}

	userResponse := userDto.MSUserResponseDTO{
		ID:        "user-123",
		Email:     "test@example.com",
		CompanyID: "company-123",
		CodeUser:  "code-123",
	}

	notifyResponse := userDto.MSUserNotifyResponseDTO{
		ID:     "notify-123",
		UserID: "user-123",
	}

	authUserResponse := authDto.AuthUserCreateResponseDTO{
		ID:            "auth-123",
		CodeUser:      "code-123",
		Username:      "test@example.com",
		Email:         "test@example.com",
		CompanyID:     "company-123",
		Type:          "PERSON",
		Status:        "ACTIVE",
		EmailVerified: true,
		Roles:         []string{"USER"},
	}

	// Setup mocks - sucesso até CreateAuthUser, depois falha no AcceptAllConsents
	mockCompanyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(company, nil)
	mockUserClient.On("ConfirmRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(confirmResponse, nil)
	mockUserClient.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").Return(userResponse, nil)
	mockUserClient.On("CreateUserNotify", mock.Anything, mock.Anything, "test-app", "corr-123").Return(notifyResponse, nil)
	mockAuthClient.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").Return(authUserResponse, nil)

	// Falha no AcceptAll
	mockUserConsentClient.On("AcceptAll", mock.Anything, mock.Anything, "test-app", "corr-123").Return(userConsentDto.UserConsentAcceptAllResponseDTO{}, assert.AnError)

	// Compensação - DeleteAllByUserId não deve ser chamado pois AcceptAll falhou
	// Compensação - HardDeleteUser deve ser chamado
	mockAuthClient.On("HardDeleteUser", mock.Anything, "user-123", "test-app", "corr-123").Return(nil)
	// Compensação - DeleteUser deve ser chamado
	mockUserClient.On("DeleteUser", mock.Anything, "user-123", "test-app", "corr-123").Return(nil)

	command := appdto.RegisterConfirmCommand{
		Context:               context.Background(),
		Email:                 "test@example.com",
		RegistrationSessionID: "session-123",
		Token:                 "token-123",
		TenantId:          "test-app",
		CorrelationID:         "corr-123",
	}

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result.Token)
	assert.Equal(t, int64(0), result.TokenExpiresIn)

	// Verify all mocks were called including compensations
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
	mockUserConsentClient.AssertExpectations(t)

	// CommunicationClient should not be called since SAGA failed
	mockCommunicationClient.AssertNotCalled(t, "SendMessage")
}

func TestRegisterConfirmUseCase_SAGATimeout(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	sagaExecutor := saga.NewInMemorySagaExecutor(logger)

	mockUserClient := &MockUserClient{}
	mockAuthClient := &MockAuthClient{}
	mockUserConsentClient := &MockUserConsentClient{}
	mockCompanyClient := &MockCompanyClient{}
	mockCommunicationClient := &MockCommunicationClient{}

	useCase := &registerConfirmUseCaseImpl{
		logger:              logger,
		userClient:          mockUserClient,
		authClient:          mockAuthClient,
		userConsentClient:   mockUserConsentClient,
		companyClient:       mockCompanyClient,
		communicationClient: mockCommunicationClient,
		sagaExecutor:        sagaExecutor,
	}

	// Mock data
	company := companyDto.MSCompanyResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	confirmResponse := userDto.MSUserRegisterConfirmResponseDTO{
		Email:    "test@example.com",
		NameFull: "Test User",
		Phone:    "+5511999999999",
		Type:     "PERSON",
		Message:  "Registration confirmed",
	}

	userResponse := userDto.MSUserResponseDTO{
		ID:        "user-123",
		Email:     "test@example.com",
		CompanyID: "company-123",
		CodeUser:  "code-123",
	}

	// Setup mocks - sucesso até CreateUser, depois timeout no CreateUserNotify
	mockCompanyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(company, nil)
	mockUserClient.On("ConfirmRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(confirmResponse, nil)
	mockUserClient.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").Return(userResponse, nil)

	// Timeout no CreateUserNotify (simula timeout com delay)
	mockUserClient.On("CreateUserNotify", mock.Anything, mock.Anything, "test-app", "corr-123").Run(func(args mock.Arguments) {
		time.Sleep(6 * time.Second) // Timeout maior que o configurado (5s)
	}).Return(userDto.MSUserNotifyResponseDTO{}, nil)

	// Não configurar CreateAuthUser pois o teste deve falhar antes

	// Compensação - DeleteUser deve ser chamado
	mockUserClient.On("DeleteUser", mock.Anything, "user-123", "test-app", "corr-123").Return(nil)

	command := appdto.RegisterConfirmCommand{
		Context:               context.Background(),
		Email:                 "test@example.com",
		RegistrationSessionID: "session-123",
		Token:                 "token-123",
		TenantId:          "test-app",
		CorrelationID:         "corr-123",
	}

	// Act
	result, err := useCase.Execute(command)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", result.Token)
	assert.Equal(t, int64(0), result.TokenExpiresIn)

	// Verify compensation was called
	mockCompanyClient.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)

	// Other clients should not be called since SAGA failed early
	mockAuthClient.AssertNotCalled(t, "CreateUser")
	mockUserConsentClient.AssertNotCalled(t, "AcceptAll")
	mockCommunicationClient.AssertNotCalled(t, "SendMessage")
}
