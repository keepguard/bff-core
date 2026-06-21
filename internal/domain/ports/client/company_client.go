package client

import (
	"context"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
)

// CompanyClient interface para comunicação com o serviço de empresas (ms-company)
type CompanyClient interface {
	// GetByXApplication busca uma empresa pelo X-Application
	GetByXApplication(ctx context.Context, xApplication, correlationID string) (companyDto.MSCompanyResponseDTO, error)
}
