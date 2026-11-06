package http

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRateLimiterMiddleware_Disabled(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/register/init", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	logger, _ := zap.NewDevelopment()
	mw := NewRateLimiterMiddleware(nil, config.RateLimitConfig{Enabled: false}, logger, nil)

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
	mw := NewRateLimiterMiddleware(nil, config.RateLimitConfig{Enabled: true}, logger, nil)

	handler := mw.Limit("register_init", config.RateLimitRule{Limit: 5, Window: time.Minute})(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func limitOnce(t *testing.T, client *redis.Client, logged *bytes.Buffer, realIP string, limit int) int {
	t.Helper()
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(logged), zapcore.DebugLevel)
	mw := NewRateLimiterMiddleware(client, config.RateLimitConfig{Enabled: true}, zap.New(core), nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/register/init", nil)
	req.Header.Set("X-Real-IP", realIP)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := mw.Limit("register_init", config.RateLimitRule{Limit: limit, Window: time.Minute})(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c)
	require.NoError(t, err)
	return rec.Code
}

// §13 e §11: Redis fora, o aviso não leva o IP completo, nem dentro da chave.
func TestRateLimiterRedisErrorLogTruncatesIP(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 100 * time.Millisecond, MaxRetries: -1})
	defer client.Close()
	var logged bytes.Buffer
	assert.Equal(t, http.StatusOK, limitOnce(t, client, &logged, "203.0.113.77", 5), "segue em frente")
	assert.Contains(t, logged.String(), `"identifier":"203.0.113.0"`)
	assert.NotContains(t, logged.String(), "203.0.113.77")
}

// §13 e §11: limite excedido com aviso truncado; a chave no Redis segue com o IP completo,
// e dois IPs da mesma faixa continuam com limites separados. Roda contra redis-server
// temporário; sem ele na máquina, não roda.
func TestRateLimiterExceededLogTruncatesIPAndKeepsKey(t *testing.T) {
	path, err := exec.LookPath("redis-server")
	if err != nil {
		t.Skip("redis-server não instalado")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	server := exec.Command(path, "--port", fmt.Sprint(port), "--save", "", "--appendonly", "no", "--dir", t.TempDir())
	require.NoError(t, server.Start())
	t.Cleanup(func() { _ = server.Process.Kill(); _ = server.Wait() })
	client := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("127.0.0.1:%d", port)})
	defer client.Close()
	ctx := context.Background()
	require.Eventually(t, func() bool { return client.Ping(ctx).Err() == nil }, 2*time.Second, 20*time.Millisecond)

	var logged bytes.Buffer
	assert.Equal(t, http.StatusOK, limitOnce(t, client, &logged, "203.0.113.77", 1))
	assert.Equal(t, http.StatusTooManyRequests, limitOnce(t, client, &logged, "203.0.113.77", 1))
	assert.Contains(t, logged.String(), `"identifier":"203.0.113.0"`)
	assert.NotContains(t, logged.String(), "203.0.113.77")

	count, err := client.Get(ctx, "rl:bff-core:register_init:203.0.113.77").Int()
	require.NoError(t, err)
	assert.Equal(t, 2, count, "chave com o IP completo")
	assert.Equal(t, http.StatusOK, limitOnce(t, client, &logged, "203.0.113.78", 1), "mesma faixa, limite próprio")
}
