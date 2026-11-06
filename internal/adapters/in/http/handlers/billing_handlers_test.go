package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

	reactivateBody any
	forwarded      []forwardedCall
	forwardStatus  int
	forwardErr     error
}

type forwardedCall struct {
	scope port.BillingScope
	call  port.BillingCall
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
func (s *stubBillingClient) DeletePlan(context.Context, port.BillingScope, string) error {
	return nil
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
func (s *stubBillingClient) GrantLifetimeSubscription(context.Context, port.BillingScope, any) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"active","planCode":"VIP","interval":"lifetime"}`), nil
}
func (s *stubBillingClient) CancelSubscription(context.Context, port.BillingScope, string, bool) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"canceled"}`), nil
}
func (s *stubBillingClient) Forward(_ context.Context, scope port.BillingScope, call port.BillingCall) (port.BillingResponse, error) {
	s.forwarded = append(s.forwarded, forwardedCall{scope: scope, call: call})
	if s.forwardErr != nil {
		return port.BillingResponse{}, s.forwardErr
	}
	status := s.forwardStatus
	if status == 0 {
		status = http.StatusOK
	}
	return port.BillingResponse{Status: status, Body: json.RawMessage(`{"ok":true}`)}, nil
}
func (s *stubBillingClient) ReactivateSubscription(_ context.Context, _ port.BillingScope, _ string, body any) (json.RawMessage, error) {
	s.reactivateBody = body
	return json.RawMessage(`{"status":"active","cancelAtPeriodEnd":false}`), nil
}
func (s *stubBillingClient) PreviewPlanChange(context.Context, port.BillingScope, string, map[string]string) (json.RawMessage, error) {
	return json.RawMessage(`{"type":"upgrade","chargeCents":1000,"creditCents":900}`), nil
}
func (s *stubBillingClient) ChangePlan(context.Context, port.BillingScope, string, any) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"active"}`), nil
}
func (s *stubBillingClient) ClearScheduledChange(context.Context, port.BillingScope, string) (json.RawMessage, error) {
	return json.RawMessage(`{"status":"active"}`), nil
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

func TestCreateSubscriptionAllowsCreditCardWithoutToken(t *testing.T) {
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month","paymentMethod":"credit_card"}`, &pkg.JWTClaims{
		UserID: "user-1", Roles: []string{"USER"},
	})
	stub := &stubBillingClient{created: json.RawMessage(`{"status":"active"}`), createdCode: 201}
	h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if stub.createCalls != 1 {
		t.Fatalf("esperado 1 chamada para CreateSubscription, obteve %d", stub.createCalls)
	}
	bodyMap, ok := stub.createBody.(map[string]any)
	if !ok {
		t.Fatalf("corpo esperado map[string]any, obteve %T", stub.createBody)
	}
	if bodyMap["paymentMethod"] != "credit_card" {
		t.Fatalf("paymentMethod esperado credit_card, obteve %v", bodyMap["paymentMethod"])
	}
	if _, hasToken := bodyMap["creditCardToken"]; hasToken {
		t.Fatalf("não deve ter creditCardToken no body")
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
	user           appdto.MSUserResponseDTO
	patched        appdto.MSUserResponseDTO
	patchErr       error
	gets           int
	patches        int
	patchCPF       string
	getByEmailUser appdto.UserByEmailResponseDTO
	getByEmailErr  error
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
	if s.getByEmailErr != nil {
		return appdto.UserByEmailResponseDTO{}, s.getByEmailErr
	}
	return s.getByEmailUser, nil
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

func TestGrantLifetimeBillingSubscriptionHandler(t *testing.T) {
	billing := &stubBillingClient{}
	c, rec := billingContext(http.MethodPost, "/core/billing/subscriptions/grant-lifetime",
		`{"targetUserId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","planCode":"VIP"}`,
		&pkg.JWTClaims{
			TenantId: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			UserID:   "cccccccc-cccc-cccc-cccc-cccccccccccc",
			Roles:    []string{"ADMIN"},
		})

	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop())
	if err := h.GrantLifetimeBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
}

