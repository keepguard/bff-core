package http_test

// R1.1 do bff-core: rotas de billing que Achadinhos, InvestBot, o backoffice e o Echouno usam,
// em company sem configuração. O golden foi gerado com este mesmo arquivo na origin/main
// (9b33eba) e compara o pedido que chega ao ms-billing e a resposta ao cliente, byte a byte.
// Para refazer, só com a mudança justificada: UPDATE_GOLDEN=1 go test.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	httpserver "github.com/keepguard/bff-core/internal/adapters/in/http"
	"github.com/keepguard/bff-core/internal/adapters/in/http/handlers"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/in/http/middleware"
	httpclient "github.com/keepguard/bff-core/internal/adapters/out/http/client"
	billingdecorator "github.com/keepguard/bff-core/internal/adapters/out/http/decorator/billing"
	appbilling "github.com/keepguard/bff-core/internal/application/billing"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/keepguard/bff-core/internal/infrastructure/logger"
	"go.uber.org/zap"
)

const (
	testJWTSecret = "segredo-so-de-teste-com-mais-de-32-bytes"
	testCompany   = "4e65bc72-c60b-4382-977d-c15314d117a8"
	testTenant    = "aaaaaaaa-0000-4000-8000-000000000001"
	testUser      = "1a2b3c4d-0000-4000-8000-000000000001"
	testOther     = "1a2b3c4d-0000-4000-8000-000000000002"
	testSubID     = "bbbbbbbb-0000-4000-8000-000000000001"
	testCorrID    = "cccccccc-0000-4000-8000-000000000001"
)

type companyStub struct{ port.CompanyClient }

func (companyStub) GetByTenantId(context.Context, string, string) (appdto.MSCompanyResponseDTO, error) {
	return appdto.MSCompanyResponseDTO{ID: testCompany}, nil
}

// upstreamCall é o que chegou ao ms-billing falso.
type upstreamCall struct {
	Method  string              `json:"method"`
	URI     string              `json:"uri"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

// fakeBilling faz o papel do ms-billing: status e corpo por "MÉTODO caminho".
type fakeBilling struct {
	mu      sync.Mutex
	calls   []upstreamCall
	replies map[string]fakeReply
	server  *httptest.Server
}

type fakeReply struct {
	status int
	body   string
}

func newFakeBilling(t *testing.T) *fakeBilling {
	t.Helper()
	f := &fakeBilling{replies: map[string]fakeReply{}}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		headers := map[string][]string{}
		for key, values := range r.Header {
			headers[key] = values
		}
		f.mu.Lock()
		f.calls = append(f.calls, upstreamCall{Method: r.Method, URI: r.URL.RequestURI(), Headers: headers, Body: string(body)})
		f.mu.Unlock()
		reply, ok := f.replies[r.Method+" "+r.URL.Path]
		if !ok {
			reply = fakeReply{status: http.StatusOK, body: `{"ok":true}`}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(reply.status)
		if reply.status != http.StatusNoContent {
			_, _ = w.Write([]byte(reply.body))
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeBilling) drain() []upstreamCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := f.calls
	f.calls = nil
	return out
}

// signToken assina um JWT como o do ms-auth. Fora dos claims pedidos, valores fixos.
func signToken(t *testing.T, sub string, roles, authorities []string, extra map[string]any) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss": "ms-auth", "sub": sub, "tenant_id": testTenant, "roles": roles, "authorities": authorities,
		"exp": time.Now().Add(time.Hour).Unix(), "jti": "jti-fixo", "device_id": "dev-claim",
	}
	for key, value := range extra {
		claims[key] = value
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func baseConfig(billingURL string) *config.Config {
	cfg := &config.Config{}
	cfg.JWT.Secret = testJWTSecret
	cfg.Server.Port = "0"
	cfg.Server.RequestTimeout = 10 * time.Second
	cfg.Services.Billing.BaseURL = billingURL
	cfg.Services.Billing.Timeout = 5 * time.Second
	return cfg
}

// startBillingRouter sobe o servidor real, em porta livre, com os handlers e o cliente reais
// de billing, e devolve a URL base. Socket de verdade: o pedido passa por todos os middlewares.
func startBillingRouter(t *testing.T, cfg *config.Config, options ...func(httpserver.Server) httpserver.Server) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	cfg.Server.Port = fmt.Sprint(port)

	zapLogger := zap.NewNop()
	client := billingdecorator.New(httpclient.NewBillingClient(cfg, zapLogger), zapLogger, nil, "ms-billing")
	billingHandlers := handlers.NewBillingHandlers(appbilling.NewBillingPort(client), zapLogger)
	combined := handlers.NewCombinedHandlers(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, billingHandlers)
	rateLimiter := middlewarePkg.NewRateLimiterMiddleware(nil, cfg.RateLimit, zapLogger, nil)
	server := httpserver.NewServer(cfg, logger.NewNop(), nil, rateLimiter, nil, companyStub{}, nil).WithBillingClient(client)
	for _, option := range options {
		server = option(server)
	}
	server.SetupRoutes(combined)
	go func() { _ = server.Start() }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Stop(ctx)
	})
	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	for i := 0; i < 200; i++ {
		resp, err := http.Get(base + "/health")
		if err == nil {
			resp.Body.Close()
			return base
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("servidor não subiu")
	return ""
}

type goldenCase struct {
	name    string
	method  string
	path    string
	body    string
	token   string
	headers map[string]string
}

type goldenResult struct {
	Name         string         `json:"name"`
	Status       int            `json:"status"`
	ContentType  string         `json:"contentType"`
	ResponseBody string         `json:"responseBody"`
	Upstream     []upstreamCall `json:"upstream"`
}

var uuidInBody = regexp.MustCompile(`"timestamp":"[^"]*"`)

