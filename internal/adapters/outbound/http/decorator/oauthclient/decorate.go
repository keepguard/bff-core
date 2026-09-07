package oauthclient

import (
	"context"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.OAuthClientClient
	cfg   observe.Config
}

func New(inner port.OAuthClientClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.OAuthClientClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) Search(ctx context.Context, companyID, bearerToken, correlationID string, query map[string]string) (appdto.PaginatedOAuthClients, error) {
	return observe.Call(d.cfg, "Search", "GET", "/oauth/clients", correlationID, func() (appdto.PaginatedOAuthClients, error) {
		return d.inner.Search(ctx, companyID, bearerToken, correlationID, query)
	})
}

func (d *decorated) GetByID(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error) {
	return observe.Call(d.cfg, "GetByID", "GET", "/oauth/clients/{id}", correlationID, func() (appdto.OAuthClientDTO, error) {
		return d.inner.GetByID(ctx, companyID, bearerToken, correlationID, id)
	})
}

func (d *decorated) Create(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.OAuthClientCreateRequest) (appdto.OAuthClientDTO, error) {
	return observe.Call(d.cfg, "Create", "POST", "/oauth/clients", correlationID, func() (appdto.OAuthClientDTO, error) {
		return d.inner.Create(ctx, companyID, bearerToken, correlationID, body)
	})
}

func (d *decorated) Update(ctx context.Context, companyID, bearerToken, correlationID, id string, body appdto.OAuthClientUpdateRequest) (appdto.OAuthClientDTO, error) {
	return observe.Call(d.cfg, "Update", "PUT", "/oauth/clients/{id}", correlationID, func() (appdto.OAuthClientDTO, error) {
		return d.inner.Update(ctx, companyID, bearerToken, correlationID, id, body)
	})
}

func (d *decorated) ListServiceRoles(ctx context.Context, companyID, bearerToken, correlationID string) ([]appdto.OAuthServiceRoleDTO, error) {
	return observe.Call(d.cfg, "ListServiceRoles", "GET", "/oauth/service-roles", correlationID, func() ([]appdto.OAuthServiceRoleDTO, error) {
		return d.inner.ListServiceRoles(ctx, companyID, bearerToken, correlationID)
	})
}

func (d *decorated) Block(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error) {
	return observe.Call(d.cfg, "Block", "POST", "/oauth/clients/{id}/block", correlationID, func() (appdto.OAuthClientDTO, error) {
		return d.inner.Block(ctx, companyID, bearerToken, correlationID, id)
	})
}

func (d *decorated) Unblock(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error) {
	return observe.Call(d.cfg, "Unblock", "POST", "/oauth/clients/{id}/unblock", correlationID, func() (appdto.OAuthClientDTO, error) {
		return d.inner.Unblock(ctx, companyID, bearerToken, correlationID, id)
	})
}

func (d *decorated) Delete(ctx context.Context, companyID, bearerToken, correlationID, id string) error {
	return observe.CallErr(d.cfg, "Delete", "DELETE", "/oauth/clients/{id}", correlationID, func() error {
		return d.inner.Delete(ctx, companyID, bearerToken, correlationID, id)
	})
}
