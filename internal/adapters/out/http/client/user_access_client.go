package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	authDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/auth"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
)

const userAccessTimeout = 3 * time.Second

type userAccessClient struct {
	httpClient *resty.Client
	baseURL    string
}

// NewUserAccessClient consulta os papéis na fonte, no ms-auth. Sem retry nem cache: a
// resposta decide uma escrita de dinheiro, e quem não consegue perguntar recusa.
func NewUserAccessClient(cfg *config.Config) port.UserAccessClient {
	httpClient := resty.New()
	httpClient.SetTimeout(userAccessTimeout)
	return &userAccessClient{httpClient: httpClient, baseURL: cfg.Services.Auth.BaseURL}
}

func (c *userAccessClient) GetUserAccess(ctx context.Context, codeUser, companyID, bearerToken, correlationID string) (port.UserAccess, error) {
	resp, err := c.httpClient.R().SetContext(ctx).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("Authorization", "Bearer "+strings.TrimPrefix(bearerToken, "Bearer ")).
		Get(c.baseURL + "/api/v1/users/code-user/" + url.PathEscape(codeUser) + "/access")
	if err != nil {
		return port.UserAccess{}, MapNetworkError(err, "ms-auth")
	}
	switch resp.StatusCode() {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return port.UserAccess{}, port.ErrUserAccessRefused
	default:
		return port.UserAccess{}, fmt.Errorf("ms-auth respondeu %d ao ler os papéis", resp.StatusCode())
	}
	var body authDto.UserAccessMSResponseDTO
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return port.UserAccess{}, fmt.Errorf("resposta inválida do ms-auth ao ler os papéis")
	}
	return port.UserAccess{Active: body.Active, Roles: body.Roles, Authorities: body.Authorities}, nil
}
