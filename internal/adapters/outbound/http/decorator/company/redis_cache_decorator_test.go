package company

import (
	"context"
	"errors"
	"testing"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStringCache struct {
	mock.Mock
}

func (m *mockStringCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func TestRedisCacheDecorator_HitSkipsHTTP(t *testing.T) {
	inner := new(MockCompanyClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "company_cache:tenantId:tenant-1").
		Return(`{"id":"company-1","name":"Keep"}`, nil)

	decorator := NewRedisCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetByTenantId(context.Background(), "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "company-1", result.ID)
	inner.AssertNotCalled(t, "GetByTenantId")
}

func TestRedisCacheDecorator_MissFallsBackToHTTP(t *testing.T) {
	inner := new(MockCompanyClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "company_cache:tenantId:tenant-1").Return("", nil)
	cache.On("Get", mock.Anything, "company_cache:xapp:tenant-1").Return("", nil)
	inner.On("GetByTenantId", mock.Anything, "tenant-1", "corr-1").
		Return(companyDto.MSCompanyResponseDTO{ID: "company-http"}, nil)

	decorator := NewRedisCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetByTenantId(context.Background(), "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "company-http", result.ID)
	inner.AssertExpectations(t)
}

func TestRedisCacheDecorator_RedisErrorFallsBackToHTTP(t *testing.T) {
	inner := new(MockCompanyClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "company_cache:tenantId:tenant-1").
		Return("", errors.New("redis down"))
	inner.On("GetByTenantId", mock.Anything, "tenant-1", "corr-1").
		Return(companyDto.MSCompanyResponseDTO{ID: "company-http"}, nil)

	decorator := NewRedisCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetByTenantId(context.Background(), "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "company-http", result.ID)
}

func TestRedisCacheDecorator_NilCacheUsesInner(t *testing.T) {
	inner := new(MockCompanyClient)
	inner.On("GetByTenantId", mock.Anything, "tenant-1", "corr-1").
		Return(companyDto.MSCompanyResponseDTO{ID: "company-http"}, nil)

	decorator := NewRedisCacheDecorator(inner, nil, nil, nil)
	result, err := decorator.GetByTenantId(context.Background(), "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "company-http", result.ID)
}
