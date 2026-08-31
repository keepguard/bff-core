package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func TestCreateCollectorAgentHandler_Created(t *testing.T) {
	body := `{
		"name":"Coletor API",
		"collectorType":"API_REST",
		"collectorConfig":{"url":"http://example.com"},
		"schedule":{"daysOfWeek":[1],"startTime":"09:00","endTime":"17:00","intervalMinutes":30,"timezone":"UTC"},
		"enabled":false
	}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/agents", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.Set("token", "jwt-token")

	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, zap.NewNop())
	if err := h.CreateCollectorAgentHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var created appdto.CollectorAgentDetailDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Name != "Coletor API" || created.Enabled {
		t.Fatalf("unexpected create payload: %+v", created)
	}
}

func TestDeleteCollectorAgentHandler_NoContent(t *testing.T) {
	c, rec := oauthContext(http.MethodDelete, "/api/v1/core/collector/agents/a1", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("a1")
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, zap.NewNop())
	if err := h.DeleteCollectorAgentHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateCollectorAgentHandler_RequiresName(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/agents", strings.NewReader(`{"collectorType":"API_REST","collectorConfig":{},"schedule":{}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.Set("token", "jwt-token")

	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, zap.NewNop())
	if err := h.CreateCollectorAgentHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateCollectorAgentHandler_RejectsWithoutRole(t *testing.T) {
	e := echo.New()
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, zap.NewNop())
	handler := middlewarePkg.RequireAnyRole("ADMIN", "SYSTEM")(h.CreateCollectorAgentHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/agents", strings.NewReader(`{"name":"x","collectorType":"API_REST"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}, TenantId: "tenant-1"})
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}
