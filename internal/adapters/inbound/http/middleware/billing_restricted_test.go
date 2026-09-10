package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domainclient "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
)

type mockBillingClientForMiddleware struct {
	rawResponse json.RawMessage
	err         error
}

func (m *mockBillingClientForMiddleware) GetEntitlement(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return m.rawResponse, m.err
}
func (m *mockBillingClientForMiddleware) GetCompanyEntitlement(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return m.rawResponse, m.err
}
func (m *mockBillingClientForMiddleware) ListPlans(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) SavePlan(ctx context.Context, scope domainclient.BillingScope, body any) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) PatchPlan(ctx context.Context, scope domainclient.BillingScope, code string, body any) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) GetGatewayAccount(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) PutGatewayAccount(ctx context.Context, scope domainclient.BillingScope, body any) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) ListGatewayAccounts(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) PutGatewayAccountByGateway(ctx context.Context, scope domainclient.BillingScope, gateway string, body any) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) SetPrimaryGateway(ctx context.Context, scope domainclient.BillingScope, gateway string) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) GetSubscription(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) CreateSubscription(ctx context.Context, scope domainclient.BillingScope, body any) (json.RawMessage, int, error) {
	return nil, 0, nil
}
func (m *mockBillingClientForMiddleware) CancelSubscription(ctx context.Context, scope domainclient.BillingScope, id string) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) ListInvoices(ctx context.Context, scope domainclient.BillingScope, query map[string]string) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) ListEntitlements(ctx context.Context, scope domainclient.BillingScope, query map[string]string) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) GetInvoice(ctx context.Context, scope domainclient.BillingScope, id string) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockBillingClientForMiddleware) ForwardAsaasWebhook(ctx context.Context, accessToken string, body []byte) (json.RawMessage, int, error) {
	return nil, 0, nil
}
func (m *mockBillingClientForMiddleware) ForwardStripeWebhook(ctx context.Context, token string, body []byte) (json.RawMessage, int, error) {
	return nil, 0, nil
}

func TestRequireBillingPlatformNotRestricted_BlocksAdminWhenRestricted(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/agents", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &pkg.JWTClaims{
		TenantId:    "company-123",
		UserID:      "user-456",
		Roles:       []string{"ADMIN"},
		Authorities: []string{"collector:write"},
	}
	c.Set("claims", claims)

	mockClient := &mockBillingClientForMiddleware{
		rawResponse: json.RawMessage(`{"status":"restricted","allowsWrite":false,"allowsIngest":false}`),
	}

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	}

	mw := RequireBillingPlatformNotRestricted(mockClient)
	err := mw(handler)(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handlerCalled {
		t.Errorf("expected handler not to be called for restricted company")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestRequireBillingPlatformNotRestricted_AllowsSystemBypass(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/agents", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &pkg.JWTClaims{
		TenantId:    "company-123",
		UserID:      "system-001",
		Roles:       []string{"SYSTEM"},
		Authorities: []string{"collector:write"},
	}
	c.Set("claims", claims)

	mockClient := &mockBillingClientForMiddleware{
		rawResponse: json.RawMessage(`{"status":"restricted","allowsWrite":false,"allowsIngest":false}`),
	}

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	}

	mw := RequireBillingPlatformNotRestricted(mockClient)
	err := mw(handler)(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !handlerCalled {
		t.Errorf("expected SYSTEM role to bypass restriction")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRequireBillingPlatformNotRestricted_AllowsWhenActive(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/agents", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &pkg.JWTClaims{
		TenantId:    "company-123",
		UserID:      "admin-456",
		Roles:       []string{"ADMIN"},
		Authorities: []string{"collector:write"},
	}
	c.Set("claims", claims)

	mockClient := &mockBillingClientForMiddleware{
		rawResponse: json.RawMessage(`{"status":"active","allowsWrite":true,"allowsIngest":true}`),
	}

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	}

	mw := RequireBillingPlatformNotRestricted(mockClient)
	err := mw(handler)(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !handlerCalled {
		t.Errorf("expected handler to be called for active company")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
