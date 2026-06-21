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

func (m *MockUserClient) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, xApplication, correlationID string) (userDto.MSUserResponseDTO, error) {
	args := m.Called(ctx, req, xApplication, correlationID)
	return args.Get(0).(userDto.MSUserResponseDTO), args.Error(1)
}

func (m *MockUserClient) GetUserByCodeUser(ctx context.Context, codeUser, token, xApplication, correlationID string) (userDto.MSUserResponseDTO, error) {
	args := m.Called(ctx, codeUser, token, xApplication, correlationID)
	return args.Get(0).(userDto.MSUserResponseDTO), args.Error(1)
}

func (m *MockUserClient) GetByEmail(ctx context.Context, email, xApplication, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	args := m.Called(ctx, email, xApplication, correlationID)
	return args.Get(0).(authDto.UserByEmailResponseDTO), args.Error(1)
}

func (m *MockUserClient) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, xApplication, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	args := m.Called(ctx, req, xApplication, correlationID)
	return args.Get(0).(userDto.MSUserNotifyResponseDTO), args.Error(1)
}

func (m *MockUserClient) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, xApplication, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	args := m.Called(ctx, req, xApplication, correlationID)
	return args.Get(0).(userDto.MSUserRegisterInitResponseDTO), args.Error(1)
}

func (m *MockUserClient) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, xApplication, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	args := m.Called(ctx, req, xApplication, correlationID)
	return args.Get(0).(userDto.MSUserRegisterConfirmResponseDTO), args.Error(1)
}

func (m *MockUserClient) DeleteUser(ctx context.Context, userID, xApplication, correlationID string) error {
	args := m.Called(ctx, userID, xApplication, correlationID)
	return args.Error(0)
}
