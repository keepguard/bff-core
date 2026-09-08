package client

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	domainclient "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

type billingClient struct {
	httpClient *resty.Client
	baseURL    string
	logger     *zap.Logger
}

func NewBillingClient(cfg *config.Config, logger *zap.Logger) domainclient.BillingClient {
	timeout := cfg.Services.Billing.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	httpClient := resty.New()
	httpClient.SetTimeout(timeout)
	httpClient.SetRetryCount(0)
	return &billingClient{httpClient: httpClient, baseURL: cfg.Services.Billing.BaseURL, logger: logger}
}

func (c *billingClient) headers(ctx context.Context, scope domainclient.BillingScope) *resty.Request {
	req := c.httpClient.R().SetContext(ctx).
		SetHeader("X-Company-Id", scope.CompanyID).
		SetHeader("X-User-Id", scope.UserID).
		SetHeader("X-Caller-Admin", strconv.FormatBool(scope.Admin)).
		SetHeader("X-Correlation-ID", scope.CorrelationID)
	if token := strings.TrimSpace(domainclient.BearerTokenFromContext(ctx)); token != "" {
		req.SetHeader("Authorization", "Bearer "+strings.TrimPrefix(token, "Bearer "))
	}
	return req
}

func (c *billingClient) GetEntitlement(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return c.get(ctx, scope, "/api/v1/billing/entitlement")
}

func (c *billingClient) ListPlans(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return c.get(ctx, scope, "/api/v1/billing/plans")
}

func (c *billingClient) SavePlan(ctx context.Context, scope domainclient.BillingScope, body any) (json.RawMessage, error) {
	return c.send(ctx, scope, "POST", "/api/v1/billing/plans", body, 201)
}

func (c *billingClient) PatchPlan(ctx context.Context, scope domainclient.BillingScope, code string, body any) (json.RawMessage, error) {
	return c.send(ctx, scope, "PATCH", "/api/v1/billing/plans/"+code, body, 200)
}

func (c *billingClient) GetGatewayAccount(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return c.get(ctx, scope, "/api/v1/billing/gateway-account")
}

func (c *billingClient) PutGatewayAccount(ctx context.Context, scope domainclient.BillingScope, body any) (json.RawMessage, error) {
	return c.send(ctx, scope, "PUT", "/api/v1/billing/gateway-account", body, 200)
}

func (c *billingClient) GetSubscription(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return c.get(ctx, scope, "/api/v1/billing/subscription")
}

func (c *billingClient) CreateSubscription(ctx context.Context, scope domainclient.BillingScope, body any) (json.RawMessage, int, error) {
	resp, err := c.headers(ctx, scope).SetBody(body).Post(c.baseURL + "/api/v1/billing/subscriptions")
	if err != nil {
		return nil, 0, MapNetworkError(err, "ms-billing")
	}
	if resp.StatusCode() != 201 && resp.StatusCode() != 202 {
		return nil, resp.StatusCode(), MapHTTPError(resp.StatusCode(), resp.Body(), "ms-billing")
	}
	return cloneBody(resp.Body()), resp.StatusCode(), nil
}

func (c *billingClient) CancelSubscription(ctx context.Context, scope domainclient.BillingScope, id string) (json.RawMessage, error) {
	return c.send(ctx, scope, "POST", "/api/v1/billing/subscriptions/"+id+"/cancel", nil, 200)
}

func (c *billingClient) ListInvoices(ctx context.Context, scope domainclient.BillingScope) (json.RawMessage, error) {
	return c.get(ctx, scope, "/api/v1/billing/invoices")
}

func (c *billingClient) GetInvoice(ctx context.Context, scope domainclient.BillingScope, id string) (json.RawMessage, error) {
	return c.get(ctx, scope, "/api/v1/billing/invoices/"+id)
}

func (c *billingClient) ForwardAsaasWebhook(ctx context.Context, accessToken string, body []byte) (json.RawMessage, int, error) {
	resp, err := c.httpClient.R().SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("asaas-access-token", accessToken).
		SetBody(body).
		Post(c.baseURL + "/api/v1/billing/webhooks/asaas")
	if err != nil {
		return nil, 0, MapNetworkError(err, "ms-billing")
	}
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, resp.StatusCode(), MapHTTPError(resp.StatusCode(), resp.Body(), "ms-billing")
	}
	return cloneBody(resp.Body()), resp.StatusCode(), nil
}

func (c *billingClient) get(ctx context.Context, scope domainclient.BillingScope, path string) (json.RawMessage, error) {
	resp, err := c.headers(ctx, scope).Get(c.baseURL + path)
	if err != nil {
		return nil, MapNetworkError(err, "ms-billing")
	}
	if resp.StatusCode() != 200 {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "ms-billing")
	}
	return cloneBody(resp.Body()), nil
}

func (c *billingClient) send(ctx context.Context, scope domainclient.BillingScope, method, path string, body any, want int) (json.RawMessage, error) {
	req := c.headers(ctx, scope)
	if body != nil {
		req.SetBody(body)
	}
	var resp *resty.Response
	var err error
	url := c.baseURL + path
	switch method {
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "PATCH":
		resp, err = req.Patch(url)
	default:
		resp, err = req.Get(url)
	}
	if err != nil {
		return nil, MapNetworkError(err, "ms-billing")
	}
	if resp.StatusCode() != want {
		return nil, MapHTTPError(resp.StatusCode(), resp.Body(), "ms-billing")
	}
	return cloneBody(resp.Body()), nil
}