func normalizeUpstream(calls []upstreamCall) []upstreamCall {
	for i := range calls {
		if auth := calls[i].Headers["Authorization"]; len(auth) == 1 {
			calls[i].Headers["Authorization"] = []string{"Bearer <token>"}
		}
		// Ordem de JSON de map é fixa em Go; o corpo fica como veio.
		calls[i].Body = strings.TrimSpace(calls[i].Body)
	}
	return calls
}

func doRequest(t *testing.T, base string, tc goldenCase) (int, http.Header, string) {
	t.Helper()
	var reader io.Reader
	if tc.body != "" {
		reader = strings.NewReader(tc.body)
	}
	req, err := http.NewRequest(tc.method, base+tc.path, reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tc.token)
	req.Header.Set("X-Correlation-ID", testCorrID)
	req.Header.Set("User-Agent", "GoldenTest/1.0")
	for key, value := range tc.headers {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, resp.Header, string(body)
}

func runGolden(t *testing.T, base string, upstream *fakeBilling, tc goldenCase) goldenResult {
	t.Helper()
	status, headers, body := doRequest(t, base, tc)
	return goldenResult{
		Name: tc.name, Status: status, ContentType: headers.Get("Content-Type"),
		ResponseBody: uuidInBody.ReplaceAllString(body, `"timestamp":"<t>"`), Upstream: normalizeUpstream(upstream.drain()),
	}
}

func goldenCases(t *testing.T) []goldenCase {
	admin := signToken(t, testUser, []string{"ROLE_ADMIN"}, []string{"billing:read", "billing:write"}, nil)
	user := signToken(t, testUser, []string{"ROLE_USER"}, []string{"billing:read"}, nil)
	const base = "/api/v1/core/billing"
	sub := base + "/subscriptions/" + testSubID
	return []goldenCase{
		{name: "entitlement", method: http.MethodGet, path: base + "/entitlement", token: user},
		{name: "plans", method: http.MethodGet, path: base + "/plans", token: user},
		{name: "plan create", method: http.MethodPost, path: base + "/plans", token: admin, body: `{"code":"PRO","name":"Pro","level":1,"prices":[{"interval":"month","amountCents":9700,"currency":"BRL"}]}`},
		{name: "plan patch", method: http.MethodPatch, path: base + "/plans/PRO", token: admin, body: `{"name":"Pro 2"}`},
		{name: "plan delete", method: http.MethodDelete, path: base + "/plans/PRO", token: admin},
		{name: "gateway accounts", method: http.MethodGet, path: base + "/gateway-accounts", token: admin},
		{name: "subscription", method: http.MethodGet, path: base + "/subscription", token: user},
		{name: "subscribe pix, sem Idempotency-Key", method: http.MethodPost, path: base + "/subscriptions", token: user, body: `{"planCode":"PRO","interval":"month","paymentMethod":"pix"}`},
		{name: "subscribe cartao sem token", method: http.MethodPost, path: base + "/subscriptions", token: user, body: `{"planCode":"PRO","interval":"month","paymentMethod":"credit_card"}`},
		{name: "subscribe com X-Public-IP", method: http.MethodPost, path: base + "/subscriptions", token: user, body: `{"planCode":"PRO","interval":"month","paymentMethod":"pix"}`, headers: map[string]string{"X-Public-IP": "1.2.3.4", "X-Device-Id": "dev-header", "X-Session-Id": "sess-1"}},
		{name: "subscribe recusa PAN", method: http.MethodPost, path: base + "/subscriptions", token: user, body: `{"planCode":"PRO","cardNumber":"4111"}`},
		{name: "cancel", method: http.MethodPost, path: sub + "/cancel", token: user},
		{name: "cancel immediate de assinante", method: http.MethodPost, path: sub + "/cancel?immediate=true", token: user},
		{name: "cancel immediate de admin", method: http.MethodPost, path: sub + "/cancel?immediate=true", token: admin},
		{name: "reactivate sem corpo", method: http.MethodPost, path: sub + "/reactivate", token: user},
		{name: "change preview", method: http.MethodGet, path: sub + "/change-preview?planCode=PRO&interval=year", token: user},
		{name: "change plan", method: http.MethodPost, path: sub + "/change-plan", token: user, body: `{"planCode":"PRO","interval":"year"}`},
		{name: "scheduled change", method: http.MethodDelete, path: sub + "/scheduled-change", token: user},
		{name: "invoices", method: http.MethodGet, path: base + "/invoices?page=0&size=20&status=PAID", token: user},
		{name: "invoices de admin com q", method: http.MethodGet, path: base + "/invoices?q=" + testOther, token: admin},
		{name: "invoice", method: http.MethodGet, path: base + "/invoices/" + testSubID, token: user},
		{name: "invoice nao encontrada", method: http.MethodGet, path: base + "/invoices/naoexiste", token: user},
		{name: "entitlements de admin", method: http.MethodGet, path: base + "/entitlements?page=0", token: admin},
		{name: "entitlements de assinante recusado", method: http.MethodGet, path: base + "/entitlements", token: user},
		{name: "grant lifetime", method: http.MethodPost, path: base + "/subscriptions/grant-lifetime", token: admin, body: `{"targetUserId":"` + testOther + `","planCode":"VIP"}`},
	}
}

func TestBillingRegressionGolden(t *testing.T) {
	upstream := newFakeBilling(t)
	upstream.replies["GET /api/v1/billing/entitlement"] = fakeReply{http.StatusOK, `{"status":"active","allowsProduct":true,"planCode":"PRO"}`}
	upstream.replies["POST /api/v1/billing/plans"] = fakeReply{http.StatusCreated, `{"code":"PRO"}`}
	upstream.replies["DELETE /api/v1/billing/plans/PRO"] = fakeReply{http.StatusNoContent, ""}
	upstream.replies["POST /api/v1/billing/subscriptions"] = fakeReply{http.StatusCreated, `{"status":"pending_gateway"}`}
	upstream.replies["POST /api/v1/billing/subscriptions/grant-lifetime"] = fakeReply{http.StatusCreated, `{"status":"active"}`}
	upstream.replies["GET /api/v1/billing/invoices/naoexiste"] = fakeReply{http.StatusNotFound, `{"timestamp":"x","status":404,"error":"NOT_FOUND","message":"Fatura não encontrada"}`}
	router := startBillingRouter(t, baseConfig(upstream.server.URL))

	results := make([]goldenResult, 0)
	for _, tc := range goldenCases(t) {
		results = append(results, runGolden(t, router, upstream, tc))
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	got, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	const goldenPath = "testdata/billing_regression_golden.json"
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != string(got) {
		_ = os.WriteFile(os.TempDir()+"/billing_regression_got.json", got, 0o644)
		t.Fatalf("resposta ou pedido ao ms-billing mudou em company sem configuração; veja %s/billing_regression_got.json", os.TempDir())
	}
}
