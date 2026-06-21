package company

import (
	"context"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	"github.com/stretchr/testify/mock"
)

// MockCompanyClient é um mock para CompanyClient
type MockCompanyClient struct {
	mock.Mock
}

func (m *MockCompanyClient) GetByXApplication(ctx context.Context, xApplication, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	args := m.Called(ctx, xApplication, correlationID)
	return args.Get(0).(companyDto.MSCompanyResponseDTO), args.Error(1)
}
