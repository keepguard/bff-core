package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockUserPort struct {
	mock.Mock
}

func (m *mockUserPort) GetMe(ctx context.Context, query appdto.GetMeQuery) (appdto.MeProfileViewDTO, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(appdto.MeProfileViewDTO), args.Error(1)
}

func TestGetMeHandler_UsesSubAndReturnsProfile(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1", TenantId: "tenant-1"})

	users := new(mockUserPort)
	users.On("GetMe", mock.Anything, appdto.GetMeQuery{
		TenantID:      "tenant-1",
		CorrelationID: "corr-1",
		Token:         "jwt-token",
		CodeUser:      "user-sub-1",
	}).Return(appdto.MeProfileViewDTO{
		Email:         "rafael@exemplo.com",
		DisplayHandle: "rafael.soares",
		PhoneE164:     "+5511999999999",
		Type:          "PERSON",
		Status:        "ACTIVE",
		PersonProfile: &appdto.MePersonProfileViewDTO{FullName: "Rafael Soares"},
	}, nil)

	h := NewUserHandlers(users, zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body inboundDto.MeProfileResponseDTO
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "rafael@exemplo.com", body.Email)
	assert.Equal(t, "rafael.soares", body.DisplayHandle)
	assert.Equal(t, "Rafael Soares", body.PersonProfile.FullName)
	users.AssertExpectations(t)
}

func TestGetMeHandler_UnauthorizedWhenMissingSub(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{})

	users := new(mockUserPort)
	h := NewUserHandlers(users, zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	users.AssertNotCalled(t, "GetMe")
}

func TestGetMeHandler_Propagates403(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1", TenantId: "tenant-1"})

	users := new(mockUserPort)
	users.On("GetMe", mock.Anything, mock.Anything).
		Return(appdto.MeProfileViewDTO{}, &appdto.HTTPError{Code: http.StatusForbidden, Message: "Sem permissão"})

	h := NewUserHandlers(users, zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	users.AssertExpectations(t)
}

func TestGetMeHandler_Returns404WhenCompanyMissing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1", TenantId: "tenant-1"})

	users := new(mockUserPort)
	users.On("GetMe", mock.Anything, mock.Anything).
		Return(appdto.MeProfileViewDTO{}, pkg.NewAppError("COMPANY_NOT_FOUND", "Empresa não encontrada para o tenant informado", http.StatusNotFound))

	h := NewUserHandlers(users, zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetMeHandler_UsesTenantFromJWTWithoutHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1", TenantId: "tenant-1"})

	users := new(mockUserPort)
	users.On("GetMe", mock.Anything, appdto.GetMeQuery{
		TenantID:      "tenant-1",
		CorrelationID: "corr-1",
		Token:         "jwt-token",
		CodeUser:      "user-sub-1",
	}).Return(appdto.MeProfileViewDTO{
		Email:         "rafael@exemplo.com",
		DisplayHandle: "rafael.soares",
		Type:          "PERSON",
		Status:        "ACTIVE",
	}, nil)

	h := NewUserHandlers(users, zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	users.AssertExpectations(t)
}
