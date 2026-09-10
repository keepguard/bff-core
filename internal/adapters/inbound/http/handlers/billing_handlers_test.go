package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appbilling "github.com/keepguard/bff-core/internal/application/billing"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type stubBillingClient struct {
	entitlement json.RawMessage
	created     json.RawMessage
	createdCode int
	createCalls int
	createBody  any
	webhookBody []byte
	webhookTok  string
}

func (s *stubBillingClient) GetEntitlement(context.Context, port.BillingScope) (json.RawMessage, error) {
	if s.entitlement != nil {
		return s.entitlement, nil
	}
	return json.RawMessage(`{"status":"none","allowsProduct":false}`), nil
}
func (s *stubBillingClient) GetCompanyEntitlement(context.Context, port.BillingScope) (json.RawMessage, error) {
	if s.entitlement != nil {
		return s.entitlement, nil
	}
	return json.RawMessage(`{"status":"active","allowsProduct":true,"allowsWrite":true,"allowsIngest":true}`), nil
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
func (s *stubBillingClient) CreateSubscription(_ context.Context, _ port.BillingScope, body any) (json.RawMessage, int, error) {
	s.createCalls++
	s.createBody = body
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
func (s *stubBillingClient) ForwardStripeWebhook(_ context.Context, token string, body []byte) (json.RawMessage, int, error) {
	s.webhookTok = token
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

func TestCreateSubscriptionAllowsAdminPayer(t *testing.T) {
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month","paymentMethod":"pix"}`, &pkg.JWTClaims{
		UserID: "admin-1", Roles: []string{"ROLE_ADMIN"},
	})
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{}), zap.NewNop())
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
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

const testValidCPF = "52998224725"

type stubUserClient struct {
	user     appdto.MSUserResponseDTO
	patched  appdto.MSUserResponseDTO
	patchErr error
	gets     int
	patches  int
	patchCPF string
}

func (s *stubUserClient) CreateUser(context.Context, appdto.MSUserCreateRequestDTO, string, string) (appdto.MSUserResponseDTO, error) {
	return appdto.MSUserResponseDTO{}, nil
}
func (s *stubUserClient) GetUserByCodeUser(context.Context, string, string, string, string) (appdto.MSUserResponseDTO, error) {
	s.gets++
	if s.patches > 0 && s.patched.ID != "" {
		return s.patched, nil
	}
	return s.user, nil
}
func (s *stubUserClient) GetByEmail(context.Context, string, string, string, string) (appdto.UserByEmailResponseDTO, error) {
	return appdto.UserByEmailResponseDTO{}, nil
}
func (s *stubUserClient) PatchPersonDocument(_ context.Context, userID, cpf, _, _, _ string) (appdto.MSUserResponseDTO, error) {
	s.patches++
	s.patchCPF = cpf
	if s.patchErr != nil {
		return appdto.MSUserResponseDTO{}, s.patchErr
	}
	if s.patched.ID != "" {
		return s.patched, nil
	}
	out := s.user
	out.ID = userID
	out.PersonProfile = &appdto.PersonProfileDTO{CPF: cpf, FullName: "Nome"}
	s.patched = out
	return out, nil
}
func (s *stubUserClient) CreateUserNotify(context.Context, appdto.MSUserNotifyCreateRequestDTO, string, string) (appdto.MSUserNotifyResponseDTO, error) {
	return appdto.MSUserNotifyResponseDTO{}, nil
}
func (s *stubUserClient) InitRegister(context.Context, appdto.MSUserRegisterInitRequestDTO, string, string) (appdto.MSUserRegisterInitResponseDTO, error) {
	return appdto.MSUserRegisterInitResponseDTO{}, nil
}
func (s *stubUserClient) ConfirmRegister(context.Context, appdto.MSUserRegisterConfirmRequestDTO, string, string) (appdto.MSUserRegisterConfirmResponseDTO, error) {
	return appdto.MSUserRegisterConfirmResponseDTO{}, nil
}
func (s *stubUserClient) DeleteUser(context.Context, string, string, string) error {
	return nil
}
func (s *stubUserClient) ResendRegisterToken(context.Context, appdto.MSUserRegisterResendRequestDTO, string, string) (appdto.MSUserRegisterResendResponseDTO, error) {
	return appdto.MSUserRegisterResendResponseDTO{}, nil
}

