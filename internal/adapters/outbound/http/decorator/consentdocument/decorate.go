package consentdocument

import (
	"context"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.ConsentDocumentClient
	cfg   observe.Config
}

func New(inner port.ConsentDocumentClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.ConsentDocumentClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) FindLatestPublishedByType(ctx context.Context, consentType, token, tenantId, correlationID string) (appdto.ConsentDocumentResponseDTO, error) {
	return observe.Call(d.cfg, "FindLatestPublishedByType", "GET", "/consents/latest", correlationID, func() (appdto.ConsentDocumentResponseDTO, error) {
		return d.inner.FindLatestPublishedByType(ctx, consentType, token, tenantId, correlationID)
	})
}

func (d *decorated) FindAllPublished(ctx context.Context, token, tenantId, correlationID string) ([]appdto.ConsentDocumentResponseDTO, error) {
	return observe.Call(d.cfg, "FindAllPublished", "GET", "/consents/published", correlationID, func() ([]appdto.ConsentDocumentResponseDTO, error) {
		return d.inner.FindAllPublished(ctx, token, tenantId, correlationID)
	})
}
