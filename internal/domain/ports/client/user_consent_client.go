package client

import (
	"context"

	userConsentDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user_consent"
)

// UserConsentClient interface para comunicação com o serviço de consentimentos de usuários
type UserConsentClient interface {
	// Accept registra o aceite de um consentimento
	Accept(ctx context.Context, req userConsentDto.UserConsentAcceptRequestDTO, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error)

	// FindByID busca um consentimento por ID
	FindByID(ctx context.Context, id, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error)

	// FindByUserID busca todos os consentimentos de um usuário
	FindByUserID(ctx context.Context, userID, token, tenantId, correlationID string) ([]userConsentDto.UserConsentResponseDTO, error)

	// FindByUserIDAndConsentDocumentID busca consentimentos de um usuário para um documento específico
	FindByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) ([]userConsentDto.UserConsentResponseDTO, error)

	// FindLatestByUserIDAndConsentDocumentID busca o último consentimento de um usuário para um documento
	FindLatestByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) (userConsentDto.UserConsentResponseDTO, error)

	// HasAccepted verifica se o usuário aceitou uma versão específica
	HasAccepted(ctx context.Context, userID, consentDocumentID string, version int, token, tenantId, correlationID string) (bool, error)

	// AcceptAll registra o aceite de todos os documentos de consentimento publicados
	AcceptAll(ctx context.Context, req userConsentDto.UserConsentAcceptAllRequestDTO, tenantId, correlationID string) (userConsentDto.UserConsentAcceptAllResponseDTO, error)

	// AcceptBatch registra o aceite seletivo em lote (modal LGPD).
	AcceptBatch(ctx context.Context, req userConsentDto.UserConsentAcceptBatchRequestDTO, token, tenantId, correlationID string) (userConsentDto.UserConsentAcceptAllResponseDTO, error)

	// DeleteAllByUserId deleta todos os consentimentos de um usuário (para compensação de SAGA)
	DeleteAllByUserId(ctx context.Context, userID, tenantId, correlationID string) error
}
