package client

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type OAuthClientClient interface {
	Search(ctx context.Context, companyID, bearerToken, correlationID string, query map[string]string) (appdto.PaginatedOAuthClients, error)
	GetByID(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Create(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.OAuthClientCreateRequest) (appdto.OAuthClientDTO, error)
	Update(ctx context.Context, companyID, bearerToken, correlationID, id string, body appdto.OAuthClientUpdateRequest) (appdto.OAuthClientDTO, error)
	ListServiceRoles(ctx context.Context, companyID, bearerToken, correlationID string) ([]appdto.OAuthServiceRoleDTO, error)
	Block(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Unblock(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error)
	Delete(ctx context.Context, companyID, bearerToken, correlationID, id string) error
}

type CollectorClient interface {
	ListAgents(ctx context.Context, companyID, correlationID string) ([]appdto.CollectorAgentRaw, error)
	SearchAgents(ctx context.Context, companyID, correlationID string, query map[string]string) (appdto.PaginatedCollectorAgentsRaw, error)
	GetAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error)
	CreateAgent(ctx context.Context, companyID, correlationID string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error)
	UpdateAgent(ctx context.Context, companyID, agentID, correlationID string, body appdto.CollectorAgentWriteRaw) (appdto.CollectorAgentRaw, error)
	EnableAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error)
	DisableAgent(ctx context.Context, companyID, agentID, correlationID string) (appdto.CollectorAgentRaw, error)
	DeleteAgent(ctx context.Context, companyID, agentID, correlationID string) error
}
