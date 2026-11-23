package http_test

import (
	"net/http"
	"strings"
	"testing"
)

// S2.3 §7: POST /core/billing/subscriptions/:id/card/setup repassa ao billing o corpo camelCase, a
// Idempotency-Key, o JWT, X-Company-Id e X-User-Id; recusa dado de cartão sem chamar o billing; e o 429 sai com o
// mesmo código e o Retry-After.
func TestCardSetupRouteForwards(t *testing.T) {
	upstream := newFakeBilling(t)
	path := "/api/v1/billing/subscriptions/" + testSubID + "/card/setup"
	upstream.replies["POST "+path] = fakeReply{http.StatusCreated, `{"setupId":"s-1","url":"https://checkout.stripe.com/c/pay/cs_1","expiresAt":"2026-11-23T15:00:00Z","reused":false}`}
	base := startBillingRouter(t, baseConfig(upstream.server.URL))
	subscriber := signToken(t, testUser, []string{"ROLE_USER"}, []string{"billing:read", "billing:self-write"}, map[string]any{"sid": "sid-1"})
	route := coreBase + "/subscriptions/" + testSubID + "/card/setup"

	status, _, body := doRequest(t, base, goldenCase{method: http.MethodPost, path: route, body: `{"cardDebitConsent":{"textVersion":"v1"},"locale":"pt-BR","ipAddress":"6.6.6.6","userAgent":"forjado"}`,
		token: subscriber, headers: map[string]string{"Idempotency-Key": "troca-1"}})
	if status != http.StatusCreated || !strings.Contains(body, `"setupId":"s-1"`) {
		t.Fatalf("status %d body %s", status, body)
	}
	calls := upstream.drain()
	if len(calls) != 1 || calls[0].Method != http.MethodPost || calls[0].URI != path {
		t.Fatalf("chamadas %+v", calls)
	}
	assertHeader(t, calls[0], "Idempotency-Key", "troca-1")
	assertHeader(t, calls[0], "X-User-Id", testUser)
	if got := calls[0].Headers["Authorization"]; len(got) != 1 || got[0] != "Bearer "+subscriber {
		t.Fatalf("JWT precisa ir intacto: %v", got)
	}
	if _, present := calls[0].Headers["X-Billing-Service-Key"]; present {
		t.Fatal("credencial não vai em rota de assinante")
	}
	if len(calls[0].Headers["X-Company-Id"]) != 1 {
		t.Fatal("X-Company-Id precisa ir ao billing")
	}
	for _, forged := range []string{"6.6.6.6", "forjado"} {
		if strings.Contains(calls[0].Body, forged) {
			t.Fatalf("IP e user agent do corpo não valem: %s", calls[0].Body)
		}
	}
	for _, want := range []string{`"cardDebitConsent":{"textVersion":"v1"}`, `"locale":"pt-BR"`, `"ipAddress":"127.0.0.1"`, `"userAgent":"GoldenTest/1.0"`} {
		if !strings.Contains(calls[0].Body, want) {
			t.Fatalf("corpo %s sem %s", calls[0].Body, want)
		}
	}

	// Sessão aberta reaproveitada: 200 como veio.
	upstream.replies["POST "+path] = fakeReply{http.StatusOK, `{"setupId":"s-1","reused":true}`}
	status, _, _ = doRequest(t, base, goldenCase{method: http.MethodPost, path: route, body: `{"cardDebitConsent":{"textVersion":"v1"}}`, token: subscriber, headers: map[string]string{"Idempotency-Key": "troca-2"}})
	if status != http.StatusOK {
		t.Fatalf("reaproveitada: %d", status)
	}
	upstream.drain()

	// Cartão no corpo: 422 sem chegar ao billing.
	for _, field := range []string{"cardNumber", "cvv", "pan"} {
		status, _, _ = doRequest(t, base, goldenCase{method: http.MethodPost, path: route, body: `{"` + field + `":"4242424242424242"}`, token: subscriber, headers: map[string]string{"Idempotency-Key": "k"}})
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("%s: status %d", field, status)
		}
	}
	if calls := upstream.drain(); len(calls) != 0 {
		t.Fatalf("dado de cartão não pode chegar ao billing: %d chamadas", len(calls))
	}

	// Erros do billing passam com o mesmo status e código; 429 com Retry-After.
	for _, tc := range []struct {
		status int
		code   string
	}{
		{http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND"}, {http.StatusConflict, "SUBSCRIPTION_NOT_ACTIVE"}, {http.StatusUnprocessableEntity, "CARD_SETUP_NOT_AVAILABLE"},
		{http.StatusBadGateway, "GATEWAY_UNAVAILABLE"}, {http.StatusTooManyRequests, "CARD_SETUP_RATE_LIMITED"},
	} {
		upstream.replies["POST "+path] = fakeReply{tc.status, `{"timestamp":"x","status":` + itoa(tc.status) + `,"error":"` + tc.code + `","message":"m","erro":{"codigo":"` + tc.code + `","mensagem":"m","detalhes":{"retryAfterSeconds":"600"}}}`}
		status, headers, body := doRequest(t, base, goldenCase{method: http.MethodPost, path: route, body: `{}`, token: subscriber, headers: map[string]string{"Idempotency-Key": "k-" + tc.code}})
		if status != tc.status || !strings.Contains(body, tc.code) {
			t.Fatalf("%s: status %d body %s", tc.code, status, body)
		}
		if tc.status == http.StatusTooManyRequests && headers.Get("Retry-After") != "600" {
			t.Fatalf("Retry-After: %q", headers.Get("Retry-After"))
		}
		if tc.status != http.StatusTooManyRequests && headers.Get("Retry-After") != "" {
			t.Fatalf("%s: Retry-After só no 429", tc.code)
		}
		if calls := upstream.drain(); len(calls) != 1 {
			t.Fatalf("%s: escrita não repete sozinha: %d chamadas", tc.code, len(calls))
		}
	}
}

func itoa(n int) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	out := ""
	for n > 0 {
		out = string(digits[n%10]) + out
		n /= 10
	}
	return out
}
