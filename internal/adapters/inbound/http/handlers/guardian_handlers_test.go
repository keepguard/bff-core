package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type stubGuardianClient struct {
	page appdto.PaginatedGuardianIncidents
}

func (s *stubGuardianClient) ListIncidents(_ context.Context, _, _ string, _ map[string]string) (appdto.PaginatedGuardianIncidents, error) {
	return s.page, nil
}

func (s *stubGuardianClient) GetIncident(_ context.Context, _, _, _ string) (map[string]any, error) {
	return map[string]any{"id": "1"}, nil
}

func (s *stubGuardianClient) ExecuteAction(_ context.Context, _, _, _, _, _, _ string, _ appdto.GuardianExecuteActionRequest) (map[string]any, error) {
	return map[string]any{"outcome": "SUCCESS"}, nil
}

func (s *stubGuardianClient) ListRecipients(_ context.Context, _, _ string) ([]map[string]any, error) {
	return []map[string]any{}, nil
}

func (s *stubGuardianClient) UpsertRecipient(_ context.Context, _, _ string, _ appdto.GuardianRecipientUpsertRequest) (map[string]any, error) {
	return map[string]any{"email": "a@b.c"}, nil
}

func (s *stubGuardianClient) PatchRecipient(_ context.Context, _, _, _ string, _ appdto.GuardianRecipientUpsertRequest) (map[string]any, error) {
	return map[string]any{"enabled": false}, nil
}

func TestListGuardianIncidentsHandler_RequiresTenant(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/guardian/incidents", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}})
	h := NewGuardianHandlers(&stubGuardianClient{}, zap.NewNop())
	if err := h.ListGuardianIncidentsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListGuardianIncidentsHandler_OK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/guardian/incidents?page=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	h := NewGuardianHandlers(&stubGuardianClient{page: appdto.PaginatedGuardianIncidents{Size: 20}}, zap.NewNop())
	if err := h.ListGuardianIncidentsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListGuardianIncidentsHandler_Unavailable(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/guardian/incidents", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	h := NewGuardianHandlers(nil, zap.NewNop())
	if err := h.ListGuardianIncidentsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
