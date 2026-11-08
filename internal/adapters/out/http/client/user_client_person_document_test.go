package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
)

// O ms-user em AUTHZ_MODE=enforce nega chamada anônima em PATCH /users/:id/person-document: o cliente tem de
// mandar um token. Sem client de serviço, vai o token da própria pessoa (ela é a dona do registro).
func TestPatchPersonDocumentEnviaToken(t *testing.T) {
	var recebido string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recebido = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"u1"}`))
	}))
	defer upstream.Close()

	cfg := &config.Config{}
	cfg.Services.User.BaseURL = upstream.URL
	c := &userClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}
	ctx := port.WithCompanyID(context.Background(), "4e65bc72-c60b-4382-977d-c15314d117a8")

	if _, err := c.PatchPersonDocument(ctx, "u1", "52998224725", "token-da-pessoa", "tenant", "corr"); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if recebido != "Bearer token-da-pessoa" {
		t.Fatalf("Authorization = %q, esperava o token da pessoa", recebido)
	}
}

// Com token de serviço não vai X-Tenant-Id: o token carrega o tenant do bff-core e o ms-user recusaria o tenant do produto
// (403 "Inconsistência de isolamento de tenant", visto em 08/11/2026 assinando pelo Echouno). Com o token da pessoa, vai.
func TestPatchPersonDocumentTenantSoComTokenDaPessoa(t *testing.T) {
	var auth, tenant, company string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, tenant, company = r.Header.Get("Authorization"), r.Header.Get("X-Tenant-Id"), r.Header.Get("X-Company-Id")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"u1"}`))
	}))
	defer upstream.Close()

	cfg := &config.Config{}
	cfg.Services.User.BaseURL = upstream.URL
	ctx := port.WithCompanyID(context.Background(), "4e65bc72-c60b-4382-977d-c15314d117a8")
	servico := tokenFixo("token-de-servico")

	c := &userClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop(), oauth: servico}
	if _, err := c.PatchPersonDocument(ctx, "u1", "52998224725", "token-da-pessoa", "tenant-echouno", "corr"); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if auth != "Bearer token-de-servico" || tenant != "" || company != "4e65bc72-c60b-4382-977d-c15314d117a8" {
		t.Fatalf("serviço: Authorization=%q X-Tenant-Id=%q X-Company-Id=%q; esperava token de serviço, sem tenant, com company", auth, tenant, company)
	}

	c = &userClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}
	if _, err := c.PatchPersonDocument(ctx, "u1", "52998224725", "token-da-pessoa", "tenant-echouno", "corr"); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if auth != "Bearer token-da-pessoa" || tenant != "tenant-echouno" {
		t.Fatalf("pessoa: Authorization=%q X-Tenant-Id=%q; esperava o token e o tenant da pessoa", auth, tenant)
	}
}

type tokenFixo string

func (t tokenFixo) Token() (string, error) { return string(t), nil }
