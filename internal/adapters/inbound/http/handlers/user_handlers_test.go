package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	companydecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/company"
	userdecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/user"
	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func stubCompanyClient() *companydecorator.MockCompanyClient {
	mockCompany := new(companydecorator.MockCompanyClient)
	mockCompany.On("GetByTenantId", mock.Anything, "tenant-1", "corr-1").
		Return(companyDto.MSCompanyResponseDTO{ID: "company-1"}, nil)
	return mockCompany
}

func TestGetMeHandler_UsesSubAndReturnsProfile(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1"})

	mockUser := new(userdecorator.MockUserClient)
	mockUser.On("GetUserByCodeUser", mock.Anything, "user-sub-1", "jwt-token", "tenant-1", "corr-1").
		Return(userDto.MSUserResponseDTO{
			Email:         "rafael@exemplo.com",
			DisplayHandle: "rafael.soares",
			PhoneE164:     "+5511999999999",
			Type:          "PERSON",
			Status:        "ACTIVE",
			PersonProfile: &userDto.PersonProfileDTO{FullName: "Rafael Soares"},
		}, nil)

	h := NewUserHandlers(mockUser, stubCompanyClient(), zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body inboundDto.MeProfileResponseDTO
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "rafael@exemplo.com", body.Email)
	assert.Equal(t, "rafael.soares", body.DisplayHandle)
	assert.Equal(t, "Rafael Soares", body.PersonProfile.FullName)
	mockUser.AssertExpectations(t)
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

	mockUser := new(userdecorator.MockUserClient)
	mockCompany := new(companydecorator.MockCompanyClient)
	h := NewUserHandlers(mockUser, mockCompany, zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockUser.AssertNotCalled(t, "GetUserByCodeUser")
	mockCompany.AssertNotCalled(t, "GetByTenantId")
}

func TestGetMeHandler_Propagates403(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1"})

	mockUser := new(userdecorator.MockUserClient)
	mockUser.On("GetUserByCodeUser", mock.Anything, "user-sub-1", "jwt-token", "tenant-1", "corr-1").
		Return(userDto.MSUserResponseDTO{}, &appdto.HTTPError{Code: http.StatusForbidden, Message: "Sem permissão"})

	h := NewUserHandlers(mockUser, stubCompanyClient(), zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockUser.AssertExpectations(t)
}

func TestGetMeHandler_Returns404WhenCompanyMissing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-Tenant-Id", "tenant-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("token", "jwt-token")
	c.Set("claims", &pkg.JWTClaims{Sub: "user-sub-1"})

	mockUser := new(userdecorator.MockUserClient)
	mockCompany := new(companydecorator.MockCompanyClient)
	mockCompany.On("GetByTenantId", mock.Anything, "tenant-1", "corr-1").
		Return(companyDto.MSCompanyResponseDTO{}, nil)

	h := NewUserHandlers(mockUser, mockCompany, zap.NewNop())
	err := h.GetMeHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockUser.AssertNotCalled(t, "GetUserByCodeUser")
}

func TestToMeProfile_DropsDocumentFields(t *testing.T) {
	raw := `{
		"email":"a@b.com",
		"display_handle":"handle",
		"personProfile":{"full_name":"Nome Completo","cpf":"123.456.789-00","rg":"1122233","mother_name":"Mae"}
	}`
	var user userDto.MSUserResponseDTO
	assert.NoError(t, json.Unmarshal([]byte(raw), &user))

	got := toMeProfile(user)
	encoded, err := json.Marshal(got)
	assert.NoError(t, err)
	payload := strings.ToLower(string(encoded))
	assert.NotContains(t, payload, "cpf")
	assert.NotContains(t, payload, "rg")
	assert.NotContains(t, payload, "mother")
	assert.Equal(t, "Nome Completo", got.PersonProfile.FullName)
}
