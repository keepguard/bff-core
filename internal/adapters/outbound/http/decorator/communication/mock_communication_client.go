package communication

import (
	"context"
	appdto "github.com/keepguard/bff-core/internal/application/dto"

	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
	"github.com/stretchr/testify/mock"
)

// MockCommunicationClient é um mock para CommunicationClient
type MockCommunicationClient struct {
	mock.Mock
}

func (m *MockCommunicationClient) SendNotification(ctx context.Context, req appdto.SendNotificationRequestDTO, tenantId, correlationID string) error {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Error(0)
}

func (m *MockCommunicationClient) SendMessage(ctx context.Context, req communicationDto.SendMessageRequestDTO, tenantId, correlationID string) (communicationDto.SendMessageResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(communicationDto.SendMessageResponseDTO), args.Error(1)
}
