package client

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type KnowledgeClient interface {
	Ask(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.KnowledgeAskRequest) (appdto.KnowledgeAskResponse, error)
	GetSnapshot(ctx context.Context, companyID, bearerToken, correlationID, snapshotID string) (appdto.KnowledgeSnapshotDTO, error)
	GetDocumentPreview(ctx context.Context, companyID, bearerToken, correlationID, documentID string) (appdto.KnowledgeDocumentPreviewDTO, error)
	GetCollectionResults(ctx context.Context, companyID, bearerToken, correlationID, agentID, collectedAt string, windowSeconds int) (appdto.KnowledgeCollectionResultsDTO, error)
}
