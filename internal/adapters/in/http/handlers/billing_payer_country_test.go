package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appbilling "github.com/keepguard/bff-core/internal/application/billing"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// payerCountryCase é uma das 4 rotas em que o país da conta decide a forma de pagar.
type payerCountryCase struct {
	name   string
	method string
	path   string
	body   string
	call   func(h *BillingHandlers, c echo.Context) error
}

func payerCountryCases() []payerCountryCase {
	return []payerCountryCase{
		{"subscriptions", http.MethodPost, "/api/v1/core/billing/subscriptions", `{"planCode":"basic","interval":"month","paymentMethod":"pix"}`,
			func(h *BillingHandlers, c echo.Context) error { return h.CreateBillingSubscriptionHandler(c) }},
		{"reactivate", http.MethodPost, "/api/v1/core/billing/subscriptions/sub-1/reactivate", "",
			func(h *BillingHandlers, c echo.Context) error { return h.ReactivateBillingSubscriptionHandler(c) }},
		{"change-preview", http.MethodGet, "/api/v1/core/billing/subscriptions/sub-1/change-preview?planCode=pro&interval=month", "",
			func(h *BillingHandlers, c echo.Context) error { return h.PreviewBillingPlanChangeHandler(c) }},
		{"change-plan", http.MethodPost, "/api/v1/core/billing/subscriptions/sub-1/change-plan", `{"planCode":"pro","interval":"month"}`,
			func(h *BillingHandlers, c echo.Context) error { return h.ChangeBillingPlanHandler(c) }},
	}
}

func runPayerCountry(t *testing.T, tc payerCountryCase, billing *stubBillingClient, users *stubUserClient, browserCountry string) *httptest.ResponseRecorder {
	t.Helper()
	c, rec := billingContext(tc.method, tc.path, tc.body, &pkg.JWTClaims{UserID: "user-1", CodeUser: "code-1", Roles: []string{"ROLE_USER"}})
	if tc.body == "" {
		c.Request().ContentLength = 0
	}
	c.SetParamNames("id")
	c.SetParamValues("sub-1")
	if browserCountry != "" {
		c.Request().Header.Set("X-Payer-Country", browserCountry)
	}
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop())
	if users != nil {
		h = h.WithUsers(users)
	}
	if err := tc.call(h, c); err != nil {
		t.Fatal(err)
	}
	return rec
}

func countryPtr(v string) *string { return &v }

func TestPayerCountryFromMsUserReachesBilling(t *testing.T) {
	for _, tc := range payerCountryCases() {
		t.Run(tc.name, func(t *testing.T) {
			billing := &stubBillingClient{}
			users := brPayer()
			users.user.CountryCode = countryPtr("pt")
			rec := runPayerCountry(t, tc, billing, users, "")
			if rec.Code >= 300 {
				t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
			}
			if len(billing.payerScopes) != 1 || billing.payerScopes[0].PayerCountry != "PT" {
				t.Fatalf("país esperado PT no escopo: %+v", billing.payerScopes)
			}
		})
	}
}

func TestPayerCountryIgnoresBrowserHeader(t *testing.T) {
	for _, tc := range payerCountryCases() {
		t.Run(tc.name, func(t *testing.T) {
			billing := &stubBillingClient{}
			users := brPayer()
			users.user.CountryCode = countryPtr("PT")
			rec := runPayerCountry(t, tc, billing, users, "BR")
			if rec.Code >= 300 {
				t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
			}
			if len(billing.payerScopes) != 1 || billing.payerScopes[0].PayerCountry != "PT" {
				t.Fatalf("o país do navegador não vale: %+v", billing.payerScopes)
			}
		})
	}
}

func TestPayerCountryNullIsBR(t *testing.T) {
	for _, tc := range payerCountryCases() {
		t.Run(tc.name, func(t *testing.T) {
			billing := &stubBillingClient{}
			rec := runPayerCountry(t, tc, billing, brPayer(), "PT")
			if rec.Code >= 300 {
				t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
			}
			if len(billing.payerScopes) != 1 || billing.payerScopes[0].PayerCountry != "BR" {
				t.Fatalf("null precisa ir como BR: %+v", billing.payerScopes)
			}
		})
	}
}

