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

type guardianClient struct {
	httpClient *resty.Client
	baseURL    string
	logger     *zap.Logger
}

func NewGuardianClient(cfg *config.Config, logger *zap.Logger) domainclient.GuardianClient {
	timeout := cfg.Services.Guardian.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	httpClient := resty.New()
	httpClient.SetTimeout(timeout)
	httpClient.SetRetryCount(1)
	httpClient.SetRetryWaitTime(200 * time.Millisecond)
	return &guardianClient{
		httpClient: httpClient,
		baseURL:    cfg.Services.Guardian.BaseURL,
		logger:     logger,
	}
}

func (c *guardianClient) ListIncidents(ctx context.Context, tenantID, correlationID string, query map[string]string) (appdto.PaginatedGuardianIncidents, error) {
	var out appdto.PaginatedGuardianIncidents
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
	resp, err := req.Get(c.baseURL + "/api/v1/guardian/incidents")
	if err != nil {
		return appdto.PaginatedGuardianIncidents{}, MapNetworkError(err, "guardian service")
	}
	if resp.StatusCode() != 200 {
		return appdto.PaginatedGuardianIncidents{}, MapHTTPError(resp.StatusCode(), resp.Body(), "guardian service")
	}
	return out, nil
}

func (c *guardianClient) GetIncident(ctx context.Context, tenantID, correlationID, id string) (map[string]any, error) {
	var out map[string]any
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Tenant-Id", tenantID).
		SetHeader("X-Correlation-ID", correlationID).
		SetResult(&out).
		Get(fmt.Sprintf("%s/api/v1/guardian/incidents/%s", c.baseURL, id))
	if err != nil {
		return nil, MapNetworkError(err, "guardian service")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "guardian service")
	}
	return out, nil
}

func (c *guardianClient) ExecuteAction(ctx context.Context, tenantID, correlationID, userID, userEmail, userRole, id string, body appdto.GuardianExecuteActionRequest) (map[string]any, error) {
	var out map[string]any
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Tenant-Id", tenantID).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-User-ID", userID).
		SetHeader("X-User-Email", userEmail).
		SetHeader("X-User-Role", userRole).
		SetBody(body).
		SetResult(&out).
		Post(fmt.Sprintf("%s/api/v1/guardian/incidents/%s/actions", c.baseURL, id))
	if err != nil {
		return nil, MapNetworkError(err, "guardian service")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "guardian service")
	}
	return out, nil
}

func (c *guardianClient) ListRecipients(ctx context.Context, tenantID, correlationID string) ([]map[string]any, error) {
	var out []map[string]any
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Tenant-Id", tenantID).
		SetHeader("X-Correlation-ID", correlationID).
		SetResult(&out).
		Get(c.baseURL + "/api/v1/guardian/alert-recipients")
	if err != nil {
		return nil, MapNetworkError(err, "guardian service")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "guardian service")
	}
	return out, nil
}

func (c *guardianClient) UpsertRecipient(ctx context.Context, tenantID, correlationID string, body appdto.GuardianRecipientUpsertRequest) (map[string]any, error) {
	var out map[string]any
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Tenant-Id", tenantID).
		SetHeader("X-Correlation-ID", correlationID).
		SetBody(body).
		SetResult(&out).
		Put(c.baseURL + "/api/v1/guardian/alert-recipients")
	if err != nil {
		return nil, MapNetworkError(err, "guardian service")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "guardian service")
	}
	return out, nil
}

func (c *guardianClient) PatchRecipient(ctx context.Context, tenantID, correlationID, id string, body appdto.GuardianRecipientUpsertRequest) (map[string]any, error) {
	var out map[string]any
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Tenant-Id", tenantID).
		SetHeader("X-Correlation-ID", correlationID).
		SetBody(body).
		SetResult(&out).
		Patch(fmt.Sprintf("%s/api/v1/guardian/alert-recipients/%s", c.baseURL, id))
	if err != nil {
		return nil, MapNetworkError(err, "guardian service")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "guardian service")
	}
	return out, nil
}
