package client

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type KnowledgeClient interface {
	Ask(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.KnowledgeAskRequest) (appdto.KnowledgeAskResponse, error)
}
