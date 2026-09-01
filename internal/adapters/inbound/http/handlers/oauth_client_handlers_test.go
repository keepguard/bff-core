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
	return appdto.OAuthClientDTO{ID: "new-id", ClientID: body.ClientID, Status: "ACTIVE", ServiceRoleID: body.RoleID}, nil
}

func (s *stubOAuthClient) Update(_ context.Context, _, _, _, id string, body appdto.OAuthClientUpdateRequest) (appdto.OAuthClientDTO, error) {
	ttl := 28800
	if body.TokenTTLSeconds != nil {
		ttl = *body.TokenTTLSeconds
	}
	return appdto.OAuthClientDTO{
		ID:              id,
		ClientID:        "investbot-collector",
		Status:          "ACTIVE",
		Description:     body.Description,
		ServiceRoleID:   body.RoleID,
		TokenTTLSeconds: ttl,
	}, nil
}

func (s *stubOAuthClient) ListServiceRoles(_ context.Context, _, _, _ string) ([]appdto.OAuthServiceRoleDTO, error) {
	return []appdto.OAuthServiceRoleDTO{{
		ID:   "role-1",
		Name: "ROLE_SERVICE_COLLECTOR",
		Authorities: []appdto.OAuthServiceRoleAuthorityDTO{
			{Name: "knowledge:write", Description: "grava"},
			{Name: "knowledge:read", Description: "lê"},
		},
	}}, nil
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
	listErr     error
	getErr      error
	executions  []appdto.CollectorExecutionRaw
	execErr     error
	sources     []appdto.CollectorDataSourceRaw
}

func (s *oauthStubCollector) ListAgents(_ context.Context, _, _ string) ([]appdto.CollectorAgentRaw, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.agents, nil
}

func (s *oauthStubCollector) SearchAgents(_ context.Context, _, _ string, _ map[string]string) (appdto.PaginatedCollectorAgentsRaw, error) {
	return appdto.PaginatedCollectorAgentsRaw{Content: s.agents, Size: 20, TotalPages: 1, First: true, Last: true}, nil
}

func (s *oauthStubCollector) GetAgent(_ context.Context, _, agentID, _ string) (appdto.CollectorAgentRaw, error) {
	if s.getErr != nil {
		return appdto.CollectorAgentRaw{}, s.getErr
	}
	for _, agent := range s.agents {
		if agent.ID == agentID {
			return agent, nil
		}
	}
	return appdto.CollectorAgentRaw{ID: agentID}, nil
}

func (s *oauthStubCollector) CreateAgent(_ context.Context, _, _ string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error) {
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	context := "geral"
	if body.Context != nil && *body.Context != "" {
		context = *body.Context
	}
	return appdto.CollectorAgentRaw{
		ID:            "new-agent",
		Name:          body.Name,
		Context:       context,
		CollectorType: body.CollectorType,
		Enabled:       enabled,
	}, nil
}

func (s *oauthStubCollector) UpdateAgent(_ context.Context, _, agentID, _ string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error) {
	return appdto.CollectorAgentRaw{ID: agentID, Name: body.Name, Enabled: true}, nil
}

func (s *oauthStubCollector) EnableAgent(_ context.Context, _, agentID, _ string) (appdto.CollectorAgentRaw, error) {
	return appdto.CollectorAgentRaw{ID: agentID, Enabled: true}, nil
}

func (s *oauthStubCollector) DisableAgent(_ context.Context, _, agentID, _ string) (appdto.CollectorAgentRaw, error) {
	s.disabledIDs = append(s.disabledIDs, agentID)
	return appdto.CollectorAgentRaw{ID: agentID, Enabled: false}, nil
}

func (s *oauthStubCollector) TestAgent(_ context.Context, _, agentID, _ string) (appdto.CollectorAgentTestResultDTO, error) {
	return appdto.CollectorAgentTestResultDTO{
		Success:        true,
		AgentID:        agentID,
		ItemsCollected: 1,
		Preview:        []appdto.CollectorAgentTestPreviewDTO{},
	}, nil
}

func (s *oauthStubCollector) ListAgentExecutions(_ context.Context, _, _, _ string, _ int) ([]appdto.CollectorExecutionRaw, error) {
	if s.execErr != nil {
		return nil, s.execErr
	}
	if s.executions == nil {
		return []appdto.CollectorExecutionRaw{}, nil
	}
	return s.executions, nil
}

func (s *oauthStubCollector) GetExecution(_ context.Context, _, executionID, _ string) (appdto.CollectorExecutionRaw, error) {
	if s.execErr != nil {
		return appdto.CollectorExecutionRaw{}, s.execErr
	}
	for _, execution := range s.executions {
		if execution.ID == executionID {
			return execution, nil
		}
	}
	if len(s.executions) == 1 {
		return s.executions[0], nil
	}
	return appdto.CollectorExecutionRaw{ID: executionID, AgentID: "a1"}, nil
}

func (s *oauthStubCollector) DeleteAgent(_ context.Context, _, agentID, _ string) error {
	s.disabledIDs = append(s.disabledIDs, "deleted:"+agentID)
	return nil
}

func (s *oauthStubCollector) ListDataSources(_ context.Context, _, _ string, _ map[string]string) ([]appdto.CollectorDataSourceRaw, error) {
	if s.sources == nil {
		return []appdto.CollectorDataSourceRaw{}, nil
	}
	return s.sources, nil
}

