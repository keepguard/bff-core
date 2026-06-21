package company

import (
	"context"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/resilience"
)

// circuitBreakerDecorator implementa circuit breaker para CompanyClient
type circuitBreakerDecorator struct {
	inner          portsclient.CompanyClient
	circuitBreaker *resilience.CircuitBreakerManager
	serviceName    string
}

// NewCircuitBreakerDecorator cria um decorator de circuit breaker para CompanyClient
func NewCircuitBreakerDecorator(
	inner portsclient.CompanyClient,
	cbManager *resilience.CircuitBreakerManager,
	serviceName string,
) portsclient.CompanyClient {
	return &circuitBreakerDecorator{
		inner:          inner,
		circuitBreaker: cbManager,
		serviceName:    serviceName,
	}
}

// GetByXApplication implementa GetByXApplication com circuit breaker
func (d *circuitBreakerDecorator) GetByXApplication(ctx context.Context, xApplication, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.GetByXApplication(ctx, xApplication, correlationID)
	})

	if err != nil {
		return companyDto.MSCompanyResponseDTO{}, err
	}

	return result.(companyDto.MSCompanyResponseDTO), nil
}
