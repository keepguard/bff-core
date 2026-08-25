package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRateLimiterMiddleware_Disabled(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/register/init", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	logger, _ := zap.NewDevelopment()
	mw := NewRateLimiterMiddleware(nil, config.RateLimitConfig{Enabled: false}, logger)

	handler := mw.Limit("register_init", config.RateLimitRule{Limit: 5, Window: time.Minute})(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimiterMiddleware_NilRedisFailOpen(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/register/init", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	logger, _ := zap.NewDevelopment()
	mw := NewRateLimiterMiddleware(nil, config.RateLimitConfig{Enabled: true}, logger)

	handler := mw.Limit("register_init", config.RateLimitRule{Limit: 5, Window: time.Minute})(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
