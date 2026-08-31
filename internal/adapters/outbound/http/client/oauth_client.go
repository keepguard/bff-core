package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

type oauthClientHTTP struct {
	httpClient *resty.Client
	baseURL    string
	logger     *zap.Logger
}

func NewOAuthClientHTTP(cfg *config.Config, logger *zap.Logger) domainclient.OAuthClientClient {
	timeout := cfg.Services.Auth.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	httpClient := resty.New()
	httpClient.SetTimeout(timeout)
	httpClient.SetRetryCount(1)
	httpClient.SetRetryWaitTime(200 * time.Millisecond)
	return &oauthClientHTTP{
		httpClient: httpClient,
		baseURL:    cfg.Services.Auth.BaseURL,
		logger:     logger,
	}
}

func (c *oauthClientHTTP) Search(ctx context.Context, companyID, bearerToken, correlationID string, query map[string]string) (appdto.PaginatedOAuthClients, error) {
	var out appdto.PaginatedOAuthClients
	req := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+bearerToken).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetResult(&out)
	for key, value := range query {
		if value != "" {
			req.SetQueryParam(key, value)
		}
	}
	resp, err := req.Get(c.baseURL + "/api/v1/auth/oauth/clients")
	if err != nil {
		return appdto.PaginatedOAuthClients{}, MapNetworkError(err, "auth service")
	}
	if resp.StatusCode() != http.StatusOK {
		return appdto.PaginatedOAuthClients{}, MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}
	if out.Content == nil {
		out.Content = []appdto.OAuthClientDTO{}
	}
	return out, nil
}

func (c *oauthClientHTTP) GetByID(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error) {
	var out appdto.OAuthClientDTO
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+bearerToken).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetResult(&out).
		Get(fmt.Sprintf("%s/api/v1/auth/oauth/clients/%s", c.baseURL, id))
	if err != nil {
		return appdto.OAuthClientDTO{}, MapNetworkError(err, "auth service")
	}
	if resp.StatusCode() != http.StatusOK {
		return appdto.OAuthClientDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}
	return out, nil
}

func (c *oauthClientHTTP) Create(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.OAuthClientCreateRequest) (appdto.OAuthClientDTO, error) {
	var out appdto.OAuthClientDTO
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+bearerToken).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&out).
		Post(c.baseURL + "/api/v1/auth/oauth/clients")
	if err != nil {
		return appdto.OAuthClientDTO{}, MapNetworkError(err, "auth service")
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return appdto.OAuthClientDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}
	return out, nil
}

func (c *oauthClientHTTP) Block(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error) {
	return c.mutate(ctx, companyID, bearerToken, correlationID, id, "block")
}

func (c *oauthClientHTTP) Unblock(ctx context.Context, companyID, bearerToken, correlationID, id string) (appdto.OAuthClientDTO, error) {
	return c.mutate(ctx, companyID, bearerToken, correlationID, id, "unblock")
}

func (c *oauthClientHTTP) mutate(ctx context.Context, companyID, bearerToken, correlationID, id, action string) (appdto.OAuthClientDTO, error) {
	var out appdto.OAuthClientDTO
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+bearerToken).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetResult(&out).
		Post(fmt.Sprintf("%s/api/v1/auth/oauth/clients/%s/%s", c.baseURL, id, action))
	if err != nil {
		return appdto.OAuthClientDTO{}, MapNetworkError(err, "auth service")
	}
	if resp.StatusCode() != http.StatusOK {
		return appdto.OAuthClientDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}
	return out, nil
}

func (c *oauthClientHTTP) Delete(ctx context.Context, companyID, bearerToken, correlationID, id string) error {
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+bearerToken).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		Delete(fmt.Sprintf("%s/api/v1/auth/oauth/clients/%s", c.baseURL, id))
	if err != nil {
		return MapNetworkError(err, "auth service")
	}
	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusOK {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}
	return nil
}
