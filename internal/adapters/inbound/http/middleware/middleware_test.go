package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func TestMiddleware_RequestID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewMiddleware(zap.NewNop())
	handler := middleware.RequestIDMiddleware()(func(c echo.Context) error {
		requestID := c.Response().Header().Get(echo.HeaderXRequestID)
		if requestID == "" {
			t.Fatal("expected request ID to be set")
		}
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMiddleware_Logging(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", "corr-123")
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/test")

	middleware := NewMiddleware(zap.NewNop())
	handler := middleware.LoggingMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_CORS(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewMiddleware(zap.NewNop())
	handler := middleware.CORSMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMiddleware_Security(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewMiddleware(zap.NewNop())
	handler := middleware.SecurityMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verificar headers de segurança
	headers := rec.Header()
	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("expected X-Content-Type-Options header")
	}
	if headers.Get("X-Frame-Options") != "DENY" {
		t.Fatal("expected X-Frame-Options header")
	}
}

func TestMiddleware_Metrics(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/test")

	middleware := NewMiddleware(zap.NewNop())
	handler := middleware.MetricsMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMiddleware_Timeout(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewMiddleware(zap.NewNop())
	handler := middleware.TimeoutMiddleware(1 * time.Second)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTraceID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "trace-123")

	traceID := GetTraceID(c)
	if traceID != "trace-123" {
		t.Fatalf("expected trace-123, got %s", traceID)
	}
}

func TestGetUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "user-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userID := GetUserID(c)
	if userID != "user-123" {
		t.Fatalf("expected user-123, got %s", userID)
	}
}

func TestSetUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetUserID(c, "user-456")

	userID := c.Request().Header.Get("X-User-ID")
	if userID != "user-456" {
		t.Fatalf("expected user-456, got %s", userID)
	}
}

func TestGetCorrelationID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	correlationID := GetCorrelationID(c)
	if correlationID != "corr-123" {
		t.Fatalf("expected corr-123, got %s", correlationID)
	}
}

func TestGetCorrelationID_Fallback(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set(echo.HeaderXRequestID, "req-123")

	correlationID := GetCorrelationID(c)
	if correlationID != "req-123" {
		t.Fatalf("expected req-123, got %s", correlationID)
	}
}

func TestSetCorrelationID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetCorrelationID(c, "corr-789")

	correlationID := c.Request().Header.Get("X-Correlation-ID")
	if correlationID != "corr-789" {
		t.Fatalf("expected corr-789, got %s", correlationID)
	}
}

func TestGetTenantId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-Id", "app-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tenantId := GetTenantId(c)
	if tenantId != "app-123" {
		t.Fatalf("expected app-123, got %s", tenantId)
	}
}

func TestSetTenantId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetTenantId(c, "app-456")

	tenantId := c.Request().Header.Get("X-Tenant-Id")
	if tenantId != "app-456" {
		t.Fatalf("expected app-456, got %s", tenantId)
	}
}
