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

	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
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

func TestListCollectorAgentsHandler_MapsLastExecution(t *testing.T) {
	stub := &oauthStubCollector{
		agents: []appdto.CollectorAgentRaw{{
			ID:            "a1",
			Name:          "Money Times RSS",
			CollectorType: "API_REST",
			Enabled:       true,
			LastExecution: &appdto.CollectorLastExecutionRaw{
				ID:        "exec-1",
				StartedAt: "2026-09-04T02:10:00Z",
				Status:    "SUCCESS",
			},
		}},
	}
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/collector/agents", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	h := NewCollectorAgentHandlers(stub, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.ListCollectorAgentsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out appdto.PaginatedCollectorAgents
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Content) != 1 || out.Content[0].LastExecution == nil || out.Content[0].LastExecution.Status != "SUCCESS" {
		t.Fatalf("unexpected last execution: %+v", out.Content)
	}
}

func TestDeleteCollectorAgentHandler_NoContent(t *testing.T) {
	c, rec := oauthContext(http.MethodDelete, "/api/v1/core/collector/agents/a1", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("a1")
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
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

	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.CreateCollectorAgentHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateCollectorAgentHandler_RejectsWithoutRole(t *testing.T) {
	e := echo.New()
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
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

func TestListCollectorAgentExecutionsHandler_OK(t *testing.T) {
	finished := "2026-08-31T19:47:00Z"
	stub := &oauthStubCollector{
		executions: []appdto.CollectorExecutionRaw{
			{
				ID:             "exec-1",
				AgentID:        "a1",
				StartedAt:      "2026-08-31T19:46:00Z",
				FinishedAt:     &finished,
				Status:         "SUCCESS",
				ItemsCollected: 1,
				ItemsUploaded:  1,
			},
		},
	}
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/collector/agents/a1/executions", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("a1")
	h := NewCollectorAgentHandlers(stub, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.ListCollectorAgentExecutionsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out []appdto.CollectorExecutionDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].ID != "exec-1" || out[0].AgentID != "a1" || out[0].ItemsUploaded != 1 || out[0].Status != "SUCCESS" {
		t.Fatalf("unexpected executions: %+v", out)
	}
}

func TestListCollectorAgentExecutionsHandler_NotFound(t *testing.T) {
	stub := &oauthStubCollector{
		execErr: &appdto.HTTPError{Code: http.StatusNotFound, Message: "Recurso não encontrado"},
	}
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/collector/agents/missing/executions", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("missing")
	h := NewCollectorAgentHandlers(stub, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.ListCollectorAgentExecutionsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListCollectorAgentExecutionsHandler_RejectsWithoutRole(t *testing.T) {
	e := echo.New()
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	handler := middlewarePkg.RequireAnyRole("ADMIN", "SYSTEM")(h.ListCollectorAgentExecutionsHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/collector/agents/a1/executions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("a1")
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}, TenantId: "tenant-1"})
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListCollectorDataSourcesHandler_OK(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/collector/data-sources", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	stub := &oauthStubCollector{
		sources: []appdto.CollectorDataSourceRaw{
			{
				ID:              "src-1",
				Slug:            "status-invest",
				Name:            "Status Invest",
				CollectorType:   "API_REST",
				DefaultContext:  "investimentos",
				DefaultSchedule: []byte(`{"days_of_week":[1],"start_time":"09:00","end_time":"17:00","interval_minutes":60,"timezone":"UTC"}`),
				ConfigTemplate:  []byte(`{"url":"https://statusinvest.com.br"}`),
				Variables:       []byte(`[{"key":"ticker","label":"Ticker","required":true}]`),
			},
		},
	}
	h := NewCollectorAgentHandlers(stub, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.ListCollectorDataSourcesHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out []appdto.CollectorDataSourceDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Slug != "status-invest" || out[0].CollectorType != "API_REST" || out[0].Scope != "company" {
		t.Fatalf("unexpected sources: %+v", out)
	}
}

func TestCreateCollectorDataSourceHandler_Created(t *testing.T) {
	body := `{
		"name":"CVM fatos",
		"slug":"cvm-fatos",
		"collectorType":"HTML_SCRAPER",
		"defaultSchedule":{"daysOfWeek":[1],"startTime":"09:00","endTime":"17:00","intervalMinutes":60,"timezone":"UTC"},
		"configTemplate":{"url":"https://exemplo.com/{{ticker_lower}}"},
		"variables":[{"key":"ticker","label":"Ticker","required":true}]
	}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/data-sources", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.Set("token", "jwt-token")

	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.CreateCollectorDataSourceHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var created appdto.CollectorDataSourceDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Slug != "cvm-fatos" || created.Scope != "company" || !created.Enabled {
		t.Fatalf("unexpected create payload: %+v", created)
	}
}

func TestDeleteCollectorDataSourceHandler_NoContent(t *testing.T) {
	c, rec := oauthContext(http.MethodDelete, "/api/v1/core/collector/data-sources/src-1", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("src-1")
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.DeleteCollectorDataSourceHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPropagateCollectorDataSourceHandler_OK(t *testing.T) {
	body := `{"fields":["url"],"dryRun":true}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/data-sources/src-1/propagate", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.SetParamNames("id")
	c.SetParamValues("src-1")

	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.PropagateCollectorDataSourceHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out appdto.PropagateDataSourceDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !out.DryRun || out.TotalLinked != 2 || out.Updated != 2 || len(out.Previews) != 1 || out.Previews[0].AfterURL != "https://new.com/PETR4" {
		t.Fatalf("unexpected propagate payload: %+v", out)
	}
}

func TestGetCollectorExecutionPayloadsHandler_UsesPayloadRefs(t *testing.T) {
	stub := &oauthStubCollector{
		executions: []appdto.CollectorExecutionRaw{
			{
				ID:             "exec-1",
				AgentID:        "a1",
				StartedAt:      "2026-09-01T03:18:00Z",
				Status:         "SUCCESS",
				ItemsCollected: 1,
				ItemsUploaded:  1,
				Metadata: map[string]any{
					"payload_refs": []any{
						map[string]any{"kind": "snapshot", "id": "snap-1"},
					},
				},
			},
		},
	}
	knowledge := &stubKnowledgeClient{
		snapshot: appdto.KnowledgeSnapshotDTO{
			ID:            "snap-1",
			CollectorType: "API_REST",
			EntityHint:    "PETR4",
			Payload:       map[string]any{"ticker": "PETR4", "price": 32.1},
		},
	}
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/collector/executions/exec-1/payloads", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("executionId")
	c.SetParamValues("exec-1")
	h := NewCollectorAgentHandlers(stub, &oauthStubCompany{id: "company-1"}, knowledge, &stubServiceToken{}, zap.NewNop())
	if err := h.GetCollectorExecutionPayloadsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out []appdto.ExecutionPayloadItemDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Kind != "snapshot" || out[0].ID != "snap-1" {
		t.Fatalf("unexpected payloads: %+v", out)
	}
	if out[0].Payload["ticker"] != "PETR4" {
		t.Fatalf("expected ticker in payload, got %+v", out[0].Payload)
	}
}

func TestGetCollectorExecutionPayloadsHandler_FallbackCollectionResults(t *testing.T) {
	stub := &oauthStubCollector{
		executions: []appdto.CollectorExecutionRaw{
			{
				ID:             "exec-2",
				AgentID:        "a1",
				StartedAt:      "2026-09-01T03:18:00Z",
				Status:         "SUCCESS",
				ItemsCollected: 1,
				ItemsUploaded:  1,
			},
		},
	}
	knowledge := &stubKnowledgeClient{
		results: appdto.KnowledgeCollectionResultsDTO{
			Snapshots: []appdto.KnowledgeSnapshotDTO{{
				ID:      "snap-fallback",
				Payload: map[string]any{"ok": true},
			}},
		},
	}
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/collector/executions/exec-2/payloads", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("executionId")
	c.SetParamValues("exec-2")
	h := NewCollectorAgentHandlers(stub, &oauthStubCompany{id: "company-1"}, knowledge, &stubServiceToken{}, zap.NewNop())
	if err := h.GetCollectorExecutionPayloadsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out []appdto.ExecutionPayloadItemDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].ID != "snap-fallback" {
		t.Fatalf("unexpected fallback payloads: %+v", out)
	}
}

func TestBulkCollectorAgentsHandler_RunAccepted(t *testing.T) {
	body := `{"action":"run","ids":["a1","a2"]}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/agents/bulk", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.BulkCollectorAgentsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out appdto.CollectorBulkResultDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Action != "run" || out.Requested != 2 || out.BulkID == "" {
		t.Fatalf("unexpected bulk payload: %+v", out)
	}
}

func TestGetCollectorBulkOperationHandler_OK(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/collector/agents/bulk-operations/bulk-1", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("bulk-1")
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.GetCollectorBulkOperationHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetCollectorActiveBulkOperationHandler_OK(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/collector/agents/bulk-operations/active", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	h := NewCollectorAgentHandlers(&oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.GetCollectorActiveBulkOperationHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out appdto.CollectorBulkProgressDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != "bulk-active" || out.Status != "running" {
		t.Fatalf("unexpected active bulk: %+v", out)
	}
}

func TestListCollectorIncidentsHandler_MapsCamelCase(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/collector/incidents?status=open", nil)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	h := NewCollectorAgentHandlers(&oauthStubCollector{
		incidents: []appdto.CollectorIncidentRaw{
			{ID: "inc-1", AgentID: "a1", AgentName: "ARZZ3", Classification: "source_changed", Status: "open", Occurrences: 2},
		},
	}, &oauthStubCompany{id: "company-1"}, nil, nil, zap.NewNop())
	if err := h.ListCollectorIncidentsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out appdto.PaginatedCollectorIncidents
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Content) != 1 || out.Content[0].AgentName != "ARZZ3" || out.Content[0].Classification != "source_changed" {
		t.Fatalf("unexpected incidents: %+v", out.Content)
	}
}
