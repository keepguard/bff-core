package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	domainclient "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

type llmClient struct {
	httpClient *resty.Client
	baseURL    string
	tokens     domainclient.ServiceTokenClient
	logger     *zap.Logger
}

func NewLlmClient(cfg *config.Config, tokens domainclient.ServiceTokenClient, logger *zap.Logger) domainclient.LlmClient {
	timeout := cfg.Services.Llm.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	httpClient := resty.New()
	httpClient.SetTimeout(timeout)
	httpClient.SetRetryCount(1)
	httpClient.SetRetryWaitTime(200 * time.Millisecond)
	return &llmClient{httpClient: httpClient, baseURL: cfg.Services.Llm.BaseURL, tokens: tokens, logger: logger}
}

// headers monta a request com a identidade de serviço do BFF: o gateway LLM
// autoriza pelas authorities do client OAuth bff-core (llm:read/llm:write), não
// pelo token de quem está logado no backoffice — esse já foi autorizado nas
// middlewares de entrada.
func (c *llmClient) headers(ctx context.Context, req *resty.Request, tenantID, correlationID string) (*resty.Request, error) {
	req.SetContext(ctx)
	req.SetHeader("X-Tenant-Id", tenantID).SetHeader("X-Correlation-ID", correlationID)
	bearer, err := c.serviceBearer(ctx)
	if err != nil {
		return nil, err
	}
	req.SetHeader("Authorization", bearer)
	return req, nil
}

func (c *llmClient) serviceBearer(ctx context.Context) (string, error) {
	if c.tokens == nil {
		return "", MapHTTPError(503, []byte(`{"message":"Token de serviço do BFF indisponível"}`), "llm gateway")
	}
	companyID := companyHeader(ctx)
	if strings.TrimSpace(companyID) == "" {
		return "", MapHTTPError(400, []byte(`{"message":"company não resolvida para o tenant informado"}`), "llm gateway")
	}
	token, err := c.tokens.GetToken(ctx, companyID)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Erro ao obter token OAuth do BFF para o gateway LLM", zap.Error(err))
		}
		return "", MapHTTPError(503, []byte(`{"message":"Gateway LLM indisponível"}`), "llm gateway")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", MapHTTPError(503, []byte(`{"message":"Gateway LLM indisponível"}`), "llm gateway")
	}
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}
	return token, nil
}

func (c *llmClient) ListProviders(ctx context.Context, tenantID, correlationID string) (json.RawMessage, error) {
	return c.getRaw(ctx, tenantID, correlationID, "/api/v1/llm/providers", nil)
}

func (c *llmClient) CreateProvider(ctx context.Context, tenantID, correlationID string, body any) (json.RawMessage, error) {
	return c.sendJSON(ctx, tenantID, correlationID, "POST", "/api/v1/llm/providers", body, 201)
}

func (c *llmClient) UpdateProvider(ctx context.Context, tenantID, correlationID, id string, body any) (json.RawMessage, error) {
	return c.sendJSON(ctx, tenantID, correlationID, "PATCH", "/api/v1/llm/providers/"+id, body, 200)
}

func (c *llmClient) SetProviderEnabled(ctx context.Context, tenantID, correlationID, id string, enabled bool) (json.RawMessage, error) {
	action := "enable"
	if !enabled {
		action = "disable"
	}
	return c.sendJSON(ctx, tenantID, correlationID, "POST", fmt.Sprintf("/api/v1/llm/providers/%s/%s", id, action), nil, 200)
}

func (c *llmClient) SetProviderDefault(ctx context.Context, tenantID, correlationID, id string) (json.RawMessage, error) {
	return c.sendJSON(ctx, tenantID, correlationID, "POST", fmt.Sprintf("/api/v1/llm/providers/%s/default", id), nil, 200)
}

func (c *llmClient) Complete(ctx context.Context, tenantID, companyID, correlationID string, body any) (json.RawMessage, error) {
	req, err := c.headers(ctx, c.httpClient.R(), tenantID, correlationID)
	if err != nil {
		return nil, err
	}
	req.SetHeader("X-Company-Id", companyID).SetBody(body)
	resp, err := req.Post(c.baseURL + "/api/v1/llm/complete")
	if err != nil {
		return nil, MapNetworkError(err, "llm gateway")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "llm gateway")
	}
	return cloneBody(resp.Body()), nil
}

