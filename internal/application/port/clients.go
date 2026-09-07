package port

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type UserClient interface {
	CreateUser(ctx context.Context, req appdto.MSUserCreateRequestDTO, tenantId, correlationID string) (appdto.MSUserResponseDTO, error)
	GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (appdto.MSUserResponseDTO, error)
	GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (appdto.UserByEmailResponseDTO, error)
	CreateUserNotify(ctx context.Context, req appdto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (appdto.MSUserNotifyResponseDTO, error)
	InitRegister(ctx context.Context, req appdto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (appdto.MSUserRegisterInitResponseDTO, error)
	ConfirmRegister(ctx context.Context, req appdto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (appdto.MSUserRegisterConfirmResponseDTO, error)
	DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error
	ResendRegisterToken(ctx context.Context, req appdto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (appdto.MSUserRegisterResendResponseDTO, error)
}

type AuthClient interface {
	ValidateToken(ctx context.Context, token, tenantId, correlationID string) error
	GenerateResetToken(ctx context.Context, req appdto.GenerateResetTokenMSRequestDTO, tenantId, correlationID string) (appdto.GenerateResetTokenMSResponseDTO, error)
	CreateUser(ctx context.Context, req appdto.AuthUserCreateRequestDTO, tenantId, correlationID string) (appdto.AuthUserCreateResponseDTO, error)
	RegisterLogin(ctx context.Context, req appdto.AuthRegisterLoginRequestDTO, tenantId, correlationID, clientId string) (appdto.AuthLoginResponseDTO, error)
	HardDeleteUser(ctx context.Context, idUserExternal, tenantId, correlationID string) error
}

type CompanyClient interface {
	GetByTenantId(ctx context.Context, tenantId, correlationID string) (appdto.MSCompanyResponseDTO, error)
}

type CommunicationClient interface {
	SendNotification(ctx context.Context, req appdto.SendNotificationRequestDTO, tenantId, correlationID string) error
	SendMessage(ctx context.Context, req appdto.SendMessageRequestDTO, tenantId, correlationID string) (appdto.SendMessageResponseDTO, error)
}

type ConsentDocumentClient interface {
	FindLatestPublishedByType(ctx context.Context, consentType, token, tenantId, correlationID string) (appdto.ConsentDocumentResponseDTO, error)
	FindAllPublished(ctx context.Context, token, tenantId, correlationID string) ([]appdto.ConsentDocumentResponseDTO, error)
}

type UserConsentClient interface {
	Accept(ctx context.Context, req appdto.UserConsentAcceptRequestDTO, token, tenantId, correlationID string) (appdto.UserConsentResponseDTO, error)
	FindByID(ctx context.Context, id, token, tenantId, correlationID string) (appdto.UserConsentResponseDTO, error)
	FindByUserID(ctx context.Context, userID, token, tenantId, correlationID string) ([]appdto.UserConsentResponseDTO, error)
	FindByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) ([]appdto.UserConsentResponseDTO, error)
	FindLatestByUserIDAndConsentDocumentID(ctx context.Context, userID, consentDocumentID, token, tenantId, correlationID string) (appdto.UserConsentResponseDTO, error)
	HasAccepted(ctx context.Context, userID, consentDocumentID string, version int, token, tenantId, correlationID string) (bool, error)
	AcceptAll(ctx context.Context, req appdto.UserConsentAcceptAllRequestDTO, tenantId, correlationID string) (appdto.UserConsentAcceptAllResponseDTO, error)
	AcceptBatch(ctx context.Context, req appdto.UserConsentAcceptBatchRequestDTO, token, tenantId, correlationID string) (appdto.UserConsentAcceptAllResponseDTO, error)
	DeleteAllByUserId(ctx context.Context, userID, tenantId, correlationID string) error
}
