package register

import (
	"context"

	outport "github.com/keepguard/bff-core/internal/application/port/out"
	"github.com/stretchr/testify/mock"
)

// MockMessagePublisher é um mock para MessagePublisher
type MockMessagePublisher struct {
	mock.Mock
}

// PublishMessage implementa MessagePublisher.PublishMessage
func (m *MockMessagePublisher) PublishMessage(ctx context.Context, message outport.MessageDTO) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// Close implementa MessagePublisher.Close
func (m *MockMessagePublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}
