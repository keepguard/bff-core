package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type stubOAuthClient struct {
	blocked   bool
	deleted   bool
	unblocked bool
}

func (s *stubOAuthClient) Search(_ context.Context, _, _, _ string, _ map[string]string) (appdto.PaginatedOAuthClients, error) {
	return appdto.PaginatedOAuthClients{Content: []appdto.OAuthClientDTO{}, Size: 20}, nil
}

func (s *stubOAuthClient) GetByID(_ context.Context, _, _, _, id string) (appdto.OAuthClientDTO, error) {
	return appdto.OAuthClientDTO{ID: id, ClientID: "investbot-collector", Status: "ACTIVE"}, nil
}

func (s *stubOAuthClient) Create(_ context.Context, _, _, _ string, body appdto.OAuthClientCreateRequest) (appdto.OAuthClientDTO, error) {
	return appdto.OAuthClientDTO{ID: "new-id", ClientID: body.ClientID, Status: "ACTIVE"}, nil
}

func (s *stubOAuthClient) Block(_ context.Context, _, _, _, _ string) (appdto.OAuthClientDTO, error) {
	s.blocked = true
	return appdto.OAuthClientDTO{ID: "c1", ClientID: "investbot-collector", Status: "BLOCKED"}, nil
}

func (s *stubOAuthClient) Unblock(_ context.Context, _, _, _, _ string) (appdto.OAuthClientDTO, error) {
	s.unblocked = true
	return appdto.OAuthClientDTO{ID: "c1", ClientID: "investbot-collector", Status: "ACTIVE"}, nil
}

func (s *stubOAuthClient) Delete(_ context.Context, _, _, _, _ string) error {
	s.deleted = true
	return nil
}

type oauthStubCompany struct {
	id string
}

func (s *oauthStubCompany) GetByTenantId(_ context.Context, _, _ string) (companyDto.MSCompanyResponseDTO, error) {
	return companyDto.MSCompanyResponseDTO{ID: s.id, TenantId: "tenant-1"}, nil
}

type oauthStubCollector struct {
	agents      []appdto.CollectorAgentRaw
	disabledIDs []string
}

func (s *oauthStubCollector) ListAgents(_ context.Context, _, _ string) ([]appdto.CollectorAgentRaw, error) {
	return s.agents, nil
}

func (s *oauthStubCollector) DisableAgent(_ context.Context, _, agentID, _ string) error {
	s.disabledIDs = append(s.disabledIDs, agentID)
	return nil
}

func oauthContext(method, path string, companyID string, claims *pkg.JWTClaims) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	if companyID != "" {
		req = req.WithContext(client.WithCompanyID(req.Context(), companyID))
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if claims != nil {
		c.Set("claims", claims)
	}
	c.Set("token", "jwt-token")
	return c, rec
}

func TestListOAuthClientsHandler_UsesCompanyFromJWTContext(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/oauth/clients", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, &oauthStubCollector{}, zap.NewNop())
	if err := h.ListOAuthClientsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListOAuthClientsHandler_ResolvesCompanyFromJWTWhenContextEmpty(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/oauth/clients", "", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-from-jwt"}, &oauthStubCollector{}, zap.NewNop())
	if err := h.ListOAuthClientsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListOAuthClientsHandler_RequiresJWTTenant(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/oauth/clients", "", &pkg.JWTClaims{Roles: []string{"ADMIN"}})
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, &oauthStubCollector{}, zap.NewNop())
	if err := h.ListOAuthClientsHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBlockOAuthClientHandler_DisablesEnabledAgents(t *testing.T) {
	c, rec := oauthContext(http.MethodPost, "/api/v1/core/oauth/clients/c1/block", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("c1")
	collector := &oauthStubCollector{agents: []appdto.CollectorAgentRaw{
		{ID: "a1", Name: "Coletor A", Enabled: true, CollectorType: "API_REST"},
		{ID: "a2", Name: "Coletor B", Enabled: false, CollectorType: "HTML_SCRAPER"},
	}}
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, collector, zap.NewNop())
	if err := h.BlockOAuthClientHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if len(collector.disabledIDs) != 1 || collector.disabledIDs[0] != "a1" {
		t.Fatalf("expected only enabled agent a1 to be disabled, got %v", collector.disabledIDs)
	}
}

func TestUnblockOAuthClientHandler_DoesNotDisableAgents(t *testing.T) {
	c, rec := oauthContext(http.MethodPost, "/api/v1/core/oauth/clients/c1/unblock", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("c1")
	collector := &oauthStubCollector{agents: []appdto.CollectorAgentRaw{
		{ID: "a1", Name: "Coletor A", Enabled: true},
	}}
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, collector, zap.NewNop())
	if err := h.UnblockOAuthClientHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if len(collector.disabledIDs) != 0 {
		t.Fatalf("unblock should not disable agents, got %v", collector.disabledIDs)
	}
}

func TestDeleteOAuthClientHandler_DisablesEnabledAgents(t *testing.T) {
	c, rec := oauthContext(http.MethodDelete, "/api/v1/core/oauth/clients/c1", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("c1")
	collector := &oauthStubCollector{agents: []appdto.CollectorAgentRaw{
		{ID: "a1", Enabled: true},
		{ID: "a3", Enabled: true},
	}}
	oauth := &stubOAuthClient{}
	h := NewOAuthClientHandlers(oauth, &oauthStubCompany{id: "company-1"}, collector, zap.NewNop())
	if err := h.DeleteOAuthClientHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !oauth.deleted {
		t.Fatal("expected oauth client to be deleted")
	}
	if strings.Join(collector.disabledIDs, ",") != "a1,a3" {
		t.Fatalf("expected a1,a3 disabled, got %v", collector.disabledIDs)
	}
}

func TestGetOAuthClientHandler_ReturnsAgents(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/oauth/clients/c1", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("c1")
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, &oauthStubCollector{
		agents: []appdto.CollectorAgentRaw{{ID: "a1", Name: "Coletor A", Enabled: true, CollectorType: "API_REST"}},
	}, zap.NewNop())
	if err := h.GetOAuthClientHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body appdto.OAuthClientDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Agents) != 1 || body.Agents[0].Name != "Coletor A" {
		t.Fatalf("expected agent list in response, got %+v", body.Agents)
	}
}
