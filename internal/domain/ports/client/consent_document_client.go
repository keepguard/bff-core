package client

import (
	"context"

	consentDocumentDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/consent_document"
)

// ConsentDocumentClient interface para comunicação com o serviço de documentos de consentimento
type ConsentDocumentClient interface {
	// FindLatestPublishedByType busca a última versão publicada por tipo
	FindLatestPublishedByType(ctx context.Context, consentType, token, xApplication, correlationID string) (consentDocumentDto.ConsentDocumentResponseDTO, error)
}
