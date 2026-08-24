package decorator

import (
	"time"

	"github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"go.uber.org/zap"
)

// RegisterInitUseCase interface para o use case de inicialização de registro
type RegisterInitUseCase interface {
	Execute(command appdto.RegisterInitCommand) (dto.RegisterInitResponseDTO, error)
}

// RegisterConfirmUseCase interface para o use case de confirmação de registro
type RegisterConfirmUseCase interface {
	Execute(command appdto.RegisterConfirmCommand) (dto.RegisterConfirmResponseDTO, error)
}

// registerInitLoggingDecorator implementa logging para RegisterInitUseCase
type registerInitLoggingDecorator struct {
	inner  RegisterInitUseCase
	logger *zap.Logger
}

// NewRegisterInitLoggingDecorator cria um decorator de logging para RegisterInitUseCase
func NewRegisterInitLoggingDecorator(
	inner RegisterInitUseCase,
	logger *zap.Logger,
) RegisterInitUseCase {
	return &registerInitLoggingDecorator{
		inner:  inner,
		logger: logger,
	}
}

// Execute implementa Execute com logging
func (d *registerInitLoggingDecorator) Execute(command appdto.RegisterInitCommand) (dto.RegisterInitResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando caso de uso",
		zap.String("useCase", "RegisterInitUseCase"),
		zap.String("operation", "Execute"),
		zap.String("correlationID", command.CorrelationID),
		zap.String("tenantId", command.TenantId),
		zap.String("email", command.Email),
		zap.String("type", command.Type),
	)

	response, err := d.inner.Execute(command)
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

// registerConfirmLoggingDecorator implementa logging para RegisterConfirmUseCase
type registerConfirmLoggingDecorator struct {
	inner  RegisterConfirmUseCase
	logger *zap.Logger
}

// NewRegisterConfirmLoggingDecorator cria um decorator de logging para RegisterConfirmUseCase
func NewRegisterConfirmLoggingDecorator(
	inner RegisterConfirmUseCase,
	logger *zap.Logger,
) RegisterConfirmUseCase {
	return &registerConfirmLoggingDecorator{
		inner:  inner,
		logger: logger,
	}
}

// Execute implementa Execute com logging
func (d *registerConfirmLoggingDecorator) Execute(command appdto.RegisterConfirmCommand) (dto.RegisterConfirmResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando caso de uso",
		zap.String("useCase", "RegisterConfirmUseCase"),
		zap.String("operation", "Execute"),
		zap.String("correlationID", command.CorrelationID),
		zap.String("tenantId", command.TenantId),
		zap.String("email", command.Email),
		zap.String("registrationSessionId", command.RegistrationSessionId),
	)

	response, err := d.inner.Execute(command)
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
