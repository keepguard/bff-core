package client

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	billingdecorator "github.com/keepguard/bff-core/internal/adapters/out/http/decorator/billing"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Nada de chave, JWT ou token de cartão em log, no caminho feliz e no erro (contrato C2 §3).
func TestBillingForwardNeverLogsSecrets(t *testing.T) {
	const (
		serviceKey = "chave-secreta-do-bff-core"
		bearer     = "jwt.secreto.dotoken"
		cardToken  = "token-do-cartao-secreto"
	)
	status := http.StatusOK
	var gotKey, gotOperator string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey, gotOperator = r.Header.Get("X-Billing-Service-Key"), r.Header.Get("X-Operator-Id")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"error":"X","message":"recusado"}`))
	}))
	defer upstream.Close()

	var buffer bytes.Buffer
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&buffer), zapcore.DebugLevel)
	logger := zap.New(core)
	cfg := &config.Config{}
	cfg.Services.Billing.BaseURL = upstream.URL
	cfg.BillingCore.ServiceKey = serviceKey
	billing := billingdecorator.New(NewBillingClient(cfg, logger), logger, nil, "ms-billing")

	ctx := port.WithBearerToken(context.Background(), bearer)
	scope := port.BillingScope{CompanyID: "c", UserID: "u", CorrelationID: "corr", OperatorID: "op", IdempotencyKey: "idem"}
	call := port.BillingCall{Operation: "GrantCredit", Method: http.MethodPost, Path: "/api/v1/billing/credits", Endpoint: "/billing/credits", Body: map[string]any{"creditCardToken": cardToken}}
	if _, err := billing.Forward(ctx, scope, call); err != nil {
		t.Fatal(err)
	}
	if gotKey != serviceKey || gotOperator != "op" {
		t.Fatalf("credencial precisa chegar: %q %q", gotKey, gotOperator)
	}
	status = http.StatusInternalServerError
	if _, err := billing.Forward(ctx, scope, call); err == nil {
		t.Fatal("esperado erro")
	}
	for _, secret := range []string{serviceKey, bearer, cardToken} {
		if strings.Contains(buffer.String(), secret) {
			t.Fatalf("segredo em log: %s", buffer.String())
		}
	}
	if !strings.Contains(buffer.String(), "GrantCredit") {
		t.Fatalf("log esperado da operação: %s", buffer.String())
	}
}

// Sem operador, a credencial não sai, mesmo configurada: rotas de hoje seguem iguais.
func TestBillingScopeWithoutOperatorSendsNoCredential(t *testing.T) {
	var header http.Header
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Clone()
		_, _ = w.Write([]byte(`{}`))
	}))
	defer upstream.Close()
	cfg := &config.Config{}
	cfg.Services.Billing.BaseURL = upstream.URL
	cfg.BillingCore.ServiceKey = "chave"
	billing := NewBillingClient(cfg, zap.NewNop())
	if _, err := billing.GetEntitlement(context.Background(), port.BillingScope{CompanyID: "c", UserID: "u"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"X-Billing-Service-Key", "X-Operator-Id", "X-Real-Ip", "Idempotency-Key"} {
		if header.Get(name) != "" {
			t.Fatalf("%s não vai em rota de hoje", name)
		}
	}
}

func TestUserAccessClient(t *testing.T) {
	var path, auth, company string
	status, body := http.StatusOK, `{"codeUser":"u","companyId":"c","active":true,"roles":["ROLE_ADMIN"],"authorities":["billing:admin"]}`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, auth, company = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("X-Company-Id")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	defer upstream.Close()
	cfg := &config.Config{}
	cfg.Services.Auth.BaseURL = upstream.URL
	access := NewUserAccessClient(cfg)

	got, err := access.GetUserAccess(context.Background(), "abc", "c", "tok", "corr")
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/users/code-user/abc/access" || auth != "Bearer tok" || company != "c" {
		t.Fatalf("pedido %s %s %s", path, auth, company)
	}
	if !got.Active || got.Roles[0] != "ROLE_ADMIN" || got.Authorities[0] != "billing:admin" {
		t.Fatalf("resposta %+v", got)
	}
	for _, refused := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound} {
		status = refused
		if _, err := access.GetUserAccess(context.Background(), "abc", "c", "tok", "corr"); err != port.ErrUserAccessRefused {
			t.Fatalf("%d: %v", refused, err)
		}
	}
	status = http.StatusInternalServerError
	if _, err := access.GetUserAccess(context.Background(), "abc", "c", "tok", "corr"); err == nil || err == port.ErrUserAccessRefused {
		t.Fatalf("500 é indisponível: %v", err)
	}
	upstream.Close()
	if _, err := access.GetUserAccess(context.Background(), "abc", "c", "tok", "corr"); err == nil || err == port.ErrUserAccessRefused {
		t.Fatalf("rede fora é indisponível: %v", err)
	}
}
