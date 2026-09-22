package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keepguard/bff-core/internal/application/connections"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetConnectionsHealthHandler_ReturnsBackendTTL(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	svc := connections.NewService(config.ConnectionsHealthConfig{
		SnapshotTTL:  60 * time.Second,
		ProbeTimeout: time.Second,
	}, connections.NewStore(nil), zap.NewNop()).WithTargets([]connections.Target{{
		ID:       "bff-auth",
		Name:     "BFF Auth",
		Group:    "gateway",
		Endpoint: "GET /health",
		URL:      up.URL,
	}})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/connections/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewConnectionsHandlers(svc, zap.NewNop())
	err := h.GetConnectionsHealthHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body connections.Snapshot
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, int64(60), body.TTLSeconds)
	assert.False(t, body.CheckedAt.IsZero())
	assert.False(t, body.ExpiresAt.IsZero())
	assert.False(t, body.Cached)
	assert.Equal(t, "healthy", body.Services[0].Status)
}
