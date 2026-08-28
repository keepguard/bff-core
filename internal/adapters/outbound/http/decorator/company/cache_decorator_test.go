package company

import (
	"context"
	"testing"
	"time"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCacheDecorator_GetByTenantId_CacheHit(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	expectedResponse := companyDto.MSCompanyResponseDTO{
		ID:   "123",
		CodeCompany: "TEST123",
		Name: "Test Company",
	}

	// Primeira chamada - deve ir para o mock
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(expectedResponse, nil).Once()

	// Act & Assert
	// Primeira chamada - cache miss
	result1, err1 := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")
	assert.NoError(t, err1)
	assert.Equal(t, expectedResponse, result1)

	// Segunda chamada - cache hit (não deve chamar o mock novamente)
	result2, err2 := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result2)

	// Verifica que o mock foi chamado apenas uma vez
	mockClient.AssertExpectations(t)
}

func TestCacheDecorator_GetByTenantId_CacheExpired(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             100 * time.Millisecond, // TTL muito curto
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	expectedResponse := companyDto.MSCompanyResponseDTO{
		ID:   "123",
		CodeCompany: "TEST123",
		Name: "Test Company",
	}

	// Deve chamar o mock duas vezes (cache miss + cache expired)
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(expectedResponse, nil).Twice()

	// Act & Assert
	// Primeira chamada - cache miss
	result1, err1 := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")
	assert.NoError(t, err1)
	assert.Equal(t, expectedResponse, result1)

	// Aguarda o cache expirar
	time.Sleep(150 * time.Millisecond)

	// Segunda chamada - cache expired, deve chamar o mock novamente
	result2, err2 := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result2)

	// Verifica que o mock foi chamado duas vezes
	mockClient.AssertExpectations(t)
}

func TestCacheDecorator_GetByTenantId_Error(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	expectedError := assert.AnError

	// Mock retorna erro
	mockClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(companyDto.MSCompanyResponseDTO{}, expectedError).Once()

	// Act
	result, err := decorator.GetByTenantId(context.Background(), "test-app", "corr-123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, companyDto.MSCompanyResponseDTO{}, result)

	mockClient.AssertExpectations(t)
}

func TestCacheDecorator_GetByTenantId_MaxSizeExceeded(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         2, // Tamanho máximo pequeno
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	// Mock para 3 chamadas diferentes
	mockClient.On("GetByTenantId", mock.Anything, "app1", "corr-1").Return(companyDto.MSCompanyResponseDTO{ID: "1"}, nil).Once()
	mockClient.On("GetByTenantId", mock.Anything, "app2", "corr-2").Return(companyDto.MSCompanyResponseDTO{ID: "2"}, nil).Once()
	mockClient.On("GetByTenantId", mock.Anything, "app3", "corr-3").Return(companyDto.MSCompanyResponseDTO{ID: "3"}, nil).Once()

	// Act
	// Primeira chamada - app1
	result1, err1 := decorator.GetByTenantId(context.Background(), "app1", "corr-1")
	assert.NoError(t, err1)
	assert.Equal(t, "1", result1.ID)

	// Segunda chamada - app2
	result2, err2 := decorator.GetByTenantId(context.Background(), "app2", "corr-2")
	assert.NoError(t, err2)
	assert.Equal(t, "2", result2.ID)

	// Terceira chamada - app3 (deve evictar app1)
	result3, err3 := decorator.GetByTenantId(context.Background(), "app3", "corr-3")
	assert.NoError(t, err3)
	assert.Equal(t, "3", result3.ID)

	// Quarta chamada - app1 novamente (deve chamar o mock, pois foi evictado)
	mockClient.On("GetByTenantId", mock.Anything, "app1", "corr-1").Return(companyDto.MSCompanyResponseDTO{ID: "1"}, nil).Once()
	result4, err4 := decorator.GetByTenantId(context.Background(), "app1", "corr-1")
	assert.NoError(t, err4)
	assert.Equal(t, "1", result4.ID)

	// Assert
	mockClient.AssertExpectations(t)
}

func TestCacheDecorator_Stop(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 100 * time.Millisecond,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	// Act
	decorator.Stop()

	// Aguarda um pouco para garantir que o cleanup foi parado
	time.Sleep(200 * time.Millisecond)

	// Assert - não deve haver panics ou erros
	assert.NotNil(t, decorator)
}

func TestCacheDecorator_GenerateCacheKey(t *testing.T) {
	// Arrange
	mockClient := &MockCompanyClient{}
	metricsInstance := getTestMetrics()
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         10,
		CleanupInterval: 30 * time.Second,
	}

	decorator := NewCacheDecorator(mockClient, config, metricsInstance).(*cacheDecorator)

	// Act
	key1 := decorator.generateCacheKey("test-app-1")
	key2 := decorator.generateCacheKey("test-app-2")
	key3 := decorator.generateCacheKey("test-app-1") // Mesmo input

	// Assert
	assert.NotEmpty(t, key1)
	assert.NotEmpty(t, key2)
	assert.NotEmpty(t, key3)
	assert.Equal(t, key1, key3)    // Mesmo input deve gerar mesma chave
	assert.NotEqual(t, key1, key2) // Inputs diferentes devem gerar chaves diferentes
	assert.Contains(t, key1, "company:tenantId:")
}
