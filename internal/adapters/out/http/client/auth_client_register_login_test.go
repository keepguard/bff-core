package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	authDto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
)

// O aparelho em que a pessoa confirmou o cadastro vai ao ms-auth: é ele que dispensa o código de aparelho no login
// seguinte, no mesmo aparelho, quando o único canal de MFA da company é e-mail.
func TestRegisterLoginRepassaODispositivo(t *testing.T) {
	var recebido string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recebido = r.Header.Get("X-Device-Id")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"t","status":"AUTHENTICATED"}`))
	}))
	defer upstream.Close()

	cfg := &config.Config{}
	cfg.Services.Auth.BaseURL = upstream.URL
	c := &authClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}
	ctx := port.WithDeviceID(port.WithCompanyID(context.Background(), "4e65bc72-c60b-4382-977d-c15314d117a8"), "dev_abc")

	if _, err := c.RegisterLogin(ctx, authDto.AuthRegisterLoginRequestDTO{Username: "a@b.co", PasswordHash: "h", TenantId: "t"}, "t", "corr", "web"); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if recebido != "dev_abc" {
		t.Fatalf("X-Device-Id = %q, esperava dev_abc", recebido)
	}
}
