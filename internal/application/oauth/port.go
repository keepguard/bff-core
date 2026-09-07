package oauth

import (
	"context"
	"net/http"
	"strings"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/application/scope"
	"github.com/keepguard/bff-core/internal/pkg"
	"go.uber.org/zap"
)

type CompanyScope struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	BearerToken    string
}

type ListClientsQuery struct {
	CompanyScope
	Filters map[string]string
}

type GetClientQuery struct {
	CompanyScope
	ID string
}

type CreateClientCommand struct {
	CompanyScope
	Body appdto.OAuthClientCreateRequest
}

type UpdateClientCommand struct {
	CompanyScope
	ID   string
	Body appdto.OAuthClientUpdateRequest
}

type ClientIDCommand struct {
	CompanyScope
	ID string
}

type OAuthPort interface {
	Search(ctx context.Context, query ListClientsQuery) (appdto.PaginatedOAuthClients, error)
	GetByID(ctx context.Context, query GetClientQuery) (appdto.OAuthClientDTO, error)
	Create(ctx context.Context, cmd CreateClientCommand) (appdto.OAuthClientDTO, error)
	Update(ctx context.Context, cmd UpdateClientCommand) (appdto.OAuthClientDTO, error)
	ListServiceRoles(ctx context.Context, query CompanyScope) ([]appdto.OAuthServiceRoleDTO, error)
	Block(ctx context.Context, cmd ClientIDCommand) (appdto.OAuthClientDTO, error)
	Unblock(ctx context.Context, cmd ClientIDCommand) (appdto.OAuthClientDTO, error)
	Delete(ctx context.Context, cmd ClientIDCommand) error
}

type service struct {
	oauth     port.OAuthClientClient
	companies port.CompanyClient
	collector port.CollectorClient
	logger    *zap.Logger
}

func NewOAuthPort(
	oauth port.OAuthClientClient,
	companies port.CompanyClient,
	collector port.CollectorClient,
	logger *zap.Logger,
) OAuthPort {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &service{oauth: oauth, companies: companies, collector: collector, logger: logger}
}

func (s *service) requireCompany(ctx context.Context, scopeIn CompanyScope) (string, error) {
	if s.oauth == nil || s.companies == nil {
		return "", scope.Unavailable("Gestão de OAuth clients indisponível")
	}
	return scope.ResolveCompany(ctx, s.companies, scopeIn.CompanyFromCtx, scopeIn.TenantID, scopeIn.CorrelationID)
}

func (s *service) Search(ctx context.Context, query ListClientsQuery) (appdto.PaginatedOAuthClients, error) {
	companyID, err := s.requireCompany(ctx, query.CompanyScope)
	if err != nil {
		return appdto.PaginatedOAuthClients{}, err
	}
	return s.oauth.Search(ctx, companyID, query.BearerToken, query.CorrelationID, query.Filters)
}

func (s *service) GetByID(ctx context.Context, query GetClientQuery) (appdto.OAuthClientDTO, error) {
	companyID, err := s.requireCompany(ctx, query.CompanyScope)
	if err != nil {
		return appdto.OAuthClientDTO{}, err
	}
	return s.oauth.GetByID(ctx, companyID, query.BearerToken, query.CorrelationID, query.ID)
}

func (s *service) Create(ctx context.Context, cmd CreateClientCommand) (appdto.OAuthClientDTO, error) {
	companyID, err := s.requireCompany(ctx, cmd.CompanyScope)
	if err != nil {
		return appdto.OAuthClientDTO{}, err
	}
	if strings.TrimSpace(cmd.Body.ClientID) == "" {
		return appdto.OAuthClientDTO{}, pkg.NewAppError("MISSING_CLIENT_ID", "clientId é obrigatório", http.StatusBadRequest)
	}
	if strings.TrimSpace(cmd.Body.RoleID) == "" {
		return appdto.OAuthClientDTO{}, pkg.NewAppError("MISSING_ROLE_ID", "roleId é obrigatório", http.StatusBadRequest)
	}
	return s.oauth.Create(ctx, companyID, cmd.BearerToken, cmd.CorrelationID, cmd.Body)
}

func (s *service) Update(ctx context.Context, cmd UpdateClientCommand) (appdto.OAuthClientDTO, error) {
	companyID, err := s.requireCompany(ctx, cmd.CompanyScope)
	if err != nil {
		return appdto.OAuthClientDTO{}, err
	}
	if strings.TrimSpace(cmd.Body.RoleID) == "" {
		return appdto.OAuthClientDTO{}, pkg.NewAppError("MISSING_ROLE_ID", "roleId é obrigatório", http.StatusBadRequest)
	}
	return s.oauth.Update(ctx, companyID, cmd.BearerToken, cmd.CorrelationID, cmd.ID, cmd.Body)
}

func (s *service) ListServiceRoles(ctx context.Context, query CompanyScope) ([]appdto.OAuthServiceRoleDTO, error) {
	companyID, err := s.requireCompany(ctx, query)
	if err != nil {
		return nil, err
	}
	roles, err := s.oauth.ListServiceRoles(ctx, companyID, query.BearerToken, query.CorrelationID)
	if err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []appdto.OAuthServiceRoleDTO{}
	}
	return roles, nil
}

func (s *service) Block(ctx context.Context, cmd ClientIDCommand) (appdto.OAuthClientDTO, error) {
	companyID, err := s.requireCompany(ctx, cmd.CompanyScope)
	if err != nil {
		return appdto.OAuthClientDTO{}, err
	}
	result, err := s.oauth.Block(ctx, companyID, cmd.BearerToken, cmd.CorrelationID, cmd.ID)
	if err != nil {
		return appdto.OAuthClientDTO{}, err
	}
	s.disableCompanyAgents(ctx, companyID, cmd.CorrelationID)
	return result, nil
}

func (s *service) Unblock(ctx context.Context, cmd ClientIDCommand) (appdto.OAuthClientDTO, error) {
	companyID, err := s.requireCompany(ctx, cmd.CompanyScope)
	if err != nil {
		return appdto.OAuthClientDTO{}, err
	}
	return s.oauth.Unblock(ctx, companyID, cmd.BearerToken, cmd.CorrelationID, cmd.ID)
}

func (s *service) Delete(ctx context.Context, cmd ClientIDCommand) error {
	companyID, err := s.requireCompany(ctx, cmd.CompanyScope)
	if err != nil {
		return err
	}
	if err := s.oauth.Delete(ctx, companyID, cmd.BearerToken, cmd.CorrelationID, cmd.ID); err != nil {
		return err
	}
	s.disableCompanyAgents(ctx, companyID, cmd.CorrelationID)
	return nil
}

func (s *service) disableCompanyAgents(ctx context.Context, companyID, correlationID string) {
	if s.collector == nil {
		return
	}
	raw, err := s.collector.ListAgents(ctx, companyID, correlationID)
	if err != nil {
		s.logger.Warn("Collector indisponível ao desabilitar agents",
			zap.String("correlationId", correlationID),
			zap.String("companyId", companyID),
			zap.Error(err),
		)
		return
	}
	for _, agent := range raw {
		if !agent.Enabled || strings.TrimSpace(agent.ID) == "" {
			continue
		}
		if _, err := s.collector.DisableAgent(ctx, companyID, agent.ID, correlationID); err != nil {
			s.logger.Warn("Falha ao desabilitar agent após alteração do OAuth client",
				zap.String("correlationId", correlationID),
				zap.String("companyId", companyID),
				zap.String("agentId", agent.ID),
				zap.Error(err),
			)
		}
	}
}
