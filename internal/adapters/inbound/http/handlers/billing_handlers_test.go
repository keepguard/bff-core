package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appbilling "github.com/keepguard/bff-core/internal/application/billing"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type stubBillingClient struct {
	entitlement json.RawMessage
	created     json.RawMessage
	createdCode int
	webhookBody []byte
	webhookTok  string
}

func (s *stubBillingClient) GetEntitlement(context.Context, port.BillingScope) (json.RawMessage, error) {
	if s.entitlement != nil {
		return s.entitlement, nil
	}
	return json.RawMessage(`{"status":"none","allowsProduct":false}`), nil
}
func (s *stubBillingClient) ListPlans(context.Context, port.BillingScope) (json.RawMessage, error) {
	return json.RawMessage(`[]`), nil
}
func (s *stubBillingClient) SavePlan(context.Context, port.BillingScope, any) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (s *stubBillingClient) PatchPlan(context.Context, port.BillingScope, string, any) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (s *stubBillingClient) GetGatewayAccount(context.Context, port.BillingScope) (json.RawMessage, error) {
	return json.RawMessage(`{"apiKeyMasked":"****1234"}`), nil
}
func (s *stubBillingClient) PutGatewayAccount(context.Context, port.BillingScope, any) (json.RawMessage, error) {
	return json.RawMessage(`{"apiKeyMasked":"****9999"}`), nil
}
func (s *stubBillingClient) ListGatewayAccounts(context.Context, port.BillingScope) (json.RawMessage, error) {
	return json.RawMessage(`{"items":[]}`), nil
}
func (s *stubBillingClient) PutGatewayAccountByGateway(context.Context, port.BillingScope, string, any) (json.RawMessage, error) {
	return json.RawMessage(`{"gateway":"asaas","primary":true,"apiKeyMasked":"****9999"}`), nil
}
func (s *stubBillingClient) SetPrimaryGateway(context.Context, port.BillingScope, string) (json.RawMessage, error) {
	return json.RawMessage(`{"gateway":"asaas","primary":true}`), nil
}
func (s *stubBillingClient) GetSubscription(context.Context, port.BillingScope) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"pending_gateway"}`), nil
}
func (s *stubBillingClient) CreateSubscription(context.Context, port.BillingScope, any) (json.RawMessage, int, error) {
	if s.created != nil {
		code := s.createdCode
		if code == 0 {
			code = 201
		}
		return s.created, code, nil
	}
	return json.RawMessage(`{"status":"active"}`), 201, nil
}
func (s *stubBillingClient) CancelSubscription(context.Context, port.BillingScope, string) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"canceled"}`), nil
}
func (s *stubBillingClient) ListInvoices(context.Context, port.BillingScope, map[string]string) (json.RawMessage, error) {
	return json.RawMessage(`{"items":[],"page":0,"size":20,"totalElements":0,"totalPages":0}`), nil
}
func (s *stubBillingClient) ListEntitlements(context.Context, port.BillingScope, map[string]string) (json.RawMessage, error) {
	return json.RawMessage(`{"items":[],"page":0,"size":20,"totalElements":0,"totalPages":0}`), nil
}
func (s *stubBillingClient) GetInvoice(context.Context, port.BillingScope, string) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"pending"}`), nil
}
func (s *stubBillingClient) ForwardAsaasWebhook(_ context.Context, accessToken string, body []byte) (json.RawMessage, int, error) {
	s.webhookTok = accessToken
	s.webhookBody = append([]byte(nil), body...)
	return json.RawMessage(`{"received":true}`), 200, nil
}

func billingContext(method, path, body string, claims *pkg.JWTClaims) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(path)
	if claims != nil {
		c.Set("claims", claims)
	}
	ctx := port.WithCompanyID(c.Request().Context(), "company-1")
	c.SetRequest(c.Request().WithContext(ctx))
	return c, rec
}

func TestCreateSubscriptionRejectsCardNumber(t *testing.T) {
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month","cardNumber":"4111111111111111"}`, &pkg.JWTClaims{
		UserID: "user-1", Roles: []string{"USER"},
	})
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{}), zap.NewNop())
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "PAYMENT_METHOD_UNSUPPORTED") {
		t.Fatalf("expected PAYMENT_METHOD_UNSUPPORTED, got %s", rec.Body.String())
	}
}

func TestWebhookIgnoresTenantHeaderAndForwardsToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/billing/webhooks/asaas", strings.NewReader(`{"id":"evt_1","event":"PAYMENT_CREATED"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("asaas-access-token", "company-token")
	req.Header.Set("X-Tenant-Id", "tenant-of-attacker")
	req.Header.Set("X-Company-Id", "other-company")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	stub := &stubBillingClient{}
	h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	if err := h.AsaasWebhookHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if stub.webhookTok != "company-token" {
		t.Fatalf("token %s", stub.webhookTok)
	}
	if !strings.Contains(string(stub.webhookBody), "evt_1") {
		t.Fatalf("body %s", stub.webhookBody)
	}
}

func TestWebhookMissingTokenUnauthorized(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/billing/webhooks/asaas", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{}), zap.NewNop())
	if err := h.AsaasWebhookHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestGetEntitlementPassesUserScope(t *testing.T) {
	c, rec := billingContext(http.MethodGet, "/api/v1/core/billing/entitlement", "", &pkg.JWTClaims{
		UserID: "user-1", Roles: []string{"USER"},
	})
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{
		entitlement: json.RawMessage(`{"status":"restricted","allowsProduct":false}`),
	}), zap.NewNop())
	if err := h.GetBillingEntitlementHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"allowsProduct":false`) {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestCreateSubscriptionRejectsOpsCaller(t *testing.T) {
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month","paymentMethod":"pix"}`, &pkg.JWTClaims{
		UserID: "admin-1", Roles: []string{"ROLE_ADMIN"},
	})
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{}), zap.NewNop())
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "BILLING_PAYER_OPS") {
		t.Fatalf("expected BILLING_PAYER_OPS, got %s", rec.Body.String())
	}
}

func TestCreateSubscriptionAllowsUserPayer(t *testing.T) {
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month","paymentMethod":"pix"}`, &pkg.JWTClaims{
		UserID: "user-1", Roles: []string{"ROLE_USER"}, Authorities: []string{"billing:read"},
	})
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{}), zap.NewNop())
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestCreateSubscriptionRejectsManagerOpsCaller(t *testing.T) {
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month","paymentMethod":"pix"}`, &pkg.JWTClaims{
		UserID: "mgr-1", Roles: []string{"ROLE_MANAGER"}, Authorities: []string{"billing:read"},
	})
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{}), zap.NewNop())
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
