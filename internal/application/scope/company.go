package scope

import (
	"context"
	"net/http"
	"strings"

	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
)

func Unavailable(message string) *pkg.AppError {
	return pkg.NewAppError("SERVICE_UNAVAILABLE", message, http.StatusServiceUnavailable)
}

func ResolveCompany(ctx context.Context, companies port.CompanyClient, companyFromCtx, tenantID, correlationID string) (string, error) {
	if companyFromCtx != "" {
		return companyFromCtx, nil
	}
	if strings.TrimSpace(tenantID) == "" {
		return "", pkg.NewAppError("MISSING_TENANT", "tenantId do JWT é obrigatório", http.StatusBadRequest)
	}
	if companies == nil {
		return "", Unavailable("Gestão de agents indisponível")
	}
	company, err := companies.GetByTenantId(ctx, tenantID, correlationID)
	if err != nil {
		return "", err
	}
	if company.ID == "" {
		return "", pkg.NewAppError("COMPANY_NOT_FOUND", "Empresa não encontrada para o tenant autenticado", http.StatusNotFound)
	}
	return company.ID, nil
}