func (c *llmClient) ListUsage(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedLlmUsageResponse, error) {
	var out appdto.PaginatedLlmUsageResponse
	req, err := c.headers(ctx, c.httpClient.R(), tenantID, correlationID)
	if err != nil {
		return appdto.PaginatedLlmUsageResponse{}, err
	}
	req.SetResult(&out)
	for key, value := range query {
		if value != "" {
			req.SetQueryParam(key, value)
		}
	}
	resp, err := req.Get(c.baseURL + "/api/v1/llm/usage")
	if err != nil {
		return appdto.PaginatedLlmUsageResponse{}, MapNetworkError(err, "llm gateway")
	}
	if resp.StatusCode() != 200 {
		return appdto.PaginatedLlmUsageResponse{}, MapHTTPError(resp.StatusCode(), resp.Body(), "llm gateway")
	}
	return out, nil
}

func (c *llmClient) GetUsage(ctx context.Context, tenantID, correlationID, id string) (appdto.LlmUsageResponse, error) {
	var out appdto.LlmUsageResponse
	req, err := c.headers(ctx, c.httpClient.R(), tenantID, correlationID)
	if err != nil {
		return appdto.LlmUsageResponse{}, err
	}
	resp, err := req.SetResult(&out).Get(c.baseURL + "/api/v1/llm/usage/" + id)
	if err != nil {
		return appdto.LlmUsageResponse{}, MapNetworkError(err, "llm gateway")
	}
	if resp.StatusCode() != 200 {
		return appdto.LlmUsageResponse{}, MapHTTPError(resp.StatusCode(), resp.Body(), "llm gateway")
	}
	return out, nil
}

func (c *llmClient) ListAlertRules(ctx context.Context, tenantID, correlationID string) (json.RawMessage, error) {
	return c.getRaw(ctx, tenantID, correlationID, "/api/v1/llm/alert-rules", nil)
}

func (c *llmClient) CreateAlertRule(ctx context.Context, tenantID, correlationID string, body any) (json.RawMessage, error) {
	return c.sendJSON(ctx, tenantID, correlationID, "POST", "/api/v1/llm/alert-rules", body, 201)
}

func (c *llmClient) UpdateAlertRule(ctx context.Context, tenantID, correlationID, id string, body any) (json.RawMessage, error) {
	return c.sendJSON(ctx, tenantID, correlationID, "PATCH", "/api/v1/llm/alert-rules/"+id, body, 200)
}

func (c *llmClient) SetAlertRuleEnabled(ctx context.Context, tenantID, correlationID, id string, enabled bool) (json.RawMessage, error) {
	action := "enable"
	if !enabled {
		action = "disable"
	}
	return c.sendJSON(ctx, tenantID, correlationID, "POST", fmt.Sprintf("/api/v1/llm/alert-rules/%s/%s", id, action), nil, 200)
}

func (c *llmClient) ListAlertFirings(ctx context.Context, tenantID, correlationID string, query map[string]string) (json.RawMessage, error) {
	return c.getRaw(ctx, tenantID, correlationID, "/api/v1/llm/alert-firings", query)
}

func (c *llmClient) getRaw(ctx context.Context, tenantID, correlationID, path string, query map[string]string) (json.RawMessage, error) {
	req, err := c.headers(ctx, c.httpClient.R(), tenantID, correlationID)
	if err != nil {
		return nil, err
	}
	for key, value := range query {
		if value != "" {
			req.SetQueryParam(key, value)
		}
	}
	resp, err := req.Get(c.baseURL + path)
	if err != nil {
		return nil, MapNetworkError(err, "llm gateway")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "llm gateway")
	}
	return cloneBody(resp.Body()), nil
}

func (c *llmClient) sendJSON(ctx context.Context, tenantID, correlationID, method, path string, body any, want int) (json.RawMessage, error) {
	req, err := c.headers(ctx, c.httpClient.R(), tenantID, correlationID)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.SetBody(body)
	}
	var resp *resty.Response
	url := c.baseURL + path
	switch method {
	case "POST":
		resp, err = req.Post(url)
	case "PATCH":
		resp, err = req.Patch(url)
	default:
		resp, err = req.Get(url)
	}
	if err != nil {
		return nil, MapNetworkError(err, "llm gateway")
	}
	if resp.StatusCode() != want {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "llm gateway")
	}
	return cloneBody(resp.Body()), nil
}

func cloneBody(body []byte) json.RawMessage {
	if len(body) == 0 {
		return json.RawMessage("null")
	}
	out := make([]byte, len(body))
	copy(out, body)
	return json.RawMessage(out)
}
