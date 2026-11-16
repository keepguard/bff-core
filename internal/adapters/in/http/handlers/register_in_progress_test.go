package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keepguard/bff-core/internal/adapters/out/http/client"
	"github.com/labstack/echo/v4"
)

// msAuthInProgressBody mirrors the ms-auth ProblemDetail (properties nested) for 409 REGISTRATION_IN_PROGRESS.
const msAuthInProgressBody = `{"type":"https://keepguard.com/problems/conflict","title":"Conflito","status":409,` +
	`"detail":"internal detail that must not leak","instance":"/api/v1/users/create",` +
	`"properties":{"timestamp":"2026-11-16T10:00:00","path":"/api/v1/users/create","errorCode":"REGISTRATION_IN_PROGRESS","retryAfter":540}}`

func runHandleError(t *testing.T, err error) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodPost, "/register/confirm", nil), rec)
	if hErr := handleError(c, err, "corr-1"); hErr != nil {
		t.Fatalf("handleError returned %v", hErr)
	}
	var body map[string]any
	if jErr := json.Unmarshal(rec.Body.Bytes(), &body); jErr != nil {
		t.Fatalf("invalid json: %v", jErr)
	}
	return rec, body
}

func TestHandleError_RegistrationInProgressFromMsAuth(t *testing.T) {
	upstream := client.MapHTTPError(http.StatusConflict, []byte(msAuthInProgressBody), "auth service")
	// The saga may wrap the step error; the code must survive wrapping.
	for name, err := range map[string]error{"plain": upstream, "wrapped": fmt.Errorf("saga: %w", upstream)} {
		t.Run(name, func(t *testing.T) {
			rec, body := runHandleError(t, err)
			if rec.Code != http.StatusConflict {
				t.Fatalf("status = %d", rec.Code)
			}
			if body["error"] != "REGISTRATION_IN_PROGRESS" {
				t.Fatalf("error = %v", body["error"])
			}
			if body["retryAfter"] != float64(540) {
				t.Fatalf("retryAfter = %v", body["retryAfter"])
			}
			if rec.Header().Get("Retry-After") != "540" {
				t.Fatalf("Retry-After = %q", rec.Header().Get("Retry-After"))
			}
			msg, _ := body["message"].(string)
			if !strings.Contains(msg, "cadastro em andamento") || !strings.Contains(msg, "9 minutos") {
				t.Fatalf("message = %q", msg)
			}
			if strings.Contains(rec.Body.String(), "must not leak") || strings.Contains(rec.Body.String(), "/api/v1/users") {
				t.Fatalf("upstream detail leaked: %s", rec.Body.String())
			}
		})
	}
}

func TestHandleError_OtherConflictKeepsGenericPath(t *testing.T) {
	body := `{"properties":{"errorCode":"ALREADY_EXISTS"}}`
	rec, resp := runHandleError(t, client.MapHTTPError(http.StatusConflict, []byte(body), "auth service"))
	if rec.Code != http.StatusConflict || resp["error"] == "REGISTRATION_IN_PROGRESS" || resp["retryAfter"] != nil {
		t.Fatalf("unexpected: %d %v", rec.Code, resp)
	}
}

func TestRegistrationInProgressMessage(t *testing.T) {
	cases := map[int64]string{0: "alguns minutos", 30: "1 minuto.", 60: "1 minuto.", 61: "2 minutos", 900: "15 minutos"}
	for in, want := range cases {
		if got := registrationInProgressMessage(in); !strings.Contains(got, want) {
			t.Fatalf("%d -> %q (want %q)", in, got, want)
		}
	}
}