func TestPayerCountryUnavailableIs503WithoutBilling(t *testing.T) {
	for _, tc := range payerCountryCases() {
		for name, users := range map[string]*stubUserClient{
			"ms-user fora": {getErr: errors.New("connection refused")},
			"sem cliente":  nil,
		} {
			t.Run(tc.name+"/"+name, func(t *testing.T) {
				billing := &stubBillingClient{}
				rec := runPayerCountry(t, tc, billing, users, "BR")
				if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"error":"PAYER_COUNTRY_UNAVAILABLE"`) {
					t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
				}
				if len(billing.payerScopes) != 0 || billing.createCalls != 0 {
					t.Fatalf("o ms-billing não pode ser chamado: %d", len(billing.payerScopes))
				}
				if users != nil && users.patches != 0 {
					t.Fatal("nenhuma escrita no ms-user antes de saber o país")
				}
			})
		}
	}
}

func TestPaymentMethodNotAvailableKeepsCode(t *testing.T) {
	for _, tc := range payerCountryCases() {
		t.Run(tc.name, func(t *testing.T) {
			billing := &stubBillingClient{payerCallErr: &appdto.HTTPError{
				Code:    http.StatusUnprocessableEntity,
				Message: "Forma de pagamento indisponível para o país da conta.",
				Details: `{"timestamp":"2026-11-21T12:00:00Z","status":422,"error":"PAYMENT_METHOD_NOT_AVAILABLE","message":"Forma de pagamento indisponível para o país da conta."}`,
			}}
			users := brPayer()
			users.user.CountryCode = countryPtr("PT")
			rec := runPayerCountry(t, tc, billing, users, "")
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
			}
			body := rec.Body.String()
			if !strings.Contains(body, `"error":"PAYMENT_METHOD_NOT_AVAILABLE"`) || !strings.Contains(body, "Forma de pagamento indisponível para o país da conta.") {
				t.Fatalf("código e mensagem do ms-billing precisam seguir: %s", body)
			}
		})
	}
}

// Conta de fora sem CPF: o bff não pede nem grava documento; o pedido segue ao ms-billing, que recusa pelo país.
func TestPayerCountryAbroadSkipsDocument(t *testing.T) {
	tc := payerCountryCases()[0]
	tc.body = `{"planCode":"basic","interval":"month","paymentMethod":"pix","payerCpfCnpj":"52998224725"}`
	billing := &stubBillingClient{}
	users := brPayer()
	users.user.CountryCode = countryPtr("PT")
	users.user.PersonProfile = nil
	rec := runPayerCountry(t, tc, billing, users, "")
	if rec.Code >= 300 {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if users.patches != 0 {
		t.Fatalf("CPF não pode ser gravado para conta de fora: %d gravações", users.patches)
	}
	if len(billing.payerScopes) != 1 || billing.payerScopes[0].PayerCountry != "PT" {
		t.Fatalf("o pedido deve chegar ao ms-billing com PT: %+v", billing.payerScopes)
	}
}

// Conta do Brasil sem CPF continua recebendo PAYER_DOCUMENT_MISSING (regra atual), sem chamar o billing.
func TestPayerCountryBRStillNeedsDocument(t *testing.T) {
	tc := payerCountryCases()[0]
	billing := &stubBillingClient{}
	users := brPayer()
	users.user.PersonProfile = nil
	rec := runPayerCountry(t, tc, billing, users, "")
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "PAYER_DOCUMENT_MISSING") {
		t.Fatalf("esperado 422 PAYER_DOCUMENT_MISSING, veio %d %s", rec.Code, rec.Body.String())
	}
	if len(billing.payerScopes) != 0 {
		t.Fatalf("billing não pode ser chamado: %+v", billing.payerScopes)
	}
}
