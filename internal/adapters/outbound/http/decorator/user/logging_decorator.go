package user

import (
	"context"
	"time"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"go.uber.org/zap"
)

// userLoggingDecorator implementa logging para UserClient
type userLoggingDecorator struct {
	inner       portsclient.UserClient
	logger      *zap.Logger
	serviceName string
}

// NewUserLoggingDecorator cria um decorator de logging para UserClient
func NewUserLoggingDecorator(
	inner portsclient.UserClient,
	logger *zap.Logger,
	serviceName string,
) portsclient.UserClient {
	return &userLoggingDecorator{
		inner:       inner,
		logger:      logger,
		serviceName: serviceName,
	}
}

// CreateUser implementa CreateUser com logging
func (d *userLoggingDecorator) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "CreateUser"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", req.Email),
	)

	response, err := d.inner.CreateUser(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "CreateUser"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", req.Email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "CreateUser"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", req.Email),
		zap.String("userID", response.ID),
		zap.String("codeUser", response.CodeUser),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// GetUserByCodeUser implementa GetUserByCodeUser com logging
func (d *userLoggingDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetUserByCodeUser"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("codeUser", codeUser),
	)

	response, err := d.inner.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "GetUserByCodeUser"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("codeUser", codeUser),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetUserByCodeUser"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("codeUser", codeUser),
		zap.String("userID", response.ID),
		zap.String("email", response.Email),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// GetByEmail implementa GetByEmail com logging
func (d *userLoggingDecorator) GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetByEmail"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
	)

	response, err := d.inner.GetByEmail(ctx, email, tenantId, companyId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "GetByEmail"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetByEmail"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
		zap.String("userID", response.ID),
		zap.String("codeUser", response.CodeUser),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// CreateUserNotify implementa CreateUserNotify com logging
func (d *userLoggingDecorator) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "CreateUserNotify"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("userID", req.UserID),
	)

	response, err := d.inner.CreateUserNotify(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "CreateUserNotify"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("userID", req.UserID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "CreateUserNotify"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("userID", req.UserID),
		zap.String("notifyID", response.ID),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// InitRegister implementa InitRegister com logging
func (d *userLoggingDecorator) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "InitRegister"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", req.Email),
		zap.String("nameFull", req.NameFull),
	)

	response, err := d.inner.InitRegister(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "InitRegister"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", req.Email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "InitRegister"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", req.Email),
		zap.String("sessionID", response.RegistrationSessionID),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// ConfirmRegister implementa ConfirmRegister com logging
func (d *userLoggingDecorator) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "ConfirmRegister"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", req.Email),
		zap.String("sessionID", req.RegistrationSessionID),
	)

	response, err := d.inner.ConfirmRegister(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "ConfirmRegister"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", req.Email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "ConfirmRegister"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", req.Email),
		zap.String("message", response.Message),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// DeleteUser implementa DeleteUser com logging
func (d *userLoggingDecorator) DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "DeleteUser"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("userID", userID),
	)

	err := d.inner.DeleteUser(ctx, userID, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "DeleteUser"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("userID", userID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "DeleteUser"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("userID", userID),
		zap.Duration("duration", duration),
	)

	return nil
}

// ResendRegisterToken implementa ResendRegisterToken com logging
func (d *userLoggingDecorator) ResendRegisterToken(ctx context.Context, req userDto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "ResendRegisterToken"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", req.Email),
		zap.String("registrationSessionID", req.RegistrationSessionID),
	)

	response, err := d.inner.ResendRegisterToken(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "ResendRegisterToken"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", req.Email),
			zap.String("registrationSessionID", req.RegistrationSessionID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return userDto.MSUserRegisterResendResponseDTO{}, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "ResendRegisterToken"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", req.Email),
		zap.String("registrationSessionID", req.RegistrationSessionID),
		zap.Duration("duration", duration),
	)

	return response, nil
}
