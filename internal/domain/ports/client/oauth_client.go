package client

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type OAuthClientClient interface {
	Search(ctx context.Context, companyID, bearerToken, correlationID string, query map[string]string) (appdto.PaginatedOAuthClients, error)
	GetByID(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Create(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.OAuthClientCreateRequest) (appdto.OAuthClientDTO, error)
	Block(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Unblock(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Delete(ctx context.Context, companyID, bearerToken, correlationID, id string) error
}

type CollectorClient interface {
	ListAgents(ctx context.Context, companyID, correlationID string) ([]appdto.CollectorAgentRaw, error)
}