func TestLookupBillingUserHandler(t *testing.T) {
	billing := &stubBillingClient{}
	claims := &pkg.JWTClaims{
		TenantId: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		UserID:   "cccccccc-cccc-cccc-cccc-cccccccccccc",
		Roles:    []string{"ADMIN"},
	}

	t.Run("success", func(t *testing.T) {
		users := &stubUserClient{
			getByEmailUser: appdto.UserByEmailResponseDTO{
				ID:            "11111111-2222-3333-4444-555555555555",
				CodeUser:      "USR-001",
				Username:      "John Doe",
				Email:         "john@example.com",
				Status:        "ACTIVE",
				EmailVerified: true,
			},
		}
		c, rec := billingContext(http.MethodGet, "/core/billing/users/lookup?email=john@example.com", "", claims)
		h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(users)

		if err := h.LookupBillingUserHandler(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp["id"] != "USR-001" || resp["email"] != "john@example.com" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("missing_query", func(t *testing.T) {
		users := &stubUserClient{}
		c, rec := billingContext(http.MethodGet, "/core/billing/users/lookup", "", claims)
		h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(users)

		if err := h.LookupBillingUserHandler(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("user_not_found", func(t *testing.T) {
		users := &stubUserClient{
			getByEmailErr: fmt.Errorf("not found"),
		}
		c, rec := billingContext(http.MethodGet, "/core/billing/users/lookup?q=unknown@example.com", "", claims)
		h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(users)

		if err := h.LookupBillingUserHandler(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestCreateSubscriptionAttachesSessionSnapshot(t *testing.T) {
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month"}`, &pkg.JWTClaims{
		UserID: "user-1", Roles: []string{"USER"}, DeviceID: "dev_claim_1", JTI: "jti_token_1",
	})
	c.Request().Header.Set("X-Device-Id", "dev_header_1")
	c.Request().Header.Set("X-Device-Name", "Chrome no Mac")
	c.Request().Header.Set("X-Device-Type", "DESKTOP")
	c.Request().Header.Set("X-Real-IP", "189.10.20.30")
	c.Request().Header.Set("User-Agent", "GoTestBrowser/1.0")
	c.Request().Header.Set("X-Session-Id", "sess_header_1")

	client := &stubBillingClient{}
	h := NewBillingHandlers(appbilling.NewBillingPort(client), zap.NewNop())
	if err := h.CreateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	bodyMap, ok := client.createBody.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", client.createBody)
	}

	if bodyMap["deviceId"] != "dev_header_1" {
		t.Fatalf("expected deviceId dev_header_1, got %v", bodyMap["deviceId"])
	}
	if bodyMap["deviceName"] != "Chrome no Mac" {
		t.Fatalf("expected deviceName Chrome no Mac, got %v", bodyMap["deviceName"])
	}
	if bodyMap["deviceType"] != "DESKTOP" {
		t.Fatalf("expected deviceType DESKTOP, got %v", bodyMap["deviceType"])
	}
	if bodyMap["ipAddress"] != "189.10.20.30" {
		t.Fatalf("expected ipAddress 189.10.20.30, got %v", bodyMap["ipAddress"])
	}
	if bodyMap["userAgent"] != "GoTestBrowser/1.0" {
		t.Fatalf("expected userAgent GoTestBrowser/1.0, got %v", bodyMap["userAgent"])
	}
	if bodyMap["sessionId"] != "sess_header_1" {
		t.Fatalf("expected sessionId sess_header_1, got %v", bodyMap["sessionId"])
	}
}

func TestDeleteBillingPlanHandler_Success(t *testing.T) {
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{}), zap.NewNop())
	c, rec := billingContext(http.MethodDelete, "/api/v1/core/billing/plans/pro-test", "", &pkg.JWTClaims{
		UserID:   "user-admin",
		TenantId: "company-1",
		Roles:    []string{"ROLE_ADMIN"},
	})
	c.SetParamNames("code")
	c.SetParamValues("pro-test")

	if err := h.DeleteBillingPlanHandler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestReactivateBillingSubscriptionHandler_Success(t *testing.T) {
	h := NewBillingHandlers(appbilling.NewBillingPort(&stubBillingClient{}), zap.NewNop())
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions/sub-123/reactivate", "", &pkg.JWTClaims{
		UserID:      "user-1",
		TenantId:    "company-1",
		Roles:       []string{"ROLE_USER"},
		Authorities: []string{"billing:read"},
	})
	c.SetParamNames("id")
	c.SetParamValues("sub-123")

	if err := h.ReactivateBillingSubscriptionHandler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"cancelAtPeriodEnd":false`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

// K5 e §6.12: dado de cartão recusado em qualquer nível; o token segue aceito neste caminho.
func TestContainsCardData(t *testing.T) {
	refused := []string{
		`{"cardNumber":"4111"}`, `{"card_number":"4111"}`, `{"card-number":"4111"}`,
		`{"creditCardNumber":"4111"}`, `{"credit_card_number":"4111"}`,
		`{"creditCard":{"number":"4111"}}`, `{"credit_card":{"number":"4111"}}`, `{"card":{"number":"4111"}}`,
		`{"creditCard":{"holderName":"X"}}`,
		`{"cvv":"123"}`, `{"cvc":"123"}`, `{"ccv":"123"}`, `{"pan":"4111"}`,
		`{"expiryMonth":"12"}`, `{"expiryYear":"2030"}`, `{"expirationDate":"2030-12"}`, `{"expiry_month":"12"}`,
		`{"creditCardHolderInfo":{"name":"X"}}`,
		`{"nested":{"list":[{"ccv":"123"}]}}`,
	}
	for _, raw := range refused {
		var body map[string]any
		if err := json.Unmarshal([]byte(raw), &body); err != nil {
			t.Fatal(err)
		}
		if !containsCardData(body) {
			t.Fatalf("precisa recusar %s", raw)
		}
	}
	accepted := []string{
		`{"creditCardToken":"tok"}`, `{"planCode":"PRO","interval":"month","paymentMethod":"credit_card"}`,
		`{"number":"123"}`, `{"invoice":{"number":"123"}}`, `{"expiresAt":"2030-01-01T00:00:00Z"}`,
		`{"cardDebitConsent":{"textVersion":"v1"}}`, `{"prices":[{"interval":"month","amountCents":9700}]}`,
	}
	for _, raw := range accepted {
		var body map[string]any
		if err := json.Unmarshal([]byte(raw), &body); err != nil {
			t.Fatal(err)
		}
		if containsCardData(body) {
			t.Fatalf("não pode recusar %s", raw)
		}
	}
}

func TestCreateSubscriptionRefusesNewCardFields(t *testing.T) {
	for _, field := range []string{`"creditCardNumber":"4111"`, `"creditCard":{"number":"4111"}`, `"ccv":"123"`, `"expiryMonth":"12"`, `"expiryYear":"2030"`, `"expirationDate":"2030-12"`, `"creditCardHolderInfo":{"name":"X"}`} {
		t.Run(field, func(t *testing.T) {
			c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic",`+field+`}`, &pkg.JWTClaims{UserID: "user-1", Roles: []string{"ROLE_USER"}})
			stub := &stubBillingClient{}
			h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
			if err := h.CreateBillingSubscriptionHandler(c); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "PAYMENT_METHOD_UNSUPPORTED") {
				t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
			}
			if stub.createCalls != 0 {
				t.Fatal("não pode chegar ao ms-billing")
			}
		})
	}
}

