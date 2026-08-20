package client

import (
	"context"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
)

// CompanyClient interface para comunicação com o serviço de empresas (ms-company)
type CompanyClient interface {
	// GetByTenantId busca uma empresa pelo X-Tenant-Id
	GetByTenantId(ctx context.Context, tenantId, correlationID string) (companyDto.MSCompanyResponseDTO, error)
}
