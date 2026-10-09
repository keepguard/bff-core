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

func TestRequireAuditRead_AllowsAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}, Authorities: []string{"audit:read"}})

	handler := RequireAuditRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireKnowledgeRead_AllowsAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}, Authorities: []string{"knowledge:read"}})

	handler := RequireKnowledgeRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireKnowledgeRead_RejectsWithoutAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}})

	handler := RequireKnowledgeRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireLlmRead_AllowsAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"MANAGER"}, Authorities: []string{"llm:read"}})

	handler := RequireLlmRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireLlmRead_RejectsWithoutAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"MANAGER"}})

	handler := RequireLlmRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireLlmWrite_AllowsAdmin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}})

	handler := RequireLlmWrite()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireLlmWrite_RejectsReadOnly(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"MANAGER"}, Authorities: []string{"llm:read"}})

	handler := RequireLlmWrite()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireCollectorRead_AllowsAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ROLE_MANAGER"}, Authorities: []string{"collector:read"}})

	handler := RequireCollectorRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireCollectorRead_RejectsManagerWithoutAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ROLE_MANAGER"}})

	handler := RequireCollectorRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireCollectorWrite_AllowsAdminWithoutAuthorities(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}})

	handler := RequireCollectorWrite()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireCollectorWrite_RejectsReadOnly(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ROLE_MANAGER"}, Authorities: []string{"collector:read"}})

	handler := RequireCollectorWrite()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireGuardianRead_AllowsUserWithAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}, Authorities: []string{"guardian:read"}})

	handler := RequireGuardianRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireOAuthWrite_RejectsOAuthRead(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"MANAGER"}, Authorities: []string{"oauth:read"}})

	handler := RequireOAuthWrite()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireOpsRead_AllowsSystem(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ROLE_SYSTEM"}})

	handler := RequireOpsRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuditRead_RejectsWithoutAuthority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}})

	handler := RequireAuditRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireBillingOrgRead_AllowsManagerWithBillingRead(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ROLE_MANAGER"}, Authorities: []string{"billing:read"}})

	handler := RequireBillingOrgRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireBillingOrgRead_RejectsUserWithBillingRead(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ROLE_USER"}, Authorities: []string{"billing:read"}})

	handler := RequireBillingOrgRead()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
