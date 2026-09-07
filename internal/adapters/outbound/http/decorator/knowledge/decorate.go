package knowledge

import (
	"context"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.KnowledgeClient
	cfg   observe.Config
}

func New(inner port.KnowledgeClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.KnowledgeClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) Ask(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.KnowledgeAskRequest) (appdto.KnowledgeAskResponse, error) {
	return observe.Call(d.cfg, "Ask", "POST", "/ask", correlationID, func() (appdto.KnowledgeAskResponse, error) {
		return d.inner.Ask(ctx, companyID, bearerToken, correlationID, body)
	})
}

func (d *decorated) GetSnapshot(ctx context.Context, companyID, bearerToken, correlationID, snapshotID string) (appdto.KnowledgeSnapshotDTO, error) {
	return observe.Call(d.cfg, "GetSnapshot", "GET", "/snapshots/{id}", correlationID, func() (appdto.KnowledgeSnapshotDTO, error) {
		return d.inner.GetSnapshot(ctx, companyID, bearerToken, correlationID, snapshotID)
	})
}

func (d *decorated) GetDocumentPreview(ctx context.Context, companyID, bearerToken, correlationID, documentID string) (appdto.KnowledgeDocumentPreviewDTO, error) {
	return observe.Call(d.cfg, "GetDocumentPreview", "GET", "/documents/{id}", correlationID, func() (appdto.KnowledgeDocumentPreviewDTO, error) {
		return d.inner.GetDocumentPreview(ctx, companyID, bearerToken, correlationID, documentID)
	})
}

func (d *decorated) GetCollectionResults(ctx context.Context, companyID, bearerToken, correlationID, agentID, collectedAt string, windowSeconds int) (appdto.KnowledgeCollectionResultsDTO, error) {
	return observe.Call(d.cfg, "GetCollectionResults", "GET", "/collection-results", correlationID, func() (appdto.KnowledgeCollectionResultsDTO, error) {
		return d.inner.GetCollectionResults(ctx, companyID, bearerToken, correlationID, agentID, collectedAt, windowSeconds)
	})
}
