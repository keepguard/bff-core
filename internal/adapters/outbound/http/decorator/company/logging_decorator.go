package company

import (
	"context"
	"time"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"go.uber.org/zap"
)

// companyLoggingDecorator implementa logging para CompanyClient
type companyLoggingDecorator struct {
	inner       portsclient.CompanyClient
	logger      *zap.Logger
	serviceName string
}

// NewCompanyLoggingDecorator cria um decorator de logging para CompanyClient
func NewCompanyLoggingDecorator(
	inner portsclient.CompanyClient,
	logger *zap.Logger,
	serviceName string,
) portsclient.CompanyClient {
	return &companyLoggingDecorator{
		inner:       inner,
		logger:      logger,
		serviceName: serviceName,
	}
}

// GetByTenantId implementa GetByTenantId com logging
func (d *companyLoggingDecorator) GetByTenantId(ctx context.Context, tenantId, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetByTenantId"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
	)

	response, err := d.inner.GetByTenantId(ctx, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "GetByTenantId"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetByTenantId"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("companyID", response.ID),
		zap.String("companyName", response.Name),
		zap.Duration("duration", duration),
	)

	return response, nil
}
