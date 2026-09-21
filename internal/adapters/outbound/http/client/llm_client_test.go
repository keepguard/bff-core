package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpclient "github.com/keepguard/bff-core/internal/adapters/outbound/http/client"
	domainclient "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

type stubServiceTokens struct {
	token     string
	err       error
	companyID string
}

func (s *stubServiceTokens) GetToken(_ context.Context, companyID string) (string, error) {
	s.companyID = companyID
	return s.token, s.err
}

func TestLlmClientSendsServiceToken(t *testing.T) {
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

	tokens := &stubServiceTokens{token: "Bearer bff-service-jwt"}
	c := httpclient.NewLlmClient(&config.Config{
		Services: config.ServicesConfig{Llm: config.ServiceConfig{BaseURL: srv.URL}},
	}, tokens, zap.NewNop())

	// O token de quem está logado no backoffice não vai para o gateway.
	ctx := domainclient.WithBearerToken(context.Background(), "user-jwt")
	ctx = domainclient.WithCompanyID(ctx, "company-1")
	_, err := c.ListProviders(ctx, "tenant-1", "corr-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer bff-service-jwt" {
		t.Fatalf("Authorization=%q", gotAuth)
	}
	if tokens.companyID != "company-1" {
		t.Fatalf("companyID=%q", tokens.companyID)
	}
}

func TestLlmClientFailsWithoutCompanyInContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("gateway não deveria ser chamado sem company resolvida")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := httpclient.NewLlmClient(&config.Config{
		Services: config.ServicesConfig{Llm: config.ServiceConfig{BaseURL: srv.URL}},
	}, &stubServiceTokens{token: "Bearer x"}, zap.NewNop())

	if _, err := c.ListProviders(context.Background(), "tenant-1", "corr-1"); err == nil {
		t.Fatal("esperava erro sem company no contexto")
	}
}
