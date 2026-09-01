package client

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

type collectorHTTP struct {
	httpClient *resty.Client
	baseURL    string
	logger     *zap.Logger
}

func NewCollectorClient(cfg *config.Config, logger *zap.Logger) domainclient.CollectorClient {
	timeout := cfg.Services.Collector.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	httpClient := resty.New()
	httpClient.SetTimeout(timeout)
	httpClient.SetRetryCount(0)
	return &collectorHTTP{
		httpClient: httpClient,
		baseURL:    cfg.Services.Collector.BaseURL,
		logger:     logger,
	}
}

func (c *collectorHTTP) agentURL(id string) string {
	base := c.baseURL + "/api/v1/collector/agents"
	if id == "" {
		return base
	}
	return base + "/" + id
}

func (c *collectorHTTP) req(ctx context.Context, companyID, correlationID string) *resty.Request {
	return c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID)
}

func (c *collectorHTTP) DisableAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error) {
	return c.toggleAgent(ctx, companyID, agentID, correlationID, "disable")
}

func (c *collectorHTTP) EnableAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error) {
	return c.toggleAgent(ctx, companyID, agentID, correlationID, "enable")
}

func (c *collectorHTTP) toggleAgent(ctx context.Context, companyID, agentID, correlationID, action string) (appdto.CollectorAgentRaw, error) {
	if c.baseURL == "" || agentID == "" {
		return appdto.CollectorAgentRaw{}, nil
	}
	var out appdto.CollectorAgentRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetResult(&out).
		Post(c.agentURL(agentID) + "/" + action)
	if err != nil {
		return appdto.CollectorAgentRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
		return appdto.CollectorAgentRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return out, nil
}

func (c *collectorHTTP) ListAgents(ctx context.Context, companyID, correlationID string) ([]appdto.CollectorAgentRaw, error) {
	if c.baseURL == "" {
		return []appdto.CollectorAgentRaw{}, nil
	}
	var out []appdto.CollectorAgentRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetResult(&out).
		Get(c.agentURL(""))
	if err != nil {
		return nil, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	if out == nil {
		out = []appdto.CollectorAgentRaw{}
	}
	return out, nil
}

func (c *collectorHTTP) SearchAgents(ctx context.Context, companyID, correlationID string, query map[string]string) (appdto.PaginatedCollectorAgentsRaw, error) {
	if c.baseURL == "" {
		return appdto.PaginatedCollectorAgentsRaw{Content: []appdto.CollectorAgentRaw{}}, nil
	}
	var out appdto.PaginatedCollectorAgentsRaw
	req := c.req(ctx, companyID, correlationID).SetResult(&out)
	for key, value := range query {
		if value != "" {
			req.SetQueryParam(key, value)
		}
	}
	resp, err := req.Get(c.agentURL(""))
	if err != nil {
		return appdto.PaginatedCollectorAgentsRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return appdto.PaginatedCollectorAgentsRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	if out.Content == nil {
		out.Content = []appdto.CollectorAgentRaw{}
	}
	return out, nil
}

func (c *collectorHTTP) GetAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error) {
	if c.baseURL == "" {
		return appdto.CollectorAgentRaw{}, nil
	}
	var out appdto.CollectorAgentRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetResult(&out).
		Get(c.agentURL(agentID))
	if err != nil {
		return appdto.CollectorAgentRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return appdto.CollectorAgentRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return out, nil
}

func (c *collectorHTTP) CreateAgent(ctx context.Context, companyID, correlationID string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error) {
	if c.baseURL == "" {
		return appdto.CollectorAgentRaw{}, nil
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return appdto.CollectorAgentRaw{}, err
	}
	var out appdto.CollectorAgentRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&out).
		Post(c.agentURL(""))
	if err != nil {
		return appdto.CollectorAgentRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		return appdto.CollectorAgentRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return out, nil
}

func (c *collectorHTTP) UpdateAgent(ctx context.Context, companyID, agentID, correlationID string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error) {
	if c.baseURL == "" {
		return appdto.CollectorAgentRaw{}, nil
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return appdto.CollectorAgentRaw{}, err
	}
	var out appdto.CollectorAgentRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&out).
		Put(c.agentURL(agentID))
	if err != nil {
		return appdto.CollectorAgentRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return appdto.CollectorAgentRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return out, nil
}

func (c *collectorHTTP) DeleteAgent(ctx context.Context, companyID, agentID, correlationID string) error {
	if c.baseURL == "" || agentID == "" {
		return nil
	}
	resp, err := c.req(ctx, companyID, correlationID).Delete(c.agentURL(agentID))
	if err != nil {
		return MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return nil
}

func (c *collectorHTTP) ListAgentExecutions(ctx context.Context, companyID, agentID, correlationID string, limit int) ([]appdto.CollectorExecutionRaw, error) {
	if c.baseURL == "" || agentID == "" {
		return []appdto.CollectorExecutionRaw{}, nil
	}
	if _, err := c.GetAgent(ctx, companyID, agentID, correlationID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	var out []appdto.CollectorExecutionRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetResult(&out).
		Get(c.agentURL(agentID) + "/executions")
	if err != nil {
		return nil, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	if out == nil {
		out = []appdto.CollectorExecutionRaw{}
	}
	return out, nil
}

func (c *collectorHTTP) GetExecution(ctx context.Context, companyID, executionID, correlationID string) (appdto.CollectorExecutionRaw, error) {
	if c.baseURL == "" || executionID == "" {
		return appdto.CollectorExecutionRaw{}, nil
	}
	var out appdto.CollectorExecutionRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetResult(&out).
		Get(c.baseURL + "/api/v1/collector/executions/" + executionID)
	if err != nil {
		return appdto.CollectorExecutionRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return appdto.CollectorExecutionRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	if out.AgentID == "" {
		return out, nil
	}
	if _, agentErr := c.GetAgent(ctx, companyID, out.AgentID, correlationID); agentErr != nil {
		return appdto.CollectorExecutionRaw{}, agentErr
	}
	return out, nil
}

func (c *collectorHTTP) TestAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentTestResultDTO, error) {
	if c.baseURL == "" || agentID == "" {
		return appdto.CollectorAgentTestResultDTO{}, nil
	}
	testClient := resty.New()
	testClient.SetTimeout(60 * time.Second)
	testClient.SetRetryCount(0)

	var raw appdto.CollectorAgentTestResultRaw
	resp, err := testClient.R().
		SetContext(ctx).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]any{}).
		SetResult(&raw).
		Post(c.agentURL(agentID) + "/test")
	if err != nil {
		return appdto.CollectorAgentTestResultDTO{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return appdto.CollectorAgentTestResultDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return appdto.MapCollectorAgentTestResult(raw), nil
}

func (c *collectorHTTP) dataSourceURL(id string) string {
	base := c.baseURL + "/api/v1/collector/data-sources"
	if id == "" {
		return base
	}
	return base + "/" + id
}

func (c *collectorHTTP) ListDataSources(ctx context.Context, companyID, correlationID string, query map[string]string) ([]appdto.CollectorDataSourceRaw, error) {
	if c.baseURL == "" {
		return []appdto.CollectorDataSourceRaw{}, nil
	}
	var out []appdto.CollectorDataSourceRaw
	req := c.req(ctx, companyID, correlationID).SetResult(&out)
	for key, value := range query {
		if value != "" {
			req.SetQueryParam(key, value)
		}
	}
	resp, err := req.Get(c.dataSourceURL(""))
	if err != nil {
		return nil, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	if out == nil {
		out = []appdto.CollectorDataSourceRaw{}
	}
	return out, nil
}

func (c *collectorHTTP) GetDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error) {
	if c.baseURL == "" {
		return appdto.CollectorDataSourceRaw{}, nil
	}
	var out appdto.CollectorDataSourceRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetResult(&out).
		Get(c.dataSourceURL(sourceID))
	if err != nil {
		return appdto.CollectorDataSourceRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return appdto.CollectorDataSourceRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return out, nil
}

func (c *collectorHTTP) CreateDataSource(ctx context.Context, companyID, correlationID string, body appdto.CollectorDataSourceWriteRaw) (appdto.CollectorDataSourceRaw, error) {
	if c.baseURL == "" {
		return appdto.CollectorDataSourceRaw{}, nil
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return appdto.CollectorDataSourceRaw{}, err
	}
	var out appdto.CollectorDataSourceRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&out).
		Post(c.dataSourceURL(""))
	if err != nil {
		return appdto.CollectorDataSourceRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		return appdto.CollectorDataSourceRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return out, nil
}

func (c *collectorHTTP) UpdateDataSource(ctx context.Context, companyID, sourceID, correlationID string, body appdto.CollectorDataSourceWriteRaw) (appdto.CollectorDataSourceRaw, error) {
	if c.baseURL == "" {
		return appdto.CollectorDataSourceRaw{}, nil
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return appdto.CollectorDataSourceRaw{}, err
	}
	var out appdto.CollectorDataSourceRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&out).
		Put(c.dataSourceURL(sourceID))
	if err != nil {
		return appdto.CollectorDataSourceRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return appdto.CollectorDataSourceRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return out, nil
}

func (c *collectorHTTP) EnableDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error) {
	return c.toggleDataSource(ctx, companyID, sourceID, correlationID, "enable")
}

func (c *collectorHTTP) DisableDataSource(ctx context.Context, companyID, sourceID, correlationID string) (appdto.CollectorDataSourceRaw, error) {
	return c.toggleDataSource(ctx, companyID, sourceID, correlationID, "disable")
}

func (c *collectorHTTP) toggleDataSource(ctx context.Context, companyID, sourceID, correlationID, action string) (appdto.CollectorDataSourceRaw, error) {
	if c.baseURL == "" || sourceID == "" {
		return appdto.CollectorDataSourceRaw{}, nil
	}
	var out appdto.CollectorDataSourceRaw
	resp, err := c.req(ctx, companyID, correlationID).
		SetResult(&out).
		Post(c.dataSourceURL(sourceID) + "/" + action)
	if err != nil {
		return appdto.CollectorDataSourceRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
		return appdto.CollectorDataSourceRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return out, nil
}

func (c *collectorHTTP) DeleteDataSource(ctx context.Context, companyID, sourceID, correlationID string) error {
	if c.baseURL == "" || sourceID == "" {
		return nil
	}
	resp, err := c.req(ctx, companyID, correlationID).Delete(c.dataSourceURL(sourceID))
	if err != nil {
		return MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	return nil
}

func (c *collectorHTTP) PropagateDataSource(ctx context.Context, companyID, sourceID, correlationID string, body appdto.PropagateDataSourceWriteRaw) (appdto.PropagateDataSourceRaw, error) {
	if c.baseURL == "" || sourceID == "" {
		return appdto.PropagateDataSourceRaw{Previews: []appdto.PropagateAgentPreviewRaw{}}, nil
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return appdto.PropagateDataSourceRaw{}, err
	}
	propagateClient := resty.New()
	propagateClient.SetTimeout(60 * time.Second)
	propagateClient.SetRetryCount(0)
	var out appdto.PropagateDataSourceRaw
	resp, err := propagateClient.R().
		SetContext(ctx).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&out).
		Post(c.dataSourceURL(sourceID) + "/propagate")
	if err != nil {
		return appdto.PropagateDataSourceRaw{}, MapNetworkError(err, "collector service")
	}
	if resp.StatusCode() != 200 {
		return appdto.PropagateDataSourceRaw{}, MapHTTPError(resp.StatusCode(), resp.Body(), "collector service")
	}
	if out.Previews == nil {
		out.Previews = []appdto.PropagateAgentPreviewRaw{}
	}
	return out, nil
}
