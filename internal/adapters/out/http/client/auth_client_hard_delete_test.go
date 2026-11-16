package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
)

// fakeServiceTokens devolve o token por company e guarda qual company pediu.
type fakeServiceTokens struct {
	token   string
	err     error
	company string
}

func (f *fakeServiceTokens) GetToken(_ context.Context, companyID string) (string, error) {
	f.company = companyID
	return f.token, f.err
}

// A compensação da saga de cadastro chama o hard delete do ms-auth com o token de serviço do bff-core da company do
// cadastro (o ms-auth em enforce exige claim company_id = X-Company-Id e user:service-hard-delete). Sem token, a
// chamada segue sem Authorization (observe aceita): a compensação nunca para por falta de token.
func TestHardDeleteUserEnviaTokenDeServico(t *testing.T) {
	const company = "fca9d466-e5f9-4af2-af03-23bf3ea678e5"
	cases := []struct {
		name     string
		tokens   port.ServiceTokenClient
		wantAuth string
		wantAsk  bool
	}{
		{name: "token já com Bearer (formato do bffOAuthTokenHTTP)", tokens: &fakeServiceTokens{token: "Bearer abc.def.ghi"},
			wantAuth: "Bearer abc.def.ghi", wantAsk: true},
		{name: "token cru ganha o prefixo", tokens: &fakeServiceTokens{token: "abc.def.ghi"},
			wantAuth: "Bearer abc.def.ghi", wantAsk: true},
		{name: "falha ao obter token: segue sem Authorization", tokens: &fakeServiceTokens{err: errors.New("ms-auth fora")},
			wantAuth: "", wantAsk: true},
		{name: "sem provedor: segue sem Authorization", tokens: nil, wantAuth: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotAuth, gotCompany, gotPath, gotMethod string
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth, gotCompany = r.Header.Get("Authorization"), r.Header.Get("X-Company-Id")
				gotPath, gotMethod = r.URL.Path, r.Method
				w.WriteHeader(http.StatusNoContent)
			}))
			defer upstream.Close()

			cfg := &config.Config{}
			cfg.Services.Auth.BaseURL = upstream.URL
			c := &authClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop(), tokens: tc.tokens}
			ctx := port.WithCompanyID(context.Background(), company)

			if err := c.HardDeleteUser(ctx, "11111111-1111-1111-1111-111111111111", "tenant", "corr"); err != nil {
				t.Fatalf("erro: %v", err)
			}
			if gotMethod != http.MethodDelete || gotPath != "/api/v1/users/hard-delete/11111111-1111-1111-1111-111111111111" {
				t.Fatalf("rota = %s %s", gotMethod, gotPath)
			}
			if gotCompany != company {
				t.Fatalf("X-Company-Id = %q", gotCompany)
			}
			if gotAuth != tc.wantAuth {
				t.Fatalf("Authorization = %q, esperava %q", gotAuth, tc.wantAuth)
			}
			if f, ok := tc.tokens.(*fakeServiceTokens); ok && tc.wantAsk && f.company != company {
				t.Fatalf("token pedido para a company %q, esperava a do cadastro %q", f.company, company)
			}
		})
	}
}

// Recusa do ms-auth (enforce) volta como erro para a saga, como qualquer outro status diferente de 204.
func TestHardDeleteUserPropagaRecusa(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()
	cfg := &config.Config{}
	cfg.Services.Auth.BaseURL = upstream.URL
	c := &authClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}
	if err := c.HardDeleteUser(port.WithCompanyID(context.Background(), "c"), "id", "t", "corr"); err == nil {
		t.Fatal("401 deveria virar erro")
	}
}