func (s *oauthStubCollector) GetDataSource(_ context.Context, _, sourceID, _ string) (appdto.CollectorDataSourceRaw, error) {
	for _, source := range s.sources {
		if source.ID == sourceID {
			return source, nil
		}
	}
	if len(s.sources) == 1 {
		return s.sources[0], nil
	}
	return appdto.CollectorDataSourceRaw{ID: sourceID}, nil
}

func (s *oauthStubCollector) CreateDataSource(_ context.Context, _, _ string, body appdto.CollectorDataSourceWriteRaw) (appdto.CollectorDataSourceRaw, error) {
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	created := appdto.CollectorDataSourceRaw{
		ID:            "src-created",
		Slug:          body.Slug,
		Name:          body.Name,
		CollectorType: body.CollectorType,
		Enabled:       enabled,
		Scope:         "company",
		CompanyID:     "company-1",
	}
	s.sources = append(s.sources, created)
	return created, nil
}

func (s *oauthStubCollector) UpdateDataSource(_ context.Context, _, sourceID, _ string, body appdto.CollectorDataSourceWriteRaw) (appdto.CollectorDataSourceRaw, error) {
	return appdto.CollectorDataSourceRaw{ID: sourceID, Name: body.Name, Slug: body.Slug, Scope: "company", Enabled: true}, nil
}

func (s *oauthStubCollector) EnableDataSource(_ context.Context, _, sourceID, _ string) (appdto.CollectorDataSourceRaw, error) {
	return appdto.CollectorDataSourceRaw{ID: sourceID, Enabled: true, Scope: "company"}, nil
}

func (s *oauthStubCollector) DisableDataSource(_ context.Context, _, sourceID, _ string) (appdto.CollectorDataSourceRaw, error) {
	return appdto.CollectorDataSourceRaw{ID: sourceID, Enabled: false, Scope: "company"}, nil
}

func (s *oauthStubCollector) DeleteDataSource(_ context.Context, _, sourceID, _ string) error {
	s.disabledIDs = append(s.disabledIDs, "deleted-source:"+sourceID)
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

func TestGetOAuthClientHandler_ReportsCollectorUnavailable(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/oauth/clients/c1", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	c.SetParamNames("id")
	c.SetParamValues("c1")
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, &oauthStubCollector{
		listErr: echo.NewHTTPError(http.StatusBadGateway, "collector down"),
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
	if body.AgentsLoadError == "" {
		t.Fatalf("expected agentsLoadError in response, got %+v", body)
	}
	if len(body.Agents) != 0 {
		t.Fatalf("expected empty agents when collector fails, got %+v", body.Agents)
	}
}

func TestListOAuthServiceRolesHandler_ReturnsCatalog(t *testing.T) {
	c, rec := oauthContext(http.MethodGet, "/api/v1/core/oauth/clients/service-roles", "company-1", &pkg.JWTClaims{
		Roles:    []string{"ADMIN"},
		TenantId: "tenant-1",
	})
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, &oauthStubCollector{}, zap.NewNop())
	if err := h.ListOAuthServiceRolesHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body []appdto.OAuthServiceRoleDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 || body[0].Name != "ROLE_SERVICE_COLLECTOR" {
		t.Fatalf("expected collector service role, got %+v", body)
	}
}

func TestCreateOAuthClientHandler_RequiresRoleID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/oauth/clients", strings.NewReader(`{"clientId":"srv-x"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.Set("token", "jwt-token")
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, &oauthStubCollector{}, zap.NewNop())
	if err := h.CreateOAuthClientHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "MISSING_ROLE_ID") {
		t.Fatalf("expected MISSING_ROLE_ID, got %s", rec.Body.String())
	}
}

func TestUpdateOAuthClientHandler_RequiresRoleID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/core/oauth/clients/c1", strings.NewReader(`{"description":"x","tokenTtlSeconds":3600}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("c1")
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.Set("token", "jwt-token")
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, &oauthStubCollector{}, zap.NewNop())
	if err := h.UpdateOAuthClientHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "MISSING_ROLE_ID") {
		t.Fatalf("expected MISSING_ROLE_ID, got %s", rec.Body.String())
	}
}

func TestUpdateOAuthClientHandler_ValidBodyDoesNotDisableAgents(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/core/oauth/clients/c1", strings.NewReader(`{"description":"collector","roleId":"role-1","tokenTtlSeconds":3600}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(client.WithCompanyID(req.Context(), "company-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("c1")
	c.Set("claims", &pkg.JWTClaims{Roles: []string{"ADMIN"}, TenantId: "tenant-1"})
	c.Set("token", "jwt-token")
	collector := &oauthStubCollector{agents: []appdto.CollectorAgentRaw{
		{ID: "a1", Name: "Coletor A", Enabled: true},
	}}
	h := NewOAuthClientHandlers(&stubOAuthClient{}, &oauthStubCompany{id: "company-1"}, collector, zap.NewNop())
	if err := h.UpdateOAuthClientHandler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body appdto.OAuthClientDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ServiceRoleID != "role-1" || body.Description != "collector" || body.TokenTTLSeconds != 3600 {
		t.Fatalf("expected updated fields, got %+v", body)
	}
	if len(collector.disabledIDs) != 0 {
		t.Fatalf("update should not disable agents, got %v", collector.disabledIDs)
	}
}
