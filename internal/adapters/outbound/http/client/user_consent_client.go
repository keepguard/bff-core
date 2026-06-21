package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	userConsentDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user_consent"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// userConsentClient implementa UserConsentClient usando HTTP
type userConsentClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewUserConsentClient cria uma nova instância do UserConsentClient
func NewUserConsentClient(config *config.Config, logger *zap.Logger) client.UserConsentClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(2)
	httpClient.SetRetryWaitTime(500 * time.Millisecond)

	return &userConsentClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// NewUserConsentClientWithoutRetry cria uma nova instância do UserConsentClient SEM retry automático
func NewUserConsentClientWithoutRetry(config *config.Config, logger *zap.Logger) client.UserConsentClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	// SEM SetRetryCount - sem retry no Resty

	return &userConsentClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// Accept registra o aceite de um consentimento
func (c *userConsentClient) Accept(ctx context.Context, req userConsentDto.UserConsentAcceptRequestDTO, token, xApplication, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/user-consents/accept", c.config.Services.UserConsents.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return userConsentDto.UserConsentResponseDTO{}, fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return userConsentDto.UserConsentResponseDTO{}, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "user consents service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var consent userConsentDto.UserConsentResponseDTO
	if err := json.Unmarshal(resp.Body(), &consent); err != nil {
		return userConsentDto.UserConsentResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return consent, nil
}

// FindByID busca um consentimento por ID
func (c *userConsentClient) FindByID(ctx context.Context, id, token, xApplication, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/user-consents/%s", c.config.Services.UserConsents.BaseURL, id)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return userConsentDto.UserConsentResponseDTO{}, fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return userConsentDto.UserConsentResponseDTO{}, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "user consents service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var consent userConsentDto.UserConsentResponseDTO
	if err := json.Unmarshal(resp.Body(), &consent); err != nil {
		return userConsentDto.UserConsentResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return consent, nil
}

// FindByUserID busca todos os consentimentos de um usuário
func (c *userConsentClient) FindByUserID(ctx context.Context, userID, token, xApplication, correlationID string) ([]userConsentDto.UserConsentResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/user-consents/user/%s", c.config.Services.UserConsents.BaseURL, userID)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "user consents service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var consents []userConsentDto.UserConsentResponseDTO
	if err := json.Unmarshal(resp.Body(), &consents); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return consents, nil
}

// FindByUserIDAndConsentDocumentID busca consentimentos de um usuário para um documento específico
func (c *userConsentClient) FindByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, xApplication, correlationID string) ([]userConsentDto.UserConsentResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/user-consents/user/%s/document/%s", c.config.Services.UserConsents.BaseURL, userID, consentDocumentID)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "user consents service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var consents []userConsentDto.UserConsentResponseDTO
	if err := json.Unmarshal(resp.Body(), &consents); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return consents, nil
}

// FindLatestByUserIDAndConsentDocumentID busca o último consentimento de um usuário para um documento
func (c *userConsentClient) FindLatestByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, xApplication, correlationID string) (userConsentDto.UserConsentResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/user-consents/user/%s/document/%s/latest", c.config.Services.UserConsents.BaseURL, userID, consentDocumentID)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return userConsentDto.UserConsentResponseDTO{}, fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return userConsentDto.UserConsentResponseDTO{}, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "user consents service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var consent userConsentDto.UserConsentResponseDTO
	if err := json.Unmarshal(resp.Body(), &consent); err != nil {
		return userConsentDto.UserConsentResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return consent, nil
}

// HasAccepted verifica se o usuário aceitou uma versão específica
func (c *userConsentClient) HasAccepted(ctx context.Context, userID, consentDocumentID string, version int, token, xApplication, correlationID string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/user-consents/user/%s/document/%s/version/%d/check", c.config.Services.UserConsents.BaseURL, userID, consentDocumentID, version)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return false, fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return false, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "user consents service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var hasAccepted bool
	if err := json.Unmarshal(resp.Body(), &hasAccepted); err != nil {
		return false, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return hasAccepted, nil
}

// AcceptAll registra o aceite de todos os documentos de consentimento publicados
func (c *userConsentClient) AcceptAll(ctx context.Context, req userConsentDto.UserConsentAcceptAllRequestDTO, xApplication, correlationID string) (userConsentDto.UserConsentAcceptAllResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/user-consents/accept-all", c.config.Services.UserConsents.BaseURL)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return userConsentDto.UserConsentAcceptAllResponseDTO{}, fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated {
		return userConsentDto.UserConsentAcceptAllResponseDTO{}, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "user consents service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var acceptAllResponse userConsentDto.UserConsentAcceptAllResponseDTO
	if err := json.Unmarshal(resp.Body(), &acceptAllResponse); err != nil {
		return userConsentDto.UserConsentAcceptAllResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return acceptAllResponse, nil
}

// DeleteAllByUserId deleta todos os consentimentos de um usuário (para compensação de SAGA)
func (c *userConsentClient) DeleteAllByUserId(ctx context.Context, userID, xApplication, correlationID string) error {
	url := fmt.Sprintf("%s/api/v1/user-consents/user/%s", c.config.Services.UserConsents.BaseURL, userID)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Delete(url)

	if err != nil {
		return fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return MapHTTPError(resp.StatusCode(), resp.Body(), "user consents service")
	}

	return nil
}
