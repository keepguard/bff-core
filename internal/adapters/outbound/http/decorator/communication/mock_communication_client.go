package communication

import (
	"context"

	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/stretchr/testify/mock"
)

// MockCommunicationClient é um mock para CommunicationClient
type MockCommunicationClient struct {
	mock.Mock
}

func (m *MockCommunicationClient) SendNotification(ctx context.Context, req portsclient.SendNotificationRequestDTO, xApplication, correlationID string) error {
	args := m.Called(ctx, req, xApplication, correlationID)
	return args.Error(0)
}

func (m *MockCommunicationClient) SendMessage(ctx context.Context, req communicationDto.SendMessageRequestDTO, xApplication, correlationID string) (communicationDto.SendMessageResponseDTO, error) {
	args := m.Called(ctx, req, xApplication, correlationID)
	return args.Get(0).(communicationDto.SendMessageResponseDTO), args.Error(1)
}
