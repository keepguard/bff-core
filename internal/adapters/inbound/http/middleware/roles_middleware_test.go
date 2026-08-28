package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
)

func TestRequireAnyRole_AllowsAdmin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ROLE_SYSTEM"}})

	handler := RequireAnyRole("ADMIN", "SYSTEM")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAnyRole_RejectsUser(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}})

	handler := RequireAnyRole("ADMIN", "SYSTEM")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
