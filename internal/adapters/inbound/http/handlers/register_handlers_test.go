package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRegisterInitUseCase para teste
type MockRegisterInitUseCase struct {
	mock.Mock
}

func (m *MockRegisterInitUseCase) Execute(command appdto.RegisterInitCommand) (dto.RegisterInitResponseDTO, error) {
	args := m.Called(command)
	return args.Get(0).(dto.RegisterInitResponseDTO), args.Error(1)
}

// MockRegisterConfirmUseCase para teste
type MockRegisterConfirmUseCase struct {
	mock.Mock
}

func (m *MockRegisterConfirmUseCase) Execute(command appdto.RegisterConfirmCommand) (dto.RegisterConfirmResponseDTO, error) {
	args := m.Called(command)
	return args.Get(0).(dto.RegisterConfirmResponseDTO), args.Error(1)
}

// MockRegisterResendUseCase para teste
type MockRegisterResendUseCase struct {
	mock.Mock
}

func (m *MockRegisterResendUseCase) Execute(command appdto.RegisterResendCommand) (dto.RegisterResendResponseDTO, error) {
	args := m.Called(command)
	return args.Get(0).(dto.RegisterResendResponseDTO), args.Error(1)
}

func TestNewRegisterHandlers(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger := zap.NewNop()

	// Act
	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Assert
	assert.NotNil(t, handlers)
	assert.IsType(t, &RegisterHandlers{}, handlers)
}

func TestRegisterHandlers_InitRegisterHandler_Success(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Mock response
	expectedResponse := dto.RegisterInitResponseDTO{
		RegistrationSessionID: "session-123",
		Email:                 "test@example.com",
		ExpiresIn:             1800,
	}

	// Setup mock
	mockInitUseCase.On("Execute", mock.Anything).Return(expectedResponse, nil)

	// Create request
	requestBody := dto.RegisterInitRequestDTO{
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          &[]bool{false}[0],
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
	}

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "test-app")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.InitRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response dto.RegisterInitResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "session-123", response.RegistrationSessionID)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, 1800, response.ExpiresIn)

	mockInitUseCase.AssertExpectations(t)
}

func TestRegisterHandlers_InitRegisterHandler_MissingCorrelationID(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Create request without X-Correlation-ID
	requestBody := dto.RegisterInitRequestDTO{
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          &[]bool{false}[0],
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
	}

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "test-app")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.InitRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MISSING_HEADER", response.Error)
	assert.Equal(t, "Header X-Correlation-ID é obrigatório", response.Message)

	mockInitUseCase.AssertNotCalled(t, "Execute")
}

func TestRegisterHandlers_InitRegisterHandler_MissingTenantId(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Create request without X-Tenant-Id
	requestBody := dto.RegisterInitRequestDTO{
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          &[]bool{false}[0],
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
	}

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-123")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.InitRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MISSING_HEADER", response.Error)
	assert.Equal(t, "Header X-Tenant-Id é obrigatório", response.Message)

	mockInitUseCase.AssertNotCalled(t, "Execute")
}

func TestRegisterHandlers_InitRegisterHandler_InvalidJSON(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "test-app")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.InitRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", response.Error)
	assert.Equal(t, "Requisição inválida", response.Message)

	mockInitUseCase.AssertNotCalled(t, "Execute")
}

func TestRegisterHandlers_InitRegisterHandler_UseCaseError(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Setup mock to return error
	mockInitUseCase.On("Execute", mock.Anything).Return(dto.RegisterInitResponseDTO{}, errors.New("email already exists"))

	// Create request
	requestBody := dto.RegisterInitRequestDTO{
		NameFull:                   "Test User",
		Email:                      "test@example.com",
		Password:                   "password123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		AcceptedMarketing:          &[]bool{false}[0],
		IPAddress:                  "192.168.1.1",
		UserAgent:                  "Mozilla/5.0",
		Geolocation:                "São Paulo, SP",
		Type:                       "PERSON",
	}

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "test-app")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.InitRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", response.Error)
	assert.Equal(t, "Erro interno do servidor", response.Message)

	mockInitUseCase.AssertExpectations(t)
}

func TestRegisterHandlers_ConfirmRegisterHandler_Success(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Mock response
	expectedResponse := dto.RegisterConfirmResponseDTO{
		Token:          "jwt-token-123",
		TokenExpiresIn: 3600,
	}

	// Setup mock
	mockConfirmUseCase.On("Execute", mock.Anything).Return(expectedResponse, nil)

	// Create request
	requestBody := dto.RegisterConfirmRequestDTO{
		Email:                 "test@example.com",
		RegistrationSessionID: "session-123",
		Token:                 "123456",
	}

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/confirm", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "test-app")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.ConfirmRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.RegisterConfirmResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "jwt-token-123", response.Token)
	assert.Equal(t, int64(3600), response.TokenExpiresIn)

	mockConfirmUseCase.AssertExpectations(t)
}

func TestRegisterHandlers_ConfirmRegisterHandler_MissingCorrelationID(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Create request without X-Correlation-ID
	requestBody := dto.RegisterConfirmRequestDTO{
		Email:                 "test@example.com",
		RegistrationSessionID: "session-123",
		Token:                 "123456",
	}

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/confirm", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "test-app")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.ConfirmRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MISSING_HEADER", response.Error)
	assert.Equal(t, "Header X-Correlation-ID é obrigatório", response.Message)

	mockConfirmUseCase.AssertNotCalled(t, "Execute")
}

func TestRegisterHandlers_ConfirmRegisterHandler_UseCaseError(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Setup mock to return error
	mockConfirmUseCase.On("Execute", mock.Anything).Return(dto.RegisterConfirmResponseDTO{}, errors.New("invalid token"))

	// Create request
	requestBody := dto.RegisterConfirmRequestDTO{
		Email:                 "test@example.com",
		RegistrationSessionID: "session-123",
		Token:                 "123456",
	}

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/confirm", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "test-app")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.ConfirmRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", response.Error)
	assert.Equal(t, "Erro interno do servidor", response.Message)

	mockConfirmUseCase.AssertExpectations(t)
}

func TestRegisterHandlers_VerifyCommandValidation(t *testing.T) {
	// Arrange
	mockInitUseCase := &MockRegisterInitUseCase{}
	mockConfirmUseCase := &MockRegisterConfirmUseCase{}
	mockResendUseCase := &MockRegisterResendUseCase{}
	logger, _ := zap.NewDevelopment()

	handlers := NewRegisterHandlers(mockInitUseCase, mockConfirmUseCase, mockResendUseCase, logger)

	// Test with missing required fields
	requestBody := dto.RegisterInitRequestDTO{
		NameFull: "", // Missing required field
		Email:    "test@example.com",
		Password: "password123",
		Phone:    "+5511999999999",
	}

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "test-app")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	err := handlers.InitRegisterHandler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response pkg.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response.Error)

	mockInitUseCase.AssertNotCalled(t, "Execute")
}
