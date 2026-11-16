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

const compensationTestUserID = "11111111-1111-1111-1111-111111111111"

// upstreamReturning sobe um ms de teste que responde sempre com o status dado e guarda método/caminho/cabeçalho.
func upstreamReturning(t *testing.T, status int, seen *requestSeen) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			seen.method, seen.path, seen.company = r.Method, r.URL.Path, r.Header.Get("X-Company-Id")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status >= 400 {
			_, _ = w.Write([]byte(`{"error":"X","message":"boom"}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

type requestSeen struct {
	method, path, company string
}

// A compensação repetida (ou depois de um cadastro já desfeito) encontra 404: para a saga, é sucesso.
func TestHardDeleteUser_404EhSucesso(t *testing.T) {
	srv := upstreamReturning(t, http.StatusNotFound, nil)
	cfg := &config.Config{}
	cfg.Services.Auth.BaseURL = srv.URL
	c := &authClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}

	if err := c.HardDeleteUser(port.WithCompanyID(context.Background(), "company-1"), compensationTestUserID, "tenant", "corr"); err != nil {
		t.Fatalf("404 deve ser tolerado, veio erro: %v", err)
	}
}

// 500 continua erro: a tolerância é só para 404.
func TestHardDeleteUser_500ContinuaErro(t *testing.T) {
	srv := upstreamReturning(t, http.StatusInternalServerError, nil)
	cfg := &config.Config{}
	cfg.Services.Auth.BaseURL = srv.URL
	c := &authClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}

	if err := c.HardDeleteUser(port.WithCompanyID(context.Background(), "company-1"), compensationTestUserID, "tenant", "corr"); err == nil {
		t.Fatal("500 deve continuar sendo erro")
	}
}

func TestDeleteUser_404EhSucesso(t *testing.T) {
	srv := upstreamReturning(t, http.StatusNotFound, nil)
	cfg := &config.Config{}
	cfg.Services.User.BaseURL = srv.URL
	c := &userClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}

	if err := c.DeleteUser(port.WithCompanyID(context.Background(), "company-1"), compensationTestUserID, "tenant", "corr"); err != nil {
		t.Fatalf("404 deve ser tolerado, veio erro: %v", err)
	}
}

func TestDeleteUser_500ContinuaErro(t *testing.T) {
	srv := upstreamReturning(t, http.StatusBadGateway, nil)
	cfg := &config.Config{}
	cfg.Services.User.BaseURL = srv.URL
	c := &userClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}

	if err := c.DeleteUser(port.WithCompanyID(context.Background(), "company-1"), compensationTestUserID, "tenant", "corr"); err == nil {
		t.Fatal("502 deve continuar sendo erro")
	}
}

func TestDeleteAllByUserId_404EhSucesso(t *testing.T) {
	srv := upstreamReturning(t, http.StatusNotFound, nil)
	cfg := &config.Config{}
	cfg.Services.UserConsents.BaseURL = srv.URL
	c := &userConsentClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}

	if err := c.DeleteAllByUserId(port.WithCompanyID(context.Background(), "company-1"), compensationTestUserID, "tenant", "corr"); err != nil {
		t.Fatalf("404 deve ser tolerado, veio erro: %v", err)
	}
}

func TestDeleteAllByUserId_500ContinuaErro(t *testing.T) {
	srv := upstreamReturning(t, http.StatusInternalServerError, nil)
	cfg := &config.Config{}
	cfg.Services.UserConsents.BaseURL = srv.URL
	c := &userConsentClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}

	if err := c.DeleteAllByUserId(port.WithCompanyID(context.Background(), "company-1"), compensationTestUserID, "tenant", "corr"); err == nil {
		t.Fatal("500 deve continuar sendo erro")
	}
}

// ExistsByIDExternal: 200 = existe, 404 = não existe (sem erro), demais = erro. Manda o X-Company-Id do ctx.
func TestExistsByIDExternal(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		want    bool
		wantErr bool
	}{
		{name: "200 existe", status: http.StatusOK, want: true},
		{name: "404 não existe", status: http.StatusNotFound, want: false},
		{name: "500 é erro", status: http.StatusInternalServerError, wantErr: true},
		{name: "401 é erro", status: http.StatusUnauthorized, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var seen requestSeen
			srv := upstreamReturning(t, tc.status, &seen)
			cfg := &config.Config{}
			cfg.Services.Auth.BaseURL = srv.URL
			c := &authClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}

			got, err := c.ExistsByIDExternal(port.WithCompanyID(context.Background(), "company-1"), compensationTestUserID, "tenant", "corr")
			if tc.wantErr {
				if err == nil {
					t.Fatal("esperava erro")
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if got != tc.want {
				t.Fatalf("exists = %v, quer %v", got, tc.want)
			}
			if seen.method != http.MethodGet || seen.path != "/api/v1/users/id-user-external/"+compensationTestUserID {
				t.Fatalf("rota = %s %s", seen.method, seen.path)
			}
			if seen.company != "company-1" {
				t.Fatalf("X-Company-Id = %q", seen.company)
			}
		})
	}
}
