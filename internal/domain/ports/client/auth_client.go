package client

import (
	"context"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
)

// AuthClient define a interface para o cliente de autenticação (ms-auth)
type AuthClient interface {
	ValidateToken(ctx context.Context, token, xApplication, correlationID string) error
	GenerateResetToken(ctx context.Context, req authDto.GenerateResetTokenMSRequestDTO, xApplication, correlationID string) (authDto.GenerateResetTokenMSResponseDTO, error)
	CreateUser(ctx context.Context, req authDto.AuthUserCreateRequestDTO, xApplication, correlationID string) (authDto.AuthUserCreateResponseDTO, error)

	// RegisterLogin realiza login após registro usando senha criptografada
	RegisterLogin(ctx context.Context, req authDto.AuthRegisterLoginRequestDTO, xApplication, correlationID string) (authDto.AuthLoginResponseDTO, error)

	// HardDeleteUser remove permanentemente um usuário (para compensação de SAGA)
	HardDeleteUser(ctx context.Context, idUserExternal, xApplication, correlationID string) error
}