func TestReactivateForwardsOptionalBody(t *testing.T) {
	newContext := func(body string) (echo.Context, *httptest.ResponseRecorder) {
		c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions/sub-123/reactivate", body, &pkg.JWTClaims{UserID: "user-1", Roles: []string{"ROLE_USER"}})
		c.SetParamNames("id")
		c.SetParamValues("sub-123")
		return c, rec
	}
	stub := &stubBillingClient{}
	h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	c, rec := newContext("")
	if err := h.ReactivateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || stub.reactivateBody != nil {
		t.Fatalf("sem corpo fica igual: %d %v", rec.Code, stub.reactivateBody)
	}
	c, rec = newContext(`{"cardDebitConsent":{"textVersion":"v1"}}`)
	if err := h.ReactivateBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	body, ok := stub.reactivateBody.(map[string]any)
	if rec.Code != http.StatusOK || !ok || body["cardDebitConsent"] == nil {
		t.Fatalf("aceite precisa seguir: %d %v", rec.Code, stub.reactivateBody)
	}
}

func TestIdempotencyKeyReachesScope(t *testing.T) {
	stub := &stubBillingClient{}
	h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	c, _ := billingContext(http.MethodGet, "/api/v1/core/billing/credits", "", &pkg.JWTClaims{UserID: "user-1", Roles: []string{"ROLE_USER"}})
	c.Request().Header.Set("Idempotency-Key", " chave-1 ")
	if err := h.RemoveBillingCardHandler(c); err != nil {
		t.Fatal(err)
	}
	if len(stub.forwarded) != 1 || stub.forwarded[0].scope.IdempotencyKey != "chave-1" {
		t.Fatalf("escopo %+v", stub.forwarded)
	}
	if stub.forwarded[0].scope.OperatorID != "" {
		t.Fatal("rota de assinante não carrega operador")
	}
}

func TestOperatorScopeUsesTokenSubject(t *testing.T) {
	stub := &stubBillingClient{forwardStatus: http.StatusCreated}
	h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/credits", `{"amountCents":99999999999,"reason":"x"}`, &pkg.JWTClaims{Sub: "sub-do-token", UserID: "user-1", Roles: []string{"ROLE_ADMIN"}})
	c.Request().RemoteAddr = "200.200.200.3:1234"
	if err := h.GrantBillingCreditHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d", rec.Code)
	}
	got := stub.forwarded[0]
	if got.scope.OperatorID != "sub-do-token" || got.scope.ClientIP != "200.200.200.3" {
		t.Fatalf("escopo %+v", got.scope)
	}
	body := got.call.Body.(map[string]any)
	if body["amountCents"].(json.Number).String() != "99999999999" {
		t.Fatalf("centavos precisam ir intactos: %v", body["amountCents"])
	}
}

func TestForwardedErrorKeepsCode(t *testing.T) {
	stub := &stubBillingClient{forwardErr: &appdto.HTTPError{Code: http.StatusTooManyRequests, Message: "Muitas requisições", Details: `{"error":"DAILY_CAP_EXCEEDED","message":"x"}`}}
	h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/credits", `{"amountCents":1}`, &pkg.JWTClaims{Sub: "s", UserID: "user-1", Roles: []string{"ROLE_ADMIN"}})
	if err := h.GrantBillingCreditHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusTooManyRequests || !strings.Contains(rec.Body.String(), "DAILY_CAP_EXCEEDED") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
