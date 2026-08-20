package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// companyClient implementa CompanyClient usando HTTP
type companyClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewCompanyClient cria uma nova instância do CompanyClient
func NewCompanyClient(config *config.Config, logger *zap.Logger) client.CompanyClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(3)
	httpClient.SetRetryWaitTime(1 * time.Second)

	return &companyClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// NewCompanyClientWithoutRetry cria uma nova instância do CompanyClient SEM retry automático
func NewCompanyClientWithoutRetry(config *config.Config, logger *zap.Logger) client.CompanyClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	// SEM SetRetryCount - sem retry no Resty

	return &companyClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// GetByTenantId busca uma empresa pelo X-Tenant-Id
func (c *companyClient) GetByTenantId(ctx context.Context, tenantId, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/companies/x-tenant-id/%s", c.config.Services.Company.BaseURL, tenantId)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return companyDto.MSCompanyResponseDTO{}, fmt.Errorf("erro ao comunicar com company service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return companyDto.MSCompanyResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "company service")
	}

	var company companyDto.MSCompanyResponseDTO
	if err := json.Unmarshal(resp.Body(), &company); err != nil {
		return companyDto.MSCompanyResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return company, nil
}
