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

// GetByXApplication implementa GetByXApplication com logging
func (d *companyLoggingDecorator) GetByXApplication(ctx context.Context, xApplication, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetByXApplication"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
	)

	response, err := d.inner.GetByXApplication(ctx, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "GetByXApplication"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetByXApplication"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.String("companyID", response.ID),
		zap.String("companyName", response.Name),
		zap.Duration("duration", duration),
	)

	return response, nil
}
