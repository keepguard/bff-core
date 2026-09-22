package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockConsentPort struct {
	mock.Mock
}

func (m *mockConsentPort) AcceptBatch(ctx context.Context, cmd appdto.AcceptBatchConsentCommand) (appdto.UserConsentAcceptAllViewDTO, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(appdto.UserConsentAcceptAllViewDTO), args.Error(1)
}

type mockConsentDocumentPort struct {
	mock.Mock
}

func (m *mockConsentDocumentPort) ListPublished(ctx context.Context, query appdto.ListPublishedConsentsQuery) ([]appdto.ConsentDocumentViewDTO, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]appdto.ConsentDocumentViewDTO), args.Error(1)
}

func (m *mockConsentDocumentPort) GetLatestByType(ctx context.Context, query appdto.GetLatestConsentQuery) (appdto.ConsentDocumentViewDTO, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(appdto.ConsentDocumentViewDTO), args.Error(1)
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

	consents := new(mockConsentPort)
	consents.On("AcceptBatch", mock.Anything, mock.MatchedBy(func(cmd appdto.AcceptBatchConsentCommand) bool {
		return cmd.UserID == "user-sub-1" &&
			cmd.Email == "rafael@exemplo.com" &&
			cmd.ClientIP == "189.45.12.8" &&
			cmd.Token == "jwt-token" &&
			cmd.TenantID == "tenant-1" &&
			len(cmd.Consents) == 1 &&
			cmd.Consents[0].DocumentID == "doc-1"
	})).Return(appdto.UserConsentAcceptAllViewDTO{TotalAccepted: 1}, nil)

	h := NewConsentHandlers(consents, new(mockConsentDocumentPort), zap.NewNop())
	err := h.AcceptBatchHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp appdto.UserConsentAcceptAllViewDTO
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.TotalAccepted)
	consents.AssertExpectations(t)
}

func TestAcceptBatchHandler_UnauthorizedWithoutUser(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-consents/accept-batch", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewConsentHandlers(new(mockConsentPort), new(mockConsentDocumentPort), zap.NewNop())
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

	h := NewConsentHandlers(new(mockConsentPort), new(mockConsentDocumentPort), zap.NewNop())
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

	h := NewConsentHandlers(new(mockConsentPort), new(mockConsentDocumentPort), zap.NewNop())
	err := h.AcceptBatchHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
