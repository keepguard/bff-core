package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpclient "github.com/keepguard/bff-core/internal/adapters/outbound/http/client"
	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

func TestLlmClientForwardsInboundBearer(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/llm/providers" {
			t.Errorf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	c := httpclient.NewLlmClient(&config.Config{
		Services: config.ServicesConfig{Llm: config.ServiceConfig{BaseURL: srv.URL}},
	}, zap.NewNop())

	ctx := domainclient.WithBearerToken(context.Background(), "user-jwt")
	_, err := c.ListProviders(ctx, "tenant-1", "corr-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer user-jwt" {
		t.Fatalf("Authorization=%q", gotAuth)
	}
}
