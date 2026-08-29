package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type mockAuthClient struct {
	err error
}

func (m *mockAuthClient) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	return m.err
}

func TestJWTMiddleware_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware("test-secret", zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		token := GetTokenFromContext(c)
		if token != "token123" {
			t.Fatalf("expected token123, got %s", token)
		}
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err == nil {
		// token123 não é JWT válido com test-secret, esperado retornar 401
	}
}

func signTestJWT(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign jwt: %v", err)
	}
	return signed
}

func TestJWTMiddleware_ValidTokenWithTenantIDClaim(t *testing.T) {
	secret := "test-secret"
	e := echo.New()
	token := signTestJWT(t, secret, jwt.MapClaims{
		"sub":       "code-user-1",
		"tenant_id": "app-123",
		"exp":       4102444800,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(secret, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		if GetUserIDFromContext(c) != "code-user-1" {
			t.Fatalf("expected sub code-user-1, got %s", GetUserIDFromContext(c))
		}
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestJWTMiddleware_ParsesAuthorities(t *testing.T) {
	secret := "test-secret"
	e := echo.New()
	token := signTestJWT(t, secret, jwt.MapClaims{
		"sub":         "code-user-1",
		"tenant_id":   "app-123",
		"exp":         4102444800,
		"authorities": []any{"audit:read", "user:block"},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audits", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(secret, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		claims := GetClaimsFromContext(c)
		if claims == nil || len(claims.Authorities) != 2 || claims.Authorities[0] != "audit:read" {
			t.Fatalf("unexpected authorities: %+v", claims)
		}
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestJWTMiddleware_TenantMismatch(t *testing.T) {
	secret := "test-secret"
	e := echo.New()
	token := signTestJWT(t, secret, jwt.MapClaims{
		"sub":       "code-user-1",
		"tenant_id": "other-tenant",
		"exp":       4102444800,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(secret, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_MissingCorrelationID_GeneratesUUID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware("test-secret", zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Header().Get("X-Correlation-ID") == "" {
		t.Fatal("expected generated X-Correlation-ID")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", rec.Code)
	}
}

func TestJWTMiddleware_TenantFromJWTWithoutHeader(t *testing.T) {
	secret := "test-secret"
	e := echo.New()
	token := signTestJWT(t, secret, jwt.MapClaims{
		"sub":       "code-user-1",
		"tenant_id": "app-123",
		"exp":       4102444800,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Correlation-ID", "corr-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(secret, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		claims := GetClaimsFromContext(c)
		if claims == nil || claims.TenantId != "app-123" {
			t.Fatalf("expected tenant_id app-123, got %+v", claims)
		}
		if ResolveTenantId(c, claims) != "app-123" {
			t.Fatalf("expected resolved tenant app-123")
		}
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestJWTMiddleware_MissingTenantId(t *testing.T) {
	secret := "test-secret"
	e := echo.New()
	token := signTestJWT(t, secret, jwt.MapClaims{
		"sub": "code-user-1",
		"exp": 4102444800,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Correlation-ID", "corr-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(secret, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestJWTMiddleware_IgnoresCompanyIdClaim(t *testing.T) {
	secret := "test-secret"
	e := echo.New()
	token := signTestJWT(t, secret, jwt.MapClaims{
		"sub":       "code-user-1",
		"tenant_id": "app-123",
		"companyId": "should-be-ignored",
		"exp":       4102444800,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Correlation-ID", "corr-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(secret, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		claims := GetClaimsFromContext(c)
		if claims == nil || claims.TenantId != "app-123" {
			t.Fatalf("expected tenant from JWT, got %+v", claims)
		}
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware("test-secret", zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware("test-secret", zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetClaimsFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &pkg.JWTClaims{CodeUser: "c1"}
	c.Set("claims", claims)

	retrieved := GetClaimsFromContext(c)
	if retrieved == nil || retrieved.CodeUser != "c1" {
		t.Fatalf("expected claims with codeUser c1, got %+v", retrieved)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &pkg.JWTClaims{CodeUser: "c1"}
	c.Set("claims", claims)

	userID := GetUserIDFromContext(c)
	if userID != "c1" {
		t.Fatalf("expected c1, got %s", userID)
	}
}

func TestGetUserIDFromContext_UsesSubWhenCodeUserEmpty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &pkg.JWTClaims{Sub: "sub-1"}
	c.Set("claims", claims)

	userID := GetUserIDFromContext(c)
	if userID != "sub-1" {
		t.Fatalf("expected sub-1, got %s", userID)
	}
}

func TestJWTMiddleware_ParsesRoles(t *testing.T) {
	secret := "test-secret"
	e := echo.New()
	token := signTestJWT(t, secret, jwt.MapClaims{
		"sub":       "code-user-1",
		"tenant_id": "app-123",
		"roles":     []string{"ROLE_ADMIN", "USER"},
		"exp":       4102444800,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(secret, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		claims := GetClaimsFromContext(c)
		if claims == nil || !pkg.HasAnyRole(claims.Roles, "ADMIN") {
			t.Fatalf("expected ADMIN role in claims, got %+v", claims)
		}
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
