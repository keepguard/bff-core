package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

	authClient := &mockAuthClient{}
	middleware := NewJWTMiddleware(authClient, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		token := GetTokenFromContext(c)
		if token != "token123" {
			t.Fatalf("expected token123, got %s", token)
		}
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJWTMiddleware_MissingCorrelationID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(nil, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestJWTMiddleware_MissingTenantId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-Correlation-ID", "corr-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(nil, zap.NewNop())
	handler := middleware.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware(nil, zap.NewNop())
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

	authClient := &mockAuthClient{err: echo.NewHTTPError(http.StatusUnauthorized, "invalid token")}
	middleware := NewJWTMiddleware(authClient, zap.NewNop())
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
