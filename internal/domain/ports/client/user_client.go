package client

import (
	"context"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
)

// UserClient interface para comunicação com o serviço de usuários (ms-user)
type UserClient interface {
	// CreateUser cria um novo usuário
	CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error)

	// GetUserByCodeUser busca um usuário pelo codeUser
	GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error)

	// GetByEmail busca um usuário por email no ms-auth
	GetByEmail(ctx context.Context, email, tenantId, correlationID string) (authDto.UserByEmailResponseDTO, error)

	// CreateUserNotify cria preferências de notificação para um usuário
	CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error)

	// InitRegister inicializa o processo de registro de usuário
	InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error)

	// ConfirmRegister confirma o registro de usuário com o token
	ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error)

	// DeleteUser deleta um usuário (para compensação de SAGA)
	DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error

	// ResendRegisterToken reenvia o token de registro
	ResendRegisterToken(ctx context.Context, req userDto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error)
}
