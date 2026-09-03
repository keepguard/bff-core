package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	httpclient "github.com/keepguard/bff-core/internal/adapters/outbound/http/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/keepguard/bff-core/internal/infrastructure/oauthsecret"
	"go.uber.org/zap"
)

func TestBffOAuthTokenClient_GetTokenCachesByCompany(t *testing.T) {
	const secretBase = "test-base"
	const plain = "bff-plain-secret"
	encrypted, err := oauthsecret.EncryptAESGCM(secretBase, plain)
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/oauth/runtime/secret":
			if r.Header.Get("X-Company-Id") != "company-1" {
				t.Errorf("unexpected company header %q", r.Header.Get("X-Company-Id"))
			}
			if r.URL.Query().Get("clientId") != "bff-core" {
				t.Errorf("unexpected clientId %q", r.URL.Query().Get("clientId"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"clientId":        "bff-core",
				"secretEncrypted": encrypted,
				"status":          "ACTIVE",
			})
		case "/api/v1/auth/oauth/token":
			tokenCalls.Add(1)
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode token body: %v", err)
			}
			if body["clientSecret"] != plain {
				t.Errorf("unexpected secret %q", body["clientSecret"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"accessToken": "svc-token",
				"tokenType":   "Bearer",
				"expiresIn":   3600,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := httpclient.NewBffOAuthTokenClient(&config.Config{
		Services: config.ServicesConfig{Auth: config.ServiceConfig{BaseURL: srv.URL, Timeout: 0}},
		OAuth:    config.OAuthConfig{ClientID: "bff-core", SecretBase: secretBase, TokenRenewBeforeSec: 600},
	}, zap.NewNop())

	token1, err := c.GetToken(context.Background(), "company-1")
	if err != nil {
		t.Fatal(err)
	}
	token2, err := c.GetToken(context.Background(), "company-1")
	if err != nil {
		t.Fatal(err)
	}
	if token1 != "Bearer svc-token" || token2 != token1 {
		t.Fatalf("unexpected tokens %q %q", token1, token2)
	}
	if tokenCalls.Load() != 1 {
		t.Fatalf("expected 1 token call, got %d", tokenCalls.Load())
	}
}

func TestBffOAuthTokenClient_RequiresCompany(t *testing.T) {
	c := httpclient.NewBffOAuthTokenClient(&config.Config{}, zap.NewNop())
	_, err := c.GetToken(context.Background(), "  ")
	if err == nil || !strings.Contains(err.Error(), "company_id") {
		t.Fatalf("expected company error, got %v", err)
	}
}
