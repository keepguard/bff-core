package consent

import (
	"context"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.UserConsentClient
	cfg   observe.Config
}

func New(inner port.UserConsentClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.UserConsentClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) Accept(ctx context.Context, req appdto.UserConsentAcceptRequestDTO, token, tenantId, correlationID string) (appdto.UserConsentResponseDTO, error) {
	return observe.Call(d.cfg, "Accept", "POST", "/user-consents", correlationID, func() (appdto.UserConsentResponseDTO, error) {
		return d.inner.Accept(ctx, req, token, tenantId, correlationID)
	})
}

func (d *decorated) FindByID(ctx context.Context, id, token, tenantId, correlationID string) (appdto.UserConsentResponseDTO, error) {
	return observe.Call(d.cfg, "FindByID", "GET", "/user-consents/{id}", correlationID, func() (appdto.UserConsentResponseDTO, error) {
		return d.inner.FindByID(ctx, id, token, tenantId, correlationID)
	})
}

func (d *decorated) FindByUserID(ctx context.Context, userID, token, tenantId, correlationID string) ([]appdto.UserConsentResponseDTO, error) {
	return observe.Call(d.cfg, "FindByUserID", "GET", "/user-consents", correlationID, func() ([]appdto.UserConsentResponseDTO, error) {
		return d.inner.FindByUserID(ctx, userID, token, tenantId, correlationID)
	})
}

func (d *decorated) FindByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) ([]appdto.UserConsentResponseDTO, error) {
	return observe.Call(d.cfg, "FindByUserIDAndConsentDocumentID", "GET", "/user-consents", correlationID, func() ([]appdto.UserConsentResponseDTO, error) {
		return d.inner.FindByUserIDAndConsentDocumentID(ctx, userID, consentDocumentID, token, tenantId, correlationID)
	})
}

func (d *decorated) FindLatestByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) (appdto.UserConsentResponseDTO, error) {
	return observe.Call(d.cfg, "FindLatestByUserIDAndConsentDocumentID", "GET", "/user-consents/latest", correlationID, func() (appdto.UserConsentResponseDTO, error) {
		return d.inner.FindLatestByUserIDAndConsentDocumentID(ctx, userID, consentDocumentID, token, tenantId, correlationID)
	})
}

func (d *decorated) HasAccepted(ctx context.Context, userID, consentDocumentID string, version int, token, tenantId, correlationID string) (bool, error) {
	return observe.Call(d.cfg, "HasAccepted", "GET", "/user-consents/has-accepted", correlationID, func() (bool, error) {
		return d.inner.HasAccepted(ctx, userID, consentDocumentID, version, token, tenantId, correlationID)
	})
}

func (d *decorated) AcceptAll(ctx context.Context, req appdto.UserConsentAcceptAllRequestDTO, tenantId, correlationID string) (appdto.UserConsentAcceptAllResponseDTO, error) {
	return observe.Call(d.cfg, "AcceptAll", "POST", "/user-consents/accept-all", correlationID, func() (appdto.UserConsentAcceptAllResponseDTO, error) {
		return d.inner.AcceptAll(ctx, req, tenantId, correlationID)
	})
}

func (d *decorated) AcceptBatch(ctx context.Context, req appdto.UserConsentAcceptBatchRequestDTO, token, tenantId, correlationID string) (appdto.UserConsentAcceptAllResponseDTO, error) {
	return observe.Call(d.cfg, "AcceptBatch", "POST", "/user-consents/accept-batch", correlationID, func() (appdto.UserConsentAcceptAllResponseDTO, error) {
		return d.inner.AcceptBatch(ctx, req, token, tenantId, correlationID)
	})
}

func (d *decorated) DeleteAllByUserId(ctx context.Context, userID, tenantId, correlationID string) error {
	return observe.CallErr(d.cfg, "DeleteAllByUserId", "DELETE", "/user-consents", correlationID, func() error {
		return d.inner.DeleteAllByUserId(ctx, userID, tenantId, correlationID)
	})
}
