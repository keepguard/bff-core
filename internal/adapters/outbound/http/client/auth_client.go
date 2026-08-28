package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// authClient implementa AuthClient usando HTTP
type authClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewAuthClient cria uma nova instância do AuthClient
func NewAuthClient(config *config.Config, logger *zap.Logger) client.AuthClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(2)
	httpClient.SetRetryWaitTime(500 * time.Millisecond)

	return &authClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// NewAuthClientWithoutRetry cria uma nova instância do AuthClient SEM retry automático
func NewAuthClientWithoutRetry(config *config.Config, logger *zap.Logger) client.AuthClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	// SEM SetRetryCount - sem retry no Resty

	return &authClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
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
		TenantId:   req.TenantId,
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

// HardDeleteUser remove permanentemente um usuário (para compensação de SAGA)
func (c *authClient) HardDeleteUser(ctx context.Context, idUserExternal, tenantId, correlationID string) error {
	url := fmt.Sprintf("%s/api/v1/users/hard-delete/%s", c.config.Services.Auth.BaseURL, idUserExternal)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyHeader(ctx)).
		SetHeader("Content-Type", "application/json").
		Delete(url)

	if err != nil {
		return fmt.Errorf("erro ao comunicar com auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}

	return nil
}
