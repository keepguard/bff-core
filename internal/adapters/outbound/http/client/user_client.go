package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// userClient implementa UserClient usando HTTP
type userClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewUserClient cria uma nova instância do UserClient
func NewUserClient(config *config.Config, logger *zap.Logger) domainclient.UserClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(2)
	httpClient.SetRetryWaitTime(500 * time.Millisecond)

	return &userClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// NewUserClientWithoutRetry cria uma nova instância do UserClient SEM retry automático
func NewUserClientWithoutRetry(config *config.Config, logger *zap.Logger) domainclient.UserClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	// SEM SetRetryCount - sem retry no Resty

	return &userClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// CreateUser cria um novo usuário
func (c *userClient) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/users", c.config.Services.User.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return userDto.MSUserResponseDTO{}, fmt.Errorf("erro ao comunicar com user service: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return userDto.MSUserResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	}

	var user userDto.MSUserResponseDTO
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		return userDto.MSUserResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return user, nil
}

// GetUserByCodeUser busca um usuário pelo codeUser via API interna do ms-user.
// Não encaminha o JWT do browser: o filtro JWT do ms-user rejeita o token do usuário
// (e o endpoint público exige X-Company-Id, que o JWT de login não carrega).
func (c *userClient) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	url := fmt.Sprintf("%s/internal/v1/users/code/%s", c.config.Services.User.BaseURL, codeUser)

	req := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json")
	if companyID := domainclient.CompanyIDFromContext(ctx); companyID != "" {
		req.SetHeader("X-Company-Id", companyID)
	}
	resp, err := req.Get(url)

	if err != nil {
		return userDto.MSUserResponseDTO{}, fmt.Errorf("erro ao comunicar com user service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return userDto.MSUserResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	}

	var user userDto.MSUserResponseDTO
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		return userDto.MSUserResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return user, nil
}

// GetByEmail busca um usuário por email no ms-user
func (c *userClient) GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/users/email/%s", c.config.Services.User.BaseURL, email)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("X-Company-Id", companyId).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return authDto.UserByEmailResponseDTO{}, fmt.Errorf("erro ao comunicar com ms-user service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return authDto.UserByEmailResponseDTO{}, fmt.Errorf("ms-user service retornou erro %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var user authDto.UserByEmailResponseDTO
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		return authDto.UserByEmailResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return user, nil
}

// CreateUserNotify cria preferências de notificação para um usuário
func (c *userClient) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/users/notify", c.config.Services.User.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return userDto.MSUserNotifyResponseDTO{}, fmt.Errorf("erro ao comunicar com user service: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return userDto.MSUserNotifyResponseDTO{}, fmt.Errorf("user service retornou erro %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var userNotify userDto.MSUserNotifyResponseDTO
	if err := json.Unmarshal(resp.Body(), &userNotify); err != nil {
		return userDto.MSUserNotifyResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return userNotify, nil
}

// InitRegister inicializa o processo de registro de usuário
func (c *userClient) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/register/init", c.config.Services.User.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return userDto.MSUserRegisterInitResponseDTO{}, fmt.Errorf("erro ao comunicar com user service: %w", err)
	}

	// Tratar diferentes códigos de status HTTP
	switch resp.StatusCode() {
	case http.StatusCreated:
		// Sucesso - continuar com o processamento
	default:
		// Usar MapHTTPError para extrair mensagem detalhada do ms-user
		return userDto.MSUserRegisterInitResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	}

	var registerResp userDto.MSUserRegisterInitResponseDTO
	if err := json.Unmarshal(resp.Body(), &registerResp); err != nil {
		return userDto.MSUserRegisterInitResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return registerResp, nil
}

// ConfirmRegister confirma o registro de usuário com o token
func (c *userClient) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/register/confirm", c.config.Services.User.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json").
		Post(url)

	// Erro de rede (timeout, connection refused, etc)
	if err != nil {
		return userDto.MSUserRegisterConfirmResponseDTO{}, MapNetworkError(err, "user service")
	}

	// Tratamento específico por status code
	switch resp.StatusCode() {
	case http.StatusOK:
		// Parse resposta de sucesso
		var confirmResp userDto.MSUserRegisterConfirmResponseDTO
		if err := json.Unmarshal(resp.Body(), &confirmResp); err != nil {
			return userDto.MSUserRegisterConfirmResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
		}
		return confirmResp, nil

	case http.StatusBadRequest:
		// Usar a mensagem específica do ms-user (token inválido, max tentativas, etc)
		return userDto.MSUserRegisterConfirmResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")

	case http.StatusNotFound:
		// Sessão não encontrada
		return userDto.MSUserRegisterConfirmResponseDTO{}, &appdto.HTTPError{
			Code:    http.StatusNotFound,
			Message: "Sessão de registro não encontrada",
			Details: string(resp.Body()),
		}

	case http.StatusConflict:
		// Email já confirmado
		return userDto.MSUserRegisterConfirmResponseDTO{}, &appdto.HTTPError{
			Code:    http.StatusConflict,
			Message: "Registro já confirmado anteriormente",
			Details: string(resp.Body()),
		}

	default:
		// Outros erros (incluindo 5xx)
		return userDto.MSUserRegisterConfirmResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	}
}

// DeleteUser deleta um usuário (para compensação de SAGA)
func (c *userClient) DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error {
	url := fmt.Sprintf("%s/api/v1/users/%s", c.config.Services.User.BaseURL, userID)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json").
		Delete(url)

	if err != nil {
		return fmt.Errorf("erro ao comunicar com user service: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	}

	return nil
}

// ResendRegisterToken reenvia o token de registro
func (c *userClient) ResendRegisterToken(ctx context.Context,
	req userDto.MSUserRegisterResendRequestDTO,
	tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error) {

	url := fmt.Sprintf("%s/api/v1/register/resend", c.config.Services.User.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return userDto.MSUserRegisterResendResponseDTO{}, MapNetworkError(err, "user service")
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var resendResp userDto.MSUserRegisterResendResponseDTO
		if err := json.Unmarshal(resp.Body(), &resendResp); err != nil {
			return userDto.MSUserRegisterResendResponseDTO{}, fmt.Errorf("erro ao fazer parse: %w", err)
		}
		return resendResp, nil
	case http.StatusBadRequest:
		return userDto.MSUserRegisterResendResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	case http.StatusNotFound:
		return userDto.MSUserRegisterResendResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	default:
		return userDto.MSUserRegisterResendResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	}
}
