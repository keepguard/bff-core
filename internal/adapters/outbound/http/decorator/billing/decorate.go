package billing

import (
	"context"
	"encoding/json"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.BillingClient
	cfg   observe.Config
}

func New(inner port.BillingClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.BillingClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) GetEntitlement(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	return observe.Call(d.cfg, "GetEntitlement", "GET", "/billing/entitlement", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.GetEntitlement(ctx, scope)
	})
}

func (d *decorated) GetCompanyEntitlement(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	return observe.Call(d.cfg, "GetCompanyEntitlement", "GET", "/billing/entitlement/company", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.GetCompanyEntitlement(ctx, scope)
	})
}

func (d *decorated) ListPlans(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	return observe.Call(d.cfg, "ListPlans", "GET", "/billing/plans", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.ListPlans(ctx, scope)
	})
}

func (d *decorated) SavePlan(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "SavePlan", "POST", "/billing/plans", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.SavePlan(ctx, scope, body)
	})
}

func (d *decorated) PatchPlan(ctx context.Context, scope port.BillingScope, code string, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "PatchPlan", "PATCH", "/billing/plans/{code}", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.PatchPlan(ctx, scope, code, body)
	})
}

func (d *decorated) GetGatewayAccount(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	return observe.Call(d.cfg, "GetGatewayAccount", "GET", "/billing/gateway-account", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.GetGatewayAccount(ctx, scope)
	})
}

func (d *decorated) PutGatewayAccount(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "PutGatewayAccount", "PUT", "/billing/gateway-account", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.PutGatewayAccount(ctx, scope, body)
	})
}

func (d *decorated) ListGatewayAccounts(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	return observe.Call(d.cfg, "ListGatewayAccounts", "GET", "/billing/gateway-accounts", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.ListGatewayAccounts(ctx, scope)
	})
}

func (d *decorated) PutGatewayAccountByGateway(ctx context.Context, scope port.BillingScope, gateway string, body any) (json.RawMessage, error) {
	return observe.Call(d.cfg, "PutGatewayAccountByGateway", "PUT", "/billing/gateway-accounts/{gateway}", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.PutGatewayAccountByGateway(ctx, scope, gateway, body)
	})
}

func (d *decorated) SetPrimaryGateway(ctx context.Context, scope port.BillingScope, gateway string) (json.RawMessage, error) {
	return observe.Call(d.cfg, "SetPrimaryGateway", "POST", "/billing/gateway-accounts/{gateway}/primary", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.SetPrimaryGateway(ctx, scope, gateway)
	})
}

func (d *decorated) GetSubscription(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	return observe.Call(d.cfg, "GetSubscription", "GET", "/billing/subscription", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.GetSubscription(ctx, scope)
	})
}

func (d *decorated) CreateSubscription(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, int, error) {
	type pair struct {
		raw    json.RawMessage
		status int
	}
	out, err := observe.Call(d.cfg, "CreateSubscription", "POST", "/billing/subscriptions", scope.CorrelationID, func() (pair, error) {
		raw, status, callErr := d.inner.CreateSubscription(ctx, scope, body)
		return pair{raw: raw, status: status}, callErr
	})
	if err != nil {
		return nil, 0, err
	}
	return out.raw, out.status, nil
}

func (d *decorated) CancelSubscription(ctx context.Context, scope port.BillingScope, id string) (json.RawMessage, error) {
	return observe.Call(d.cfg, "CancelSubscription", "POST", "/billing/subscriptions/{id}/cancel", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.CancelSubscription(ctx, scope, id)
	})
}

func (d *decorated) ListInvoices(ctx context.Context, scope port.BillingScope, query map[string]string) (json.RawMessage, error) {
	return observe.Call(d.cfg, "ListInvoices", "GET", "/billing/invoices", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.ListInvoices(ctx, scope, query)
	})
}

func (d *decorated) ListEntitlements(ctx context.Context, scope port.BillingScope, query map[string]string) (json.RawMessage, error) {
	return observe.Call(d.cfg, "ListEntitlements", "GET", "/billing/entitlements", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.ListEntitlements(ctx, scope, query)
	})
}

func (d *decorated) GetInvoice(ctx context.Context, scope port.BillingScope, id string) (json.RawMessage, error) {
	return observe.Call(d.cfg, "GetInvoice", "GET", "/billing/invoices/{id}", scope.CorrelationID, func() (json.RawMessage, error) {
		return d.inner.GetInvoice(ctx, scope, id)
	})
}

func (d *decorated) ForwardAsaasWebhook(ctx context.Context, accessToken string, body []byte) (json.RawMessage, int, error) {
	return d.inner.ForwardAsaasWebhook(ctx, accessToken, body)
}

func (d *decorated) ForwardStripeWebhook(ctx context.Context, token string, body []byte) (json.RawMessage, int, error) {
	return d.inner.ForwardStripeWebhook(ctx, token, body)
}
