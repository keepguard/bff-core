package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	appbilling "github.com/keepguard/bff-core/internal/application/billing"
	"github.com/keepguard/bff-core/internal/pkg"
)

func cancelContext(body, query string, roles ...string) (echo.Context, interface{ Result() *http.Response }) {
	if len(roles) == 0 {
		roles = []string{"ROLE_USER"}
	}
	c, rec := billingContext(http.MethodPost, "/api/v1/core/billing/subscriptions/sub-123/cancel"+query, body, &pkg.JWTClaims{UserID: "user-1", Roles: roles})
	c.SetParamNames("id")
	c.SetParamValues("sub-123")
	c.Request().Header.Set("Idempotency-Key", "chave-cancelar")
	return c, rec
}

// Sem o objeto: o caminho de sempre (CancelSubscription, sem corpo), inclusive com outro JSON ou JSON quebrado.
func TestCancel_WithoutReasonUnchanged(t *testing.T) {
	for _, body := range []string{"", "{}", `{"cancellationReason":null}`, `{"foo":"bar"}`, "não é json"} {
		stub := &stubBillingClient{}
		h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
		c, rec := cancelContext(body, "?immediate=true")
		if err := h.CancelBillingSubscriptionHandler(c); err != nil {
			t.Fatal(err)
		}
		if rec.Result().StatusCode != http.StatusOK || len(stub.cancelCalls) != 1 || len(stub.forwarded) != 0 {
			t.Fatalf("%q: status %d, cancel %v, forward %v", body, rec.Result().StatusCode, stub.cancelCalls, stub.forwarded)
		}
		if stub.cancelCalls[0].id != "sub-123" || stub.cancelCalls[0].immediate {
			t.Fatalf("%q: titular nunca corta na hora: %+v", body, stub.cancelCalls[0])
		}
	}
}

// Com o objeto: só ele segue, cru, com a Idempotency-Key e a regra do immediate de sempre.
func TestCancel_WithReasonForwardsOnlyTheObject(t *testing.T) {
	stub := &stubBillingClient{}
	h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	c, rec := cancelContext(`{"cancellationReason":{"code":"nope","text":"x"},"remoteIp":"1.2.3.4","planCode":"VIP"}`, "?immediate=true")
	if err := h.CancelBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Result().StatusCode != http.StatusOK || len(stub.cancelCalls) != 0 || len(stub.forwarded) != 1 {
		t.Fatalf("status %d, cancel %v, forward %v", rec.Result().StatusCode, stub.cancelCalls, stub.forwarded)
	}
	got := stub.forwarded[0]
	if got.call.Method != http.MethodPost || got.call.Path != "/api/v1/billing/subscriptions/sub-123/cancel" || got.call.Operation != "CancelSubscription" {
		t.Fatalf("chamada %+v", got.call)
	}
	if got.call.Query != nil {
		t.Fatalf("titular: sem immediate: %v", got.call.Query)
	}
	if got.scope.IdempotencyKey != "chave-cancelar" || got.scope.OperatorID != "" {
		t.Fatalf("escopo %+v", got.scope)
	}
	raw, _ := json.Marshal(got.call.Body)
	if string(raw) != `{"cancellationReason":{"code":"nope","text":"x"}}` {
		t.Fatalf("corpo repassado: %s", raw)
	}

	// Operador (admin) com immediate=true: segue na query, como no caminho sem motivo.
	stub = &stubBillingClient{}
	h = NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	c, _ = cancelContext(`{"cancellationReason":{"code":"other"}}`, "?immediate=true", "ROLE_ADMIN")
	if err := h.CancelBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if len(stub.forwarded) != 1 || stub.forwarded[0].call.Query["immediate"] != "true" {
		t.Fatalf("admin: %+v", stub.forwarded)
	}
}

func TestCancel_BodyTooLarge(t *testing.T) {
	stub := &stubBillingClient{}
	h := NewBillingHandlers(appbilling.NewBillingPort(stub), zap.NewNop())
	c, rec := cancelContext(`{"cancellationReason":{"code":"other","text":"`+strings.Repeat("a", cancelBodyMaxBytes)+`"}}`, "")
	if err := h.CancelBillingSubscriptionHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Result().StatusCode != http.StatusRequestEntityTooLarge || len(stub.forwarded)+len(stub.cancelCalls) != 0 {
		t.Fatalf("status %d", rec.Result().StatusCode)
	}
}
