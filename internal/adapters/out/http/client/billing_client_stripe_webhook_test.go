package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// Stripe S1a §4: a assinatura (Stripe-Signature) é sobre o corpo cru. O bff-core repassa os bytes exatamente
// como chegaram (espaços, ordem das chaves, acentos) e o cabeçalho, sem reescrever o JSON.
func TestForwardStripeWebhookKeepsRawBodyAndSignature(t *testing.T) {
	raw := []byte("{\"id\":\"evt_1\",  \"type\":\"checkout.session.completed\",\n\"data\":{\"object\":{\"name\":\"José\",\"amount_total\":1500}} }")
	const signature = "t=1700000000,v1=5257a869e7ecebeda32affa62cdca3fa51cad7e77a0e56ff536d0ce8e108d8bd"
	var gotBody []byte
	var gotSignature, gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotSignature, gotPath = r.Header.Get("Stripe-Signature"), r.URL.Path
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer upstream.Close()
	cfg := &config.Config{}
	cfg.Services.Billing.BaseURL = upstream.URL
	client := NewBillingClient(cfg, zap.NewNop())

	if _, status, err := client.ForwardStripeWebhook(context.Background(), "", signature, raw); err != nil || status != http.StatusOK {
		t.Fatalf("status %d err %v", status, err)
	}
	if string(gotBody) != string(raw) {
		t.Fatalf("corpo alterado:\n%s\n%s", gotBody, raw)
	}
	if gotSignature != signature {
		t.Fatalf("Stripe-Signature não chegou: %q", gotSignature)
	}
	if gotPath != "/api/v1/billing/webhooks/stripe" {
		t.Fatalf("rota errada: %s", gotPath)
	}
}

// S1c §2: com companyRef, o repasse vai a /webhooks/stripe/{companyRef}, com o mesmo corpo cru.
func TestForwardStripeWebhookCompanyRefPath(t *testing.T) {
	var gotPath string
	var gotBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer upstream.Close()
	cfg := &config.Config{}
	cfg.Services.Billing.BaseURL = upstream.URL
	raw := []byte(`{"id":"evt_2" }`)
	if _, _, err := NewBillingClient(cfg, zap.NewNop()).ForwardStripeWebhook(context.Background(), "echouno", "t=1,v1=x", raw); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v1/billing/webhooks/stripe/echouno" || string(gotBody) != string(raw) {
		t.Fatalf("path %s body %s", gotPath, gotBody)
	}
}
