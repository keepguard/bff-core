package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

type auditClient struct {
	httpClient *resty.Client
	baseURL    string
	logger     *zap.Logger
}

func NewAuditClient(cfg *config.Config, logger *zap.Logger) domainclient.AuditClient {
	timeout := cfg.Services.Audit.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	httpClient := resty.New()
	httpClient.SetTimeout(timeout)
	httpClient.SetRetryCount(1)
	httpClient.SetRetryWaitTime(200 * time.Millisecond)
	return &auditClient{
		httpClient: httpClient,
		baseURL:    cfg.Services.Audit.BaseURL,
		logger:     logger,
	}
}

func (c *auditClient) List(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedAuditResponse, error) {
	var out appdto.PaginatedAuditResponse
	req := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Tenant-Id", tenantID).
		SetHeader("X-Correlation-ID", correlationID).
		SetResult(&out)
	for key, value := range query {
		if value != "" {
			req.SetQueryParam(key, value)
		}
	}
	resp, err := req.Get(c.baseURL + "/api/v1/audits")
	if err != nil {
		return appdto.PaginatedAuditResponse{}, MapNetworkError(err, "audit service")
	}
	if resp.StatusCode() != 200 {
		return appdto.PaginatedAuditResponse{}, MapHTTPError(resp.StatusCode(), resp.Body(), "audit service")
	}
	return out, nil
}

func (c *auditClient) GetByID(ctx context.Context, tenantID, correlationID, eventID string) (appdto.AuditDetailResponse, error) {
	var out appdto.AuditDetailResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Tenant-Id", tenantID).
		SetHeader("X-Correlation-ID", correlationID).
		SetResult(&out).
		Get(fmt.Sprintf("%s/api/v1/audits/%s", c.baseURL, eventID))
	if err != nil {
		return appdto.AuditDetailResponse{}, MapNetworkError(err, "audit service")
	}
	if resp.StatusCode() != 200 {
		return appdto.AuditDetailResponse{}, MapHTTPError(resp.StatusCode(), resp.Body(), "audit service")
	}
	return out, nil
}
