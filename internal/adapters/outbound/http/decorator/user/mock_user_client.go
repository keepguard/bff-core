package user

import (
	"context"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	"github.com/stretchr/testify/mock"
)

// MockUserClient é um mock para UserClient
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserResponseDTO), args.Error(1)
}

func (m *MockUserClient) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	args := m.Called(ctx, codeUser, token, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserResponseDTO), args.Error(1)
}

func (m *MockUserClient) GetByEmail(ctx context.Context, email, tenantId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	args := m.Called(ctx, email, tenantId, correlationID)
	return args.Get(0).(authDto.UserByEmailResponseDTO), args.Error(1)
}

func (m *MockUserClient) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserNotifyResponseDTO), args.Error(1)
}

func (m *MockUserClient) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserRegisterInitResponseDTO), args.Error(1)
}

func (m *MockUserClient) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserRegisterConfirmResponseDTO), args.Error(1)
}

func (m *MockUserClient) DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error {
	args := m.Called(ctx, userID, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockUserClient) ResendRegisterToken(ctx context.Context, req userDto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(userDto.MSUserRegisterResendResponseDTO), args.Error(1)
}
