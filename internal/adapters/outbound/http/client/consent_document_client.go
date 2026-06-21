package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	consentDocumentDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/consent_document"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// consentDocumentClient implementa ConsentDocumentClient usando HTTP
type consentDocumentClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewConsentDocumentClient cria uma nova instância do ConsentDocumentClient
func NewConsentDocumentClient(config *config.Config, logger *zap.Logger) client.ConsentDocumentClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(2)
	httpClient.SetRetryWaitTime(500 * time.Millisecond)

	return &consentDocumentClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// FindLatestPublishedByType busca a última versão publicada por tipo
func (c *consentDocumentClient) FindLatestPublishedByType(ctx context.Context, consentType, token, xApplication, correlationID string) (consentDocumentDto.ConsentDocumentResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/consent-documents/type/%s/latest-published", c.config.Services.UserConsents.BaseURL, consentType)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Application", xApplication).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return consentDocumentDto.ConsentDocumentResponseDTO{}, fmt.Errorf("erro ao comunicar com user consents service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return consentDocumentDto.ConsentDocumentResponseDTO{}, &appdto.HTTPError{
			Code:    resp.StatusCode(),
			Message: "user consents service retornou erro",
			Details: string(resp.Body()),
		}
	}

	var document consentDocumentDto.ConsentDocumentResponseDTO
	if err := json.Unmarshal(resp.Body(), &document); err != nil {
		return consentDocumentDto.ConsentDocumentResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return document, nil
}
