package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type stubKnowledgeClient struct {
	last       appdto.KnowledgeAskRequest
	lastBearer string
	sources    []appdto.KnowledgeAskSource
	snapshot   appdto.KnowledgeSnapshotDTO
	document   appdto.KnowledgeDocumentPreviewDTO
	results    appdto.KnowledgeCollectionResultsDTO
}

type stubServiceToken struct {
	token       string
	err         error
	lastCompany string
}

func (s *stubServiceToken) GetToken(_ context.Context, companyID string) (string, error) {
	s.lastCompany = companyID
	if s.err != nil {
		return "", s.err
	}
	if s.token != "" {
		return s.token, nil
	}
	return "Bearer oauth-service", nil
}

func (s *stubKnowledgeClient) Ask(_ context.Context, _, bearerToken, _ string, body appdto.KnowledgeAskRequest) (appdto.KnowledgeAskResponse, error) {
	s.last = body
	s.lastBearer = bearerToken
	sources := s.sources
	if sources == nil {
		sources = []appdto.KnowledgeAskSource{}
	}
	return appdto.KnowledgeAskResponse{
		Intent:      "HEALTH",
		Mode:        "HEURISTIC",
		Answer:      "MS Auth healthy, 16:46, 132 ms",
		Sources:     sources,
		Convergence: true,
	}, nil
}

func (s *stubKnowledgeClient) GetSnapshot(_ context.Context, _, _, _, snapshotID string) (appdto.KnowledgeSnapshotDTO, error) {
	if s.snapshot.ID != "" {
		return s.snapshot, nil
	}
	return appdto.KnowledgeSnapshotDTO{
		ID:      snapshotID,
		Payload: map[string]any{"ticker": "PETR4"},
	}, nil
}

func (s *stubKnowledgeClient) GetDocumentPreview(_ context.Context, _, _, _, documentID string) (appdto.KnowledgeDocumentPreviewDTO, error) {
	if s.document.ID != "" {
		return s.document, nil
	}
	return appdto.KnowledgeDocumentPreviewDTO{
		ID:               documentID,
		FileName:         "page.html",
		ContentType:      "text/html",
		PreviewText:      "<html>ok</html>",
		PreviewAvailable: true,
	}, nil
}

func (s *stubKnowledgeClient) GetCollectionResults(_ context.Context, _, _, _, _, _ string, _ int) (appdto.KnowledgeCollectionResultsDTO, error) {
	if s.results.Snapshots != nil || s.results.Documents != nil {
		return s.results, nil
	}
	return appdto.KnowledgeCollectionResultsDTO{
		Snapshots: []appdto.KnowledgeSnapshotDTO{},
		Documents: []appdto.KnowledgeDocumentPreviewDTO{},
	}, nil
}

func TestAskKnowledgeHandler_FillsSourceHints(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/knowledge/ask", strings.NewReader(`{"question":"qual a saude do ms-auth","context":"ops"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer jwt-token")
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.Set("token", "jwt-token")

	knowledge := &stubKnowledgeClient{}
	collector := &oauthStubCollector{agents: []appdto.CollectorAgentRaw{
		{ID: "a1", Name: "Health snapshot", Context: "ops", Prompt: "dica da fonte ops"},
		{ID: "a2", Name: "Juridico", Context: "juridico", Prompt: "dica juridica"},
	}}
	h := NewKnowledgeHandlers(knowledge, collector, &oauthStubCompany{id: "company-1"}, &stubServiceToken{}, zap.NewNop())
	if err := h.AskKnowledgeHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if knowledge.last.Question != "qual a saude do ms-auth" {
		t.Fatalf("unexpected question: %q", knowledge.last.Question)
	}
	if knowledge.lastBearer != "Bearer oauth-service" {
		t.Fatalf("expected service token, got %q", knowledge.lastBearer)
	}
	if len(knowledge.last.SourceHints) != 1 || knowledge.last.SourceHints[0].Prompt != "dica da fonte ops" {
		t.Fatalf("expected ops sourceHint, got %+v", knowledge.last.SourceHints)
	}
	var out appdto.KnowledgeAskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Mode != "HEURISTIC" || out.Answer == "" {
		t.Fatalf("unexpected response: %+v", out)
	}
}

