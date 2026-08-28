package connections

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

func TestTargets_OverridesURL(t *testing.T) {
	targets := Targets(config.ConnectionsHealthConfig{
		URLs: map[string]string{"bff-auth": "http://custom:8381/health"},
	})
	var found bool
	for _, target := range targets {
		if target.ID == "bff-auth" {
			found = true
			if target.URL != "http://custom:8381/health" {
				t.Fatalf("unexpected url %s", target.URL)
			}
		}
	}
	if !found {
		t.Fatal("bff-auth missing from catalog")
	}
}

func TestProbeAll_HealthyAndAuthAsUp(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "UP"})
	}))
	defer up.Close()
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer auth.Close()

	results := ProbeAll(context.Background(), []Target{
		{ID: "up", Name: "Up", Group: "gateway", Endpoint: "GET /health", URL: up.URL},
		{ID: "rabbit", Name: "Rabbit", Group: "infra", Endpoint: "GET /health", URL: auth.URL, TreatAuthAsUp: true},
	}, time.Second)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Status != "healthy" {
		t.Fatalf("expected healthy, got %s", results[0].Status)
	}
	if results[1].Status != "healthy" {
		t.Fatalf("expected auth-as-up healthy, got %s http=%d", results[1].Status, results[1].HTTPStatus)
	}
}

func TestService_SnapshotTTLComesFromConfig(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	svc := NewService(config.ConnectionsHealthConfig{
		SnapshotTTL:  60 * time.Second,
		ProbeTimeout: time.Second,
		URLs:         map[string]string{"bff-auth": up.URL, "bff-core": up.URL},
	}, NewStore(nil), zap.NewNop())
	svc.targets = []Target{{ID: "bff-auth", Name: "BFF Auth", Group: "gateway", Endpoint: "GET /health", URL: up.URL}}

	snap, err := svc.GetSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Cached {
		t.Fatal("first snapshot must not be cached")
	}
	if snap.TTLSeconds != 60 {
		t.Fatalf("expected ttlSeconds 60 from backend, got %d", snap.TTLSeconds)
	}
	if snap.CheckedAt.IsZero() || snap.ExpiresAt.Before(snap.CheckedAt) {
		t.Fatalf("invalid timestamps checkedAt=%s expiresAt=%s", snap.CheckedAt, snap.ExpiresAt)
	}
	if len(snap.Services) != 1 || snap.Services[0].Status != "healthy" {
		t.Fatalf("unexpected services %+v", snap.Services)
	}
}
