package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
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
	h := NewLlmHandlers(&stubLlmClient{}, zap.NewNop())
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
	h := NewLlmHandlers(&stubLlmClient{usage: appdto.PaginatedLlmUsageResponse{Content: []appdto.LlmUsageResponse{}, Size: 20}}, zap.NewNop())
	if err := h.ListLlmUsageHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateLlmProviderHandler_OK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/llm/providers", strings.NewReader(`{"name":"openai","providerType":"openai","apiKeyEnvRef":"OPENAI_KEEPGUARD_API_KEY"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	h := NewLlmHandlers(&stubLlmClient{}, zap.NewNop())
	if err := h.CreateLlmProviderHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}
