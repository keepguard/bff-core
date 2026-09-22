package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keepguard/bff-core/internal/application/audit"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type stubAuditPort struct {
	page appdto.PaginatedAuditResponse
}

func (s *stubAuditPort) List(_ context.Context, _ audit.ListAuditsQuery) (appdto.PaginatedAuditResponse, error) {
	return s.page, nil
}

func (s *stubAuditPort) Get(_ context.Context, _ audit.GetAuditQuery) (appdto.AuditDetailResponse, error) {
	return appdto.AuditDetailResponse{}, nil
}

func TestListAuditsHandler_RequiresTenant(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audits", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}})
	h := NewAuditHandlers(&stubAuditPort{}, zap.NewNop())
	if err := h.ListAuditsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListAuditsHandler_OK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/audits?page=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	h := NewAuditHandlers(&stubAuditPort{page: appdto.PaginatedAuditResponse{Content: []appdto.AuditEventResponse{}, Size: 20}}, zap.NewNop())
	if err := h.ListAuditsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
