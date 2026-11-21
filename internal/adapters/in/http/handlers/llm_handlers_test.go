package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	appllm "github.com/keepguard/bff-core/internal/application/llm"
	domainclient "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type stubLlmClient struct {
	usage appdto.PaginatedLlmUsageResponse
}

func (s *stubLlmClient) ListProviders(context.Context, string, string) (json.RawMessage, error) {
	return json.RawMessage(`[]`), nil
}
func (s *stubLlmClient) CreateProvider(context.Context, string, string, any) (json.RawMessage, error) {
	return json.RawMessage(`{"id":"p1"}`), nil
}
func (s *stubLlmClient) UpdateProvider(context.Context, string, string, string, any) (json.RawMessage, error) {
	return json.RawMessage(`{"id":"p1"}`), nil
}
func (s *stubLlmClient) SetProviderEnabled(context.Context, string, string, string, bool) (json.RawMessage, error) {
	return json.RawMessage(`{"id":"p1","enabled":true}`), nil
}
func (s *stubLlmClient) SetProviderDefault(context.Context, string, string, string) (json.RawMessage, error) {
	return json.RawMessage(`{"id":"p1","isDefault":true}`), nil
}
func (s *stubLlmClient) Complete(context.Context, string, string, string, any) (json.RawMessage, error) {
	return json.RawMessage(`{"content":"ok"}`), nil
}
func (s *stubLlmClient) ListUsage(_ context.Context, _, _ string, _ map[string]string) (appdto.PaginatedLlmUsageResponse, error) {
	return s.usage, nil
}
func (s *stubLlmClient) GetUsage(context.Context, string, string, string) (appdto.LlmUsageResponse, error) {
	return appdto.LlmUsageResponse{ID: "u1"}, nil
}
func (s *stubLlmClient) ListAlertRules(context.Context, string, string) (json.RawMessage, error) {
	return json.RawMessage(`[]`), nil
}
func (s *stubLlmClient) CreateAlertRule(context.Context, string, string, any) (json.RawMessage, error) {
	return json.RawMessage(`{"id":"r1"}`), nil
}
func (s *stubLlmClient) UpdateAlertRule(context.Context, string, string, string, any) (json.RawMessage, error) {
	return json.RawMessage(`{"id":"r1"}`), nil
}
func (s *stubLlmClient) SetAlertRuleEnabled(context.Context, string, string, string, bool) (json.RawMessage, error) {
	return json.RawMessage(`{"id":"r1","enabled":true}`), nil
}
func (s *stubLlmClient) ListAlertFirings(context.Context, string, string, map[string]string) (json.RawMessage, error) {
	return json.RawMessage(`{"content":[]}`), nil
}

func TestListLlmUsageHandler_RequiresTenant(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/llm/usage", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}})
	h := NewLlmHandlers(appllm.NewLlmPort(&stubLlmClient{}), zap.NewNop())
	if err := h.ListLlmUsageHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListLlmUsageHandler_OK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/llm/usage?page=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	h := NewLlmHandlers(appllm.NewLlmPort(&stubLlmClient{usage: appdto.PaginatedLlmUsageResponse{Content: []appdto.LlmUsageResponse{}, Size: 20}}), zap.NewNop())
	if err := h.ListLlmUsageHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateLlmProviderHandler_OK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/llm/providers", strings.NewReader(`{"name":"openai","providerType":"openai","apiKey":"sk-live-secret"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	h := NewLlmHandlers(appllm.NewLlmPort(&stubLlmClient{}), zap.NewNop())
	if err := h.CreateLlmProviderHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

type capturingLlmClient struct {
	stubLlmClient
	gotBearer string
}

func (s *capturingLlmClient) CreateProvider(ctx context.Context, _, _ string, _ any) (json.RawMessage, error) {
	s.gotBearer = domainclient.BearerTokenFromContext(ctx)
	return json.RawMessage(`{"id":"p1"}`), nil
}

// O gateway LLM é chamado com a identidade de serviço do BFF, então o token de
// quem está logado não pode vazar no contexto do client de saída.
func TestCreateLlmProviderHandler_DoesNotForwardInboundBearer(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/llm/providers", strings.NewReader(`{"name":"openai"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.Set("token", "user-jwt")
	stub := &capturingLlmClient{}
	h := NewLlmHandlers(appllm.NewLlmPort(stub), zap.NewNop())
	if err := h.CreateLlmProviderHandler(c); err != nil {
		t.Fatal(err)
	}
	if stub.gotBearer != "" {
		t.Fatalf("bearer=%q", stub.gotBearer)
	}
}

// countingLlmClient conta as chamadas que chegariam ao gateway LLM.
type countingLlmClient struct {
	stubLlmClient
	calls int
}

func (s *countingLlmClient) ListUsage(ctx context.Context, a, b string, f map[string]string) (appdto.PaginatedLlmUsageResponse, error) {
	s.calls++
	return s.stubLlmClient.ListUsage(ctx, a, b, f)
}
func (s *countingLlmClient) GetUsage(ctx context.Context, a, b, c string) (appdto.LlmUsageResponse, error) {
	s.calls++
	return s.stubLlmClient.GetUsage(ctx, a, b, c)
}
func (s *countingLlmClient) Complete(ctx context.Context, a, b, c string, body any) (json.RawMessage, error) {
	s.calls++
	return s.stubLlmClient.Complete(ctx, a, b, c, body)
}
func (s *countingLlmClient) ListProviders(ctx context.Context, a, b string) (json.RawMessage, error) {
	s.calls++
	return s.stubLlmClient.ListProviders(ctx, a, b)
}

// Sem tenant, o guard interrompe de verdade: resposta 400 e nenhuma chamada ao gateway (antes c.JSON devolvia nil e o
// handler seguia adiante).
func TestLlmGuardStopsWithoutTenant(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		body   string
		call   func(h *LlmHandlers, c echo.Context) error
	}{
		{"usage", http.MethodGet, "/api/v1/core/llm/usage", "", func(h *LlmHandlers, c echo.Context) error { return h.ListLlmUsageHandler(c) }},
		{"usage-id", http.MethodGet, "/api/v1/core/llm/usage/u1", "", func(h *LlmHandlers, c echo.Context) error { return h.GetLlmUsageHandler(c) }},
		{"complete", http.MethodPost, "/api/v1/core/llm/complete", `{"prompt":"oi"}`, func(h *LlmHandlers, c echo.Context) error { return h.CompleteLlmHandler(c) }},
		{"providers", http.MethodGet, "/api/v1/core/llm/providers", "", func(h *LlmHandlers, c echo.Context) error { return h.ListLlmProvidersHandler(c) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			var body *strings.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(tc.method, tc.path, body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}})
			client := &countingLlmClient{}
			h := NewLlmHandlers(appllm.NewLlmPort(client), zap.NewNop())
			if err := tc.call(h, c); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("esperado 400, veio %d %s", rec.Code, rec.Body.String())
			}
			if client.calls != 0 {
				t.Fatalf("gateway LLM não pode ser chamado sem tenant: %d chamadas", client.calls)
			}
		})
	}
}
