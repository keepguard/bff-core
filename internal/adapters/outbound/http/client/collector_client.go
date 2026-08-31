package client

import (
	"context"
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

func (c *collectorHTTP) ListAgents(ctx context.Context, companyID, correlationID string) ([]appdto.CollectorAgentRaw, error) {
	if c.baseURL == "" {
		return []appdto.CollectorAgentRaw{}, nil
	}
	var out []appdto.CollectorAgentRaw
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetResult(&out).
		Get(c.baseURL + "/api/v1/collector/agents")
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
