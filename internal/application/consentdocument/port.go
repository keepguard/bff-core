package consentdocument

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
)

type ConsentDocumentPort interface {
	ListPublished(ctx context.Context, query appdto.ListPublishedConsentsQuery) ([]appdto.ConsentDocumentViewDTO, error)
	GetLatestByType(ctx context.Context, query appdto.GetLatestConsentQuery) (appdto.ConsentDocumentViewDTO, error)
}

type service struct {
	documents port.ConsentDocumentClient
}

func NewConsentDocumentPort(documents port.ConsentDocumentClient) ConsentDocumentPort {
	if documents == nil {
		return nil
	}
	return &service{documents: documents}
}

func (s *service) ListPublished(ctx context.Context, query appdto.ListPublishedConsentsQuery) ([]appdto.ConsentDocumentViewDTO, error) {
	return s.documents.FindAllPublished(ctx, "", query.TenantID, query.CorrelationID)
}

func (s *service) GetLatestByType(ctx context.Context, query appdto.GetLatestConsentQuery) (appdto.ConsentDocumentViewDTO, error) {
	return s.documents.FindLatestPublishedByType(ctx, query.ConsentType, "", query.TenantID, query.CorrelationID)
}
