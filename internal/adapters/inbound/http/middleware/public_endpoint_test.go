package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestPublicEndpoint_MarkAsPublic(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewPublicEndpoint()
	handler := middleware.Middleware()(func(c echo.Context) error {
		if !IsPublicEndpoint(c) {
			t.Fatal("expected endpoint to be public")
		}
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsPublicEndpoint_NotPublic(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if IsPublicEndpoint(c) {
		t.Fatal("expected endpoint to not be public")
	}
}

func TestConditionalJWTMiddleware_PublicEndpoint(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("public_endpoint", true)

	mockJWT := &JWTMiddleware{}
	middleware := ConditionalJWTMiddleware(mockJWT)
	handler := middleware(func(c echo.Context) error {
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