func TestAskKnowledgeHandler_RejectsWithoutRole(t *testing.T) {
	e := echo.New()
	h := NewKnowledgeHandlers(&stubKnowledgeClient{}, &oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, &stubServiceToken{}, zap.NewNop())
	handler := middlewarePkg.RequireKnowledgeRead()(h.AskKnowledgeHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/knowledge/ask", strings.NewReader(`{"question":"saude"}`))
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

func TestAskKnowledgeHandler_AllowsKnowledgeReadAuthority(t *testing.T) {
	e := echo.New()
	knowledge := &stubKnowledgeClient{}
	h := NewKnowledgeHandlers(knowledge, &oauthStubCollector{}, &oauthStubCompany{id: "company-1"}, &stubServiceToken{}, zap.NewNop())
	handler := middlewarePkg.RequireKnowledgeRead()(h.AskKnowledgeHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/knowledge/ask", strings.NewReader(`{"question":"saude"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"USER"}, Authorities: []string{"knowledge:read"}, TenantId: "tenant-1"})
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAskKnowledgeHandler_FreshnessFailedFromLastExecution(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/knowledge/ask", strings.NewReader(`{"question":"saude"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})

	finished := time.Now().UTC().Add(-12 * time.Minute).Format(time.RFC3339)
	knowledge := &stubKnowledgeClient{sources: []appdto.KnowledgeAskSource{
		{Kind: "FACT", Key: "ms-auth", SourceAgentID: "a1", AgentName: "Health snapshot", DocumentID: "doc-1"},
	}}
	collector := &oauthStubCollector{
		agents: []appdto.CollectorAgentRaw{{ID: "a1", Name: "Health snapshot", Context: "ops"}},
		executions: []appdto.CollectorExecutionRaw{{
			ID:           "exec-1",
			AgentID:      "a1",
			StartedAt:    time.Now().UTC().Add(-13 * time.Minute).Format(time.RFC3339),
			FinishedAt:   &finished,
			Status:       "FAILED",
			ErrorMessage: "timeout no target",
		}},
	}
	h := NewKnowledgeHandlers(knowledge, collector, &oauthStubCompany{id: "company-1"}, &stubServiceToken{}, zap.NewNop())
	if err := h.AskKnowledgeHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out appdto.KnowledgeAskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Answer == "" {
		t.Fatalf("briefing should remain visible, got %+v", out)
	}
	if out.Freshness == nil || !out.Freshness.Failed {
		t.Fatalf("expected freshness.failed=true, got %+v", out.Freshness)
	}
	if out.Freshness.Status != "FAILED" || out.Freshness.AgentID != "a1" {
		t.Fatalf("unexpected freshness: %+v", out.Freshness)
	}
	if out.Freshness.AgeMinutes < 11 || out.Freshness.AgeMinutes > 13 {
		t.Fatalf("expected ageMinutes around 12, got %d", out.Freshness.AgeMinutes)
	}
}

func TestAskKnowledgeHandler_CollectorDownOmitsFreshness(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/knowledge/ask", strings.NewReader(`{"question":"saude"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})

	knowledge := &stubKnowledgeClient{sources: []appdto.KnowledgeAskSource{
		{Kind: "FACT", Key: "ms-auth", SourceAgentID: "a1"},
	}}
	collector := &oauthStubCollector{execErr: errors.New("collector down")}
	h := NewKnowledgeHandlers(knowledge, collector, &oauthStubCompany{id: "company-1"}, &stubServiceToken{}, zap.NewNop())
	if err := h.AskKnowledgeHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out appdto.KnowledgeAskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Answer == "" {
		t.Fatal("briefing should still return when collector is down")
	}
	if out.Freshness != nil {
		t.Fatalf("expected empty freshness, got %+v", out.Freshness)
	}
}
