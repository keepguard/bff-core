package decorator

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RegisterHandlers interface para handlers de registro
type RegisterHandlers interface {
	InitRegisterHandler(c echo.Context) error
	ConfirmRegisterHandler(c echo.Context) error
}

// UserHandlers interface para handlers de usuário
type UserHandlers interface {
	CreateUserHandler(c echo.Context) error
	GetUserByCodeHandler(c echo.Context) error
}

// registerHandlersLoggingDecorator implementa logging para RegisterHandlers
type registerHandlersLoggingDecorator struct {
	inner  RegisterHandlers
	logger *zap.Logger
}

// NewRegisterHandlersLoggingDecorator cria um decorator de logging para RegisterHandlers
func NewRegisterHandlersLoggingDecorator(
	inner RegisterHandlers,
	logger *zap.Logger,
) RegisterHandlers {
	return &registerHandlersLoggingDecorator{
		inner:  inner,
		logger: logger,
	}
}

// InitRegisterHandler implementa InitRegisterHandler com logging
func (d *registerHandlersLoggingDecorator) InitRegisterHandler(c echo.Context) error {
	start := time.Now()

	// Extrai informações da requisição
	correlationID := c.Request().Header.Get("X-Correlation-ID")
	tenantId := c.Request().Header.Get("X-Tenant-Id")
	email := c.FormValue("email")

	d.logger.Info("Iniciando handler",
		zap.String("handler", "RegisterHandlers"),
		zap.String("operation", "InitRegisterHandler"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
		zap.String("method", c.Request().Method),
		zap.String("path", c.Path()),
	)

	err := d.inner.InitRegisterHandler(c)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro no handler",
			zap.String("handler", "RegisterHandlers"),
			zap.String("operation", "InitRegisterHandler"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Handler concluído com sucesso",
		zap.String("handler", "RegisterHandlers"),
		zap.String("operation", "InitRegisterHandler"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
		zap.Duration("duration", duration),
	)

	return nil
}

// ConfirmRegisterHandler implementa ConfirmRegisterHandler com logging
func (d *registerHandlersLoggingDecorator) ConfirmRegisterHandler(c echo.Context) error {
	start := time.Now()

	// Extrai informações da requisição
	correlationID := c.Request().Header.Get("X-Correlation-ID")
	tenantId := c.Request().Header.Get("X-Tenant-Id")
	email := c.FormValue("email")

	d.logger.Info("Iniciando handler",
		zap.String("handler", "RegisterHandlers"),
		zap.String("operation", "ConfirmRegisterHandler"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
		zap.String("method", c.Request().Method),
		zap.String("path", c.Path()),
	)

	err := d.inner.ConfirmRegisterHandler(c)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro no handler",
			zap.String("handler", "RegisterHandlers"),
			zap.String("operation", "ConfirmRegisterHandler"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Handler concluído com sucesso",
		zap.String("handler", "RegisterHandlers"),
		zap.String("operation", "ConfirmRegisterHandler"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
		zap.Duration("duration", duration),
	)

	return nil
}

// userHandlersLoggingDecorator implementa logging para UserHandlers
type userHandlersLoggingDecorator struct {
	inner  UserHandlers
	logger *zap.Logger
}

// NewUserHandlersLoggingDecorator cria um decorator de logging para UserHandlers
func NewUserHandlersLoggingDecorator(
	inner UserHandlers,
	logger *zap.Logger,
) UserHandlers {
	return &userHandlersLoggingDecorator{
		inner:  inner,
		logger: logger,
	}
}

// CreateUserHandler implementa CreateUserHandler com logging
func (d *userHandlersLoggingDecorator) CreateUserHandler(c echo.Context) error {
	start := time.Now()

	// Extrai informações da requisição
	correlationID := c.Request().Header.Get("X-Correlation-ID")
	tenantId := c.Request().Header.Get("X-Tenant-Id")
	email := c.FormValue("email")

	d.logger.Info("Iniciando handler",
		zap.String("handler", "UserHandlers"),
		zap.String("operation", "CreateUserHandler"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
		zap.String("method", c.Request().Method),
		zap.String("path", c.Path()),
	)

	err := d.inner.CreateUserHandler(c)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro no handler",
			zap.String("handler", "UserHandlers"),
			zap.String("operation", "CreateUserHandler"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Handler concluído com sucesso",
		zap.String("handler", "UserHandlers"),
		zap.String("operation", "CreateUserHandler"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
		zap.Duration("duration", duration),
	)

	return nil
}

// GetUserByCodeHandler implementa GetUserByCodeHandler com logging
func (d *userHandlersLoggingDecorator) GetUserByCodeHandler(c echo.Context) error {
	start := time.Now()

	// Extrai informações da requisição
	correlationID := c.Request().Header.Get("X-Correlation-ID")
	tenantId := c.Request().Header.Get("X-Tenant-Id")
	codeUser := c.Param("codeUser")

	d.logger.Info("Iniciando handler",
		zap.String("handler", "UserHandlers"),
		zap.String("operation", "GetUserByCodeHandler"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("codeUser", codeUser),
		zap.String("method", c.Request().Method),
		zap.String("path", c.Path()),
	)

	err := d.inner.GetUserByCodeHandler(c)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro no handler",
			zap.String("handler", "UserHandlers"),
			zap.String("operation", "GetUserByCodeHandler"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("codeUser", codeUser),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Handler concluído com sucesso",
		zap.String("handler", "UserHandlers"),
		zap.String("operation", "GetUserByCodeHandler"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("codeUser", codeUser),
		zap.Duration("duration", duration),
	)

	return nil
}
