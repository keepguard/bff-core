package decorator

import (
	"context"
	"time"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/register"
	"go.uber.org/zap"
)

type registerInitLoggingDecorator struct {
	inner  register.RegisterInitUseCase
	logger *zap.Logger
}

// NewRegisterInitLoggingDecorator cria um decorator de logging para RegisterInitUseCase
func NewRegisterInitLoggingDecorator(
	inner register.RegisterInitUseCase,
	logger *zap.Logger,
) register.RegisterInitUseCase {
	return &registerInitLoggingDecorator{
		inner:  inner,
		logger: logger,
	}
}

func (d *registerInitLoggingDecorator) Execute(ctx context.Context, command appdto.RegisterInitCommand) (appdto.RegisterInitViewDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando caso de uso",
		zap.String("useCase", "RegisterInitUseCase"),
		zap.String("operation", "Execute"),
		zap.String("correlationID", command.CorrelationID),
		zap.String("tenantId", command.TenantId),
		zap.String("email", command.Email),
		zap.String("type", command.Type),
	)

	response, err := d.inner.Execute(ctx, command)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro no caso de uso",
			zap.String("useCase", "RegisterInitUseCase"),
			zap.String("operation", "Execute"),
			zap.String("correlationID", command.CorrelationID),
			zap.String("tenantId", command.TenantId),
			zap.String("email", command.Email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Caso de uso concluído com sucesso",
		zap.String("useCase", "RegisterInitUseCase"),
		zap.String("operation", "Execute"),
		zap.String("correlationID", command.CorrelationID),
		zap.String("tenantId", command.TenantId),
		zap.String("email", command.Email),
		zap.String("registrationSessionId", response.RegistrationSessionID),
		zap.Int("expiresIn", response.ExpiresIn),
		zap.Duration("duration", duration),
	)

	return response, nil
}

type registerConfirmLoggingDecorator struct {
	inner  register.RegisterConfirmUseCase
	logger *zap.Logger
}

// NewRegisterConfirmLoggingDecorator cria um decorator de logging para RegisterConfirmUseCase
func NewRegisterConfirmLoggingDecorator(
	inner register.RegisterConfirmUseCase,
	logger *zap.Logger,
) register.RegisterConfirmUseCase {
	return &registerConfirmLoggingDecorator{
		inner:  inner,
		logger: logger,
	}
}

func (d *registerConfirmLoggingDecorator) Execute(ctx context.Context, command appdto.RegisterConfirmCommand) (appdto.RegisterConfirmViewDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando caso de uso",
		zap.String("useCase", "RegisterConfirmUseCase"),
		zap.String("operation", "Execute"),
		zap.String("correlationID", command.CorrelationID),
		zap.String("tenantId", command.TenantId),
		zap.String("email", command.Email),
		zap.String("registrationSessionId", command.RegistrationSessionId),
	)

	response, err := d.inner.Execute(ctx, command)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro no caso de uso",
			zap.String("useCase", "RegisterConfirmUseCase"),
			zap.String("operation", "Execute"),
			zap.String("correlationID", command.CorrelationID),
			zap.String("tenantId", command.TenantId),
			zap.String("email", command.Email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Caso de uso concluído com sucesso",
		zap.String("useCase", "RegisterConfirmUseCase"),
		zap.String("operation", "Execute"),
		zap.String("correlationID", command.CorrelationID),
		zap.String("tenantId", command.TenantId),
		zap.String("email", command.Email),
		zap.Duration("duration", duration),
	)

	return response, nil
}
