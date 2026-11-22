package handlers

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	appbilling "github.com/keepguard/bff-core/internal/application/billing"
	"github.com/keepguard/bff-core/internal/pkg"
	"go.uber.org/zap"
)

// Stripe S0a §6: payment-options leva o país do ms-user (nunca o do navegador); lock-state repassa
// a identidade de quem chama. Ambas são leitura (GET) pelo Forward.

func TestPaymentOptions_SendsCountryFromMsUser(t *testing.T) {
	billing := &stubBillingClient{}
	users := brPayer()
	users.user.CountryCode = countryPtr("us")
	c, rec := billingContext(http.MethodGet, "/api/v1/core/billing/payment-options", "", &pkg.JWTClaims{UserID: "user-1", CodeUser: "code-1", Roles: []string{"ROLE_USER"}})
	c.Request().Header.Set("X-Payer-Country", "BR")
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(users)

	if err := h.GetBillingPaymentOptionsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if len(billing.forwarded) != 1 {
		t.Fatalf("esperada 1 chamada ao ms-billing, veio %d", len(billing.forwarded))
	}
	got := billing.forwarded[0]
	if got.call.Method != http.MethodGet || got.call.Path != "/api/v1/billing/payment-options" {
		t.Fatalf("chamada errada: %+v", got.call)
	}
	if got.scope.PayerCountry != "US" {
		t.Fatalf("país do ms-user, não do navegador: %q", got.scope.PayerCountry)
	}
	if got.scope.UserID != "user-1" {
		t.Fatalf("X-User-Id do token: %q", got.scope.UserID)
	}
}

func TestPaymentOptions_NullCountryIsBR(t *testing.T) {
	billing := &stubBillingClient{}
	c, _ := billingContext(http.MethodGet, "/api/v1/core/billing/payment-options", "", &pkg.JWTClaims{UserID: "user-1", CodeUser: "code-1"})
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(brPayer())
	if err := h.GetBillingPaymentOptionsHandler(c); err != nil {
		t.Fatal(err)
	}
	if len(billing.forwarded) != 1 || billing.forwarded[0].scope.PayerCountry != "BR" {
		t.Fatalf("null vai como BR: %+v", billing.forwarded)
	}
}

func TestPaymentOptions_CountryUnavailableIs503WithoutBilling(t *testing.T) {
	billing := &stubBillingClient{}
	c, rec := billingContext(http.MethodGet, "/api/v1/core/billing/payment-options", "", &pkg.JWTClaims{UserID: "user-1", CodeUser: "code-1"})
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop()).WithUsers(&stubUserClient{getErr: errors.New("down")})
	if err := h.GetBillingPaymentOptionsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "PAYER_COUNTRY_UNAVAILABLE") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if len(billing.forwarded) != 0 {
		t.Fatal("o ms-billing não pode ser chamado sem o país")
	}
}

func TestLockState_ForwardsCallerIdentity(t *testing.T) {
	billing := &stubBillingClient{}
	c, rec := billingContext(http.MethodGet, "/api/v1/core/billing/lock-state", "", &pkg.JWTClaims{UserID: "user-1", CodeUser: "code-1"})
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop())
	if err := h.GetBillingLockStateHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || len(billing.forwarded) != 1 {
		t.Fatalf("status %d chamadas %d", rec.Code, len(billing.forwarded))
	}
	got := billing.forwarded[0]
	if got.call.Method != http.MethodGet || got.call.Path != "/api/v1/billing/lock-state" || got.scope.UserID != "user-1" {
		t.Fatalf("chamada errada: %+v %+v", got.call, got.scope)
	}
	if got.scope.OperatorID != "" {
		t.Fatal("leitura da própria pessoa não leva credencial de Operador")
	}
}

func TestCredits_ForwardsCurrency(t *testing.T) {
	billing := &stubBillingClient{}
	c, _ := billingContext(http.MethodGet, "/api/v1/core/billing/credits?currency=USD", "", &pkg.JWTClaims{UserID: "user-1", CodeUser: "code-1"})
	h := NewBillingHandlers(appbilling.NewBillingPort(billing), zap.NewNop())
	if err := h.GetBillingCreditsHandler(c); err != nil {
		t.Fatal(err)
	}
	if len(billing.forwarded) != 1 || billing.forwarded[0].call.Query["currency"] != "USD" {
		t.Fatalf("currency deve seguir ao ms-billing: %+v", billing.forwarded)
	}
}
