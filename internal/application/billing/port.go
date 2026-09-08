package billing

import (
	"context"
	"encoding/json"

	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/application/scope"
)

type BillingPort interface {
	GetEntitlement(ctx context.Context, scope port.BillingScope) (json.RawMessage, error)
	ListPlans(ctx context.Context, scope port.BillingScope) (json.RawMessage, error)
	SavePlan(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, error)
	PatchPlan(ctx context.Context, scope port.BillingScope, code string, body any) (json.RawMessage, error)
	GetGatewayAccount(ctx context.Context, scope port.BillingScope) (json.RawMessage, error)
	PutGatewayAccount(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, error)
	GetSubscription(ctx context.Context, scope port.BillingScope) (json.RawMessage, error)
	CreateSubscription(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, int, error)
	CancelSubscription(ctx context.Context, scope port.BillingScope, id string) (json.RawMessage, error)
	ListInvoices(ctx context.Context, scope port.BillingScope) (json.RawMessage, error)
	GetInvoice(ctx context.Context, scope port.BillingScope, id string) (json.RawMessage, error)
	ForwardAsaasWebhook(ctx context.Context, accessToken string, body []byte) (json.RawMessage, int, error)
}

type service struct {
	client port.BillingClient
}

func NewBillingPort(client port.BillingClient) BillingPort {
	if client == nil {
		return nil
	}
	return &service{client: client}
}

func (s *service) require() error {
	if s == nil || s.client == nil {
		return scope.Unavailable("Billing indisponível")
	}
	return nil
}

func (s *service) GetEntitlement(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.GetEntitlement(ctx, scope)
}

func (s *service) ListPlans(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.ListPlans(ctx, scope)
}

func (s *service) SavePlan(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.SavePlan(ctx, scope, body)
}

func (s *service) PatchPlan(ctx context.Context, scope port.BillingScope, code string, body any) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.PatchPlan(ctx, scope, code, body)
}

func (s *service) GetGatewayAccount(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.GetGatewayAccount(ctx, scope)
}

func (s *service) PutGatewayAccount(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.PutGatewayAccount(ctx, scope, body)
}

func (s *service) GetSubscription(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.GetSubscription(ctx, scope)
}

func (s *service) CreateSubscription(ctx context.Context, scope port.BillingScope, body any) (json.RawMessage, int, error) {
	if err := s.require(); err != nil {
		return nil, 0, err
	}
	return s.client.CreateSubscription(ctx, scope, body)
}

func (s *service) CancelSubscription(ctx context.Context, scope port.BillingScope, id string) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.CancelSubscription(ctx, scope, id)
}

func (s *service) ListInvoices(ctx context.Context, scope port.BillingScope) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.ListInvoices(ctx, scope)
}

func (s *service) GetInvoice(ctx context.Context, scope port.BillingScope, id string) (json.RawMessage, error) {
	if err := s.require(); err != nil {
		return nil, err
	}
	return s.client.GetInvoice(ctx, scope, id)
}

func (s *service) ForwardAsaasWebhook(ctx context.Context, accessToken string, body []byte) (json.RawMessage, int, error) {
	if err := s.require(); err != nil {
		return nil, 0, err
	}
	return s.client.ForwardAsaasWebhook(ctx, accessToken, body)
}
