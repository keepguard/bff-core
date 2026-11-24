package http_test

// L3 do Plano Cobrança: GET /core/billing/subscription repassa o corpo do ms-billing sem filtrar
// campos. O savedCard com a validade (expMonth/expYear, aditivos) chega ao cliente byte a byte, e
// o corpo sem validade (Asaas, Achadinhos, InvestBot) também.

import (
	"net/http"
	"testing"
)

func TestSubscriptionPassesSavedCardExpiryThrough(t *testing.T) {
	upstream := newFakeBilling(t)
	router := startBillingRouter(t, baseConfig(upstream.server.URL))
	user := signToken(t, testUser, []string{"ROLE_USER"}, []string{"billing:read"}, nil)

	for name, body := range map[string]string{
		"com validade": `{"id":"` + testSubID + `","status":"ACTIVE","currency":"USD","gateway":"STRIPE",` +
			`"savedCard":{"last4":"4242","brand":"VISA","expMonth":12,"expYear":2026}}`,
		"sem validade": `{"id":"` + testSubID + `","status":"ACTIVE","currency":"BRL","gateway":"ASAAS",` +
			`"savedCard":{"last4":"4242","brand":"VISA"}}`,
		"sem cartão": `{"id":"` + testSubID + `","status":"ACTIVE","currency":"BRL","gateway":"ASAAS","savedCard":null}`,
	} {
		upstream.replies["GET /api/v1/billing/subscription"] = fakeReply{http.StatusOK, body}
		status, headers, got := doRequest(t, router, goldenCase{method: http.MethodGet, path: "/api/v1/core/billing/subscription", token: user})
		calls := upstream.drain()
		if status != http.StatusOK || len(calls) != 1 {
			t.Fatalf("%s: status %d chamadas %d", name, status, len(calls))
		}
		if ct := headers.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("%s: content-type %q", name, ct)
		}
		if got != body {
			t.Fatalf("%s: corpo mudou no bff-core\nquer %s\nveio %s", name, body, got)
		}
	}
}