func TestCreateSubscriptionMissingCpf(t *testing.T) {
	billing := &stubBillingClient{}
	users := &stubUserClient{user: appdto.MSUserResponseDTO{
		ID: "ms-user-1", Email: "a@b.com", PersonProfile: &appdto.PersonProfileDTO{FullName: "Nome"},
	}}
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month","paymentMethod":"pix"}`, &pkg.JWTClaims{
		UserID: "user-1", Roles: []string{"ROLE_USER"},
	})
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(users)
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "PAYER_DOCUMENT_MISSING") {
		t.Fatalf("expected PAYER_DOCUMENT_MISSING, got %s", rec.Body.String())
	}
	if billing.createCalls != 0 || users.patches != 0 {
		t.Fatalf("billing=%d patch=%d", billing.createCalls, users.patches)
	}
}

func TestCreateSubscriptionIgnoresClientCpfWhenProfileHasDocument(t *testing.T) {
	billing := &stubBillingClient{}
	users := &stubUserClient{user: appdto.MSUserResponseDTO{
		ID: "ms-user-1", Email: "a@b.com",
		PersonProfile: &appdto.PersonProfileDTO{FullName: "Nome", CPF: testValidCPF},
	}}
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions",
		`{"planCode":"basic","interval":"month","paymentMethod":"pix","payerCpfCnpj":"39053344705"}`,
		&pkg.JWTClaims{UserID: "user-1", Roles: []string{"ROLE_USER"}})
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(users)
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if users.patches != 0 {
		t.Fatalf("patch calls %d", users.patches)
	}
	body, _ := billing.createBody.(map[string]any)
	if body["payerCpfCnpj"] != testValidCPF {
		t.Fatalf("payerCpfCnpj %v", body["payerCpfCnpj"])
	}
}

func TestCreateSubscriptionUniqueCpfConflict(t *testing.T) {
	billing := &stubBillingClient{}
	users := &stubUserClient{
		user: appdto.MSUserResponseDTO{
			ID: "ms-user-1", Email: "a@b.com", PersonProfile: &appdto.PersonProfileDTO{FullName: "Nome"},
		},
		patchErr: &appdto.HTTPError{
			Code:    http.StatusConflict,
			Message: "CPF já está em uso nesta empresa",
			Details: `{"errorCode":"CPF_ALREADY_EXISTS","message":"CPF já está em uso nesta empresa"}`,
		},
	}
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions",
		`{"planCode":"basic","interval":"month","paymentMethod":"pix","payerCpfCnpj":"529.982.247-25"}`,
		&pkg.JWTClaims{UserID: "user-1", Roles: []string{"ROLE_USER"}})
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(users)
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "CPF_ALREADY_EXISTS") {
		t.Fatalf("expected CPF_ALREADY_EXISTS, got %s", rec.Body.String())
	}
	if billing.createCalls != 0 {
		t.Fatalf("billing calls %d", billing.createCalls)
	}
}

func TestCreateSubscriptionFirstWriteCpf(t *testing.T) {
	billing := &stubBillingClient{}
	users := &stubUserClient{user: appdto.MSUserResponseDTO{
		ID: "ms-user-1", Email: "a@b.com", PersonProfile: &appdto.PersonProfileDTO{FullName: "Nome"},
	}}
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions",
		`{"planCode":"basic","interval":"month","paymentMethod":"pix","payerCpfCnpj":"529.982.247-25"}`,
		&pkg.JWTClaims{UserID: "user-1", Roles: []string{"ROLE_USER"}})
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(users)
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if users.patches != 1 || users.patchCPF != testValidCPF {
		t.Fatalf("patch calls=%d cpf=%s", users.patches, users.patchCPF)
	}
	body, _ := billing.createBody.(map[string]any)
	if body["payerCpfCnpj"] != testValidCPF {
		t.Fatalf("payerCpfCnpj %v", body["payerCpfCnpj"])
	}
}

func TestStripeWebhookHandler_Forwarded(t *testing.T) {
	billing := &stubBillingClient{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/billing/webhooks/stripe", strings.NewReader(`{"id":"evt_123","type":"invoice.payment_succeeded"}`))
	req.Header.Set("stripe-signature", "t=1700000000,v1=signature_hash")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop())
	if err := h.StripeWebhookHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if billing.webhookTok != "t=1700000000,v1=signature_hash" {
		t.Fatalf("expected token forwarded, got %s", billing.webhookTok)
	}
}

