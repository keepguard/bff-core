package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	authDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/auth"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	client "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// authClient implementa AuthClient usando HTTP
type authClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
	// tokens é o token OAuth de serviço do bff-core por company (client_id bff-core, o mesmo de knowledge, collector
	// e LLM). Só o hard delete usa. Nil = sem token (o ms-auth em account_hard_delete.mode=observe aceita e registra).
	tokens client.ServiceTokenClient
}

// NewAuthClient cria uma nova instância do AuthClient. tokens pode ser nil.
func NewAuthClient(config *config.Config, tokens client.ServiceTokenClient, logger *zap.Logger) client.AuthClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(2)
	httpClient.SetRetryWaitTime(500 * time.Millisecond)

	return &authClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
		tokens:     tokens,
	}
}

// NewAuthClientWithoutRetry cria uma nova instância do AuthClient SEM retry automático. tokens pode ser nil.
func NewAuthClientWithoutRetry(config *config.Config, tokens client.ServiceTokenClient, logger *zap.Logger) client.AuthClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	// SEM SetRetryCount - sem retry no Resty

	return &authClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
		tokens:     tokens,
	}
}

// ValidateToken valida um token no ms-auth
func (c *authClient) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	url := fmt.Sprintf("%s/api/v1/auth/validate", c.config.Services.Auth.BaseURL)

	req := map[string]string{
		"token": token,
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyHeader(ctx)).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return fmt.Errorf("erro ao comunicar com auth service: %w", err)
	}

	if resp.StatusCode() != 200 {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}

	return nil
}

// GenerateResetToken gera um token de recuperação de senha no ms-auth
func (c *authClient) GenerateResetToken(ctx context.Context, req authDto.GenerateResetTokenMSRequestDTO, tenantId, correlationID string) (authDto.GenerateResetTokenMSResponseDTO, error) {
	var response authDto.GenerateResetTokenMSResponseDTO

	url := fmt.Sprintf("%s/api/v1/auth/generate-reset-token", c.config.Services.Auth.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyHeader(ctx)).
		Post(url)

	if err != nil {
		return authDto.GenerateResetTokenMSResponseDTO{}, fmt.Errorf("erro ao fazer requisição de geração de token de reset: %w", err)
	}

	if resp.StatusCode() != 200 {
		return authDto.GenerateResetTokenMSResponseDTO{}, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "auth service retornou erro",
			Details: string(resp.Body()),
		}
	}

	return response, nil
}

// RegisterLogin realiza login após registro usando senha criptografada
func (c *authClient) RegisterLogin(ctx context.Context, req authDto.AuthRegisterLoginRequestDTO, tenantId, correlationID, clientId string) (authDto.AuthLoginResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/auth/register-login", c.config.Services.Auth.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyHeader(ctx)).
		SetHeader("X-Device-Id", client.DeviceIDFromContext(ctx)).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return authDto.AuthLoginResponseDTO{}, fmt.Errorf("erro ao comunicar com auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return authDto.AuthLoginResponseDTO{}, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "auth service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var loginResponse authDto.AuthLoginResponseDTO
	if err := json.Unmarshal(resp.Body(), &loginResponse); err != nil {
		return authDto.AuthLoginResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return loginResponse, nil
}

// CreateUser cria um novo usuário no ms-auth
func (c *authClient) CreateUser(ctx context.Context, req authDto.AuthUserCreateRequestDTO, tenantId, correlationID string) (authDto.AuthUserCreateResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/users/create", c.config.Services.Auth.BaseURL)

	// Converte diretamente para o DTO do ms-auth
	authReq := authDto.AuthUserCreateRequestDTO{
		Username:       req.Username,
		Email:          req.Email,
		Password:       req.Password,
		IDUserExternal: req.IDUserExternal,
		CodeUser:       req.CodeUser,
		CompanyID:      req.CompanyID,
		CompanyCode:    req.CompanyCode,
		TenantId:       req.TenantId,
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(authReq).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyHeader(ctx)).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return authDto.AuthUserCreateResponseDTO{}, fmt.Errorf("erro ao comunicar com auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated {
		return authDto.AuthUserCreateResponseDTO{}, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "auth service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var userResponse authDto.AuthUserCreateResponseDTO
	if err := json.Unmarshal(resp.Body(), &userResponse); err != nil {
		return authDto.AuthUserCreateResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return userResponse, nil
}

// HardDeleteUser remove permanentemente um usuário (para compensação de SAGA). Vai com o token de serviço do
// bff-core da company do cadastro (claim company_id = X-Company-Id, authority user:service-hard-delete): o ms-auth
// em account_hard_delete.mode=enforce recusa sem ele. Sem token (sem provedor ou falha ao obter), a chamada segue
// sem Authorization, como antes: em observe o ms-auth aceita e registra hard_delete_sem_token; a compensação
// não pode parar por isso.
func (c *authClient) HardDeleteUser(ctx context.Context, idUserExternal, tenantId, correlationID string) error {
	url := fmt.Sprintf("%s/api/v1/users/hard-delete/%s", c.config.Services.Auth.BaseURL, idUserExternal)
	companyID := companyHeader(ctx)

	req := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyID).
		SetHeader("Content-Type", "application/json")
	if bearer := c.serviceBearer(ctx, companyID); bearer != "" {
		req.SetHeader("Authorization", bearer)
	}
	resp, err := req.Delete(url)

	if err != nil {
		return fmt.Errorf("erro ao comunicar com auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}

	return nil
}

// serviceBearer devolve "Bearer <token>" do bff-core para a company, ou "" (sem provedor, sem company ou falha).
// Nunca loga o token.
func (c *authClient) serviceBearer(ctx context.Context, companyID string) string {
	if c.tokens == nil || strings.TrimSpace(companyID) == "" {
		return ""
	}
	tok, err := c.tokens.GetToken(ctx, companyID)
	if err != nil || strings.TrimSpace(tok) == "" {
		if c.logger != nil {
			c.logger.Warn("hard delete sem token de serviço do bff-core", zap.String("company_id", companyID), zap.Error(err))
		}
		return ""
	}
	if !strings.HasPrefix(strings.ToLower(tok), "bearer ") {
		tok = "Bearer " + tok
	}
	return tok
}
