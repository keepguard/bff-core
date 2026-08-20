package company

import (
	"context"
	"net/http"
	"time"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
)

// companyMetricsDecorator implementa métricas para CompanyClient
type companyMetricsDecorator struct {
	inner       portsclient.CompanyClient
	metrics     *metrics.Metrics
	serviceName string
}

// NewCompanyMetricsDecorator cria um decorator de métricas para CompanyClient
func NewCompanyMetricsDecorator(
	inner portsclient.CompanyClient,
	metrics *metrics.Metrics,
	serviceName string,
) portsclient.CompanyClient {
	return &companyMetricsDecorator{
		inner:       inner,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

// getStatusCodeFromError extrai o status code de um erro
func (d *companyMetricsDecorator) getStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.Code
	}

	return http.StatusInternalServerError
}

// GetByTenantId implementa GetByTenantId com métricas
func (d *companyMetricsDecorator) GetByTenantId(ctx context.Context, tenantId, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.GetByTenantId(ctx, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/companies/by-x-tenant-id", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "GET", "/companies/by-x-tenant-id", errorType)
	}

	return response, err
}
