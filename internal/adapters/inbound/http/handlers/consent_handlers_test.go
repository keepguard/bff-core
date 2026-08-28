package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	userConsentDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user_consent"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockConsentClient struct {
	mock.Mock
}

func (m *mockConsentClient) Accept(ctx context.Context, req userConsentDto.UserConsentAcceptRequestDTO, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	panic("not used")
}

func (m *mockConsentClient) FindByID(ctx context.Context, id, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	panic("not used")
}

func (m *mockConsentClient) FindByUserID(ctx context.Context, userID, token, tenantId, correlationID string) ([]userConsentDto.UserConsentResponseDTO, error) {
	panic("not used")
}

func (m *mockConsentClient) FindByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) ([]userConsentDto.UserConsentResponseDTO, error) {
	panic("not used")
}

func (m *mockConsentClient) FindLatestByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	panic("not used")
}

func (m *mockConsentClient) HasAccepted(ctx context.Context, userID, consentDocumentID string, version int, token, tenantId, correlationID string) (bool, error) {
	panic("not used")
}

func (m *mockConsentClient) AcceptAll(ctx context.Context, req userConsentDto.UserConsentAcceptAllRequestDTO, tenantId, correlationID string) (userConsentDto.UserConsentAcceptAllResponseDTO, error) {
	panic("not used")
}

func (m *mockConsentClient) AcceptBatch(ctx context.Context, req userConsentDto.UserConsentAcceptBatchRequestDTO, token, tenantId, correlationID string) (userConsentDto.UserConsentAcceptAllResponseDTO, error) {
	args := m.Called(ctx, req, token, tenantId, correlationID)
	return args.Get(0).(userConsentDto.UserConsentAcceptAllResponseDTO), args.Error(1)
}

func (m *mockConsentClient) DeleteAllByUserId(ctx context.Context, userID, tenantId, correlationID string) error {
	panic("not used")
}

func TestAcceptBatchHandler_Success(t *testing.T) {
	e := echo.New()
	body := `{
		"userId": "other-user",
		"email": "rafael@exemplo.com",
		"acceptedAt": "2026-08-28T14:00:00.000Z",
		"consents": [{"documentId": "doc-1", "version": 1, "accepted": true, "contentHash": "abc"}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-consents/accept-batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	req.Header.Set("X-Public-IP", "189.45.12.8")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1"})

	mockClient := new(mockConsentClient)
	mockClient.On("AcceptBatch", mock.Anything, mock.MatchedBy(func(req userConsentDto.UserConsentAcceptBatchRequestDTO) bool {
		return req.UserID == "user-sub-1" &&
			req.Email == "rafael@exemplo.com" &&
			req.ClientIP == "189.45.12.8" &&
			len(req.Consents) == 1 &&
			req.Consents[0].DocumentID == "doc-1"
	}), "jwt-token", "tenant-1", "corr-1").
		Return(userConsentDto.UserConsentAcceptAllResponseDTO{TotalAccepted: 1}, nil)

	h := NewConsentHandlers(mockClient, zap.NewNop())
	err := h.AcceptBatchHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp userConsentDto.UserConsentAcceptAllResponseDTO
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.TotalAccepted)
	mockClient.AssertExpectations(t)
}

func TestAcceptBatchHandler_UnauthorizedWithoutUser(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-consents/accept-batch", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewConsentHandlers(new(mockConsentClient), zap.NewNop())
	err := h.AcceptBatchHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAcceptBatchHandler_RejectsEmptyConsents(t *testing.T) {
	e := echo.New()
	body := `{"email":"rafael@exemplo.com","consents":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-consents/accept-batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1"})

	h := NewConsentHandlers(new(mockConsentClient), zap.NewNop())
	err := h.AcceptBatchHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAcceptBatchHandler_RejectsMissingEmail(t *testing.T) {
	e := echo.New()
	body := `{"consents":[{"documentId":"doc-1","version":1,"accepted":true}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-consents/accept-batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1"})

	h := NewConsentHandlers(new(mockConsentClient), zap.NewNop())
	err := h.AcceptBatchHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
