package user

import (
	"context"
	"errors"
	"testing"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
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

func TestRedisUserCacheDecorator_HitSkipsHTTP(t *testing.T) {
	inner := new(MockUserClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "user_cache:user:codeuser:user-1").
		Return(`{"id":"id-1","codeUser":"user-1","companyId":"co-1","email":"a@b.c"}`, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	ctx := portsclient.WithCompanyID(context.Background(), "co-1")
	result, err := decorator.GetUserByCodeUser(ctx, "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "a@b.c", result.Email)
	inner.AssertNotCalled(t, "GetUserByCodeUser")
}

func TestRedisUserCacheDecorator_MissFallsBackToHTTP(t *testing.T) {
	inner := new(MockUserClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, mock.Anything).Return("", nil)
	inner.On("GetUserByCodeUser", mock.Anything, "user-1", "tok", "tenant-1", "corr-1").
		Return(userDto.MSUserResponseDTO{Email: "http@b.c"}, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetUserByCodeUser(context.Background(), "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "http@b.c", result.Email)
	inner.AssertExpectations(t)
}

func TestRedisUserCacheDecorator_RedisErrorFallsBackToHTTP(t *testing.T) {
	inner := new(MockUserClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "user_cache:user:codeuser:user-1").
		Return("", errors.New("redis down"))
	inner.On("GetUserByCodeUser", mock.Anything, "user-1", "tok", "tenant-1", "corr-1").
		Return(userDto.MSUserResponseDTO{Email: "http@b.c"}, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetUserByCodeUser(context.Background(), "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "http@b.c", result.Email)
}

func TestRedisUserCacheDecorator_NilCacheUsesInner(t *testing.T) {
	inner := new(MockUserClient)
	inner.On("GetUserByCodeUser", mock.Anything, "user-1", "tok", "tenant-1", "corr-1").
		Return(userDto.MSUserResponseDTO{Email: "http@b.c"}, nil)

	decorator := NewRedisUserCacheDecorator(inner, nil, nil, nil)
	result, err := decorator.GetUserByCodeUser(context.Background(), "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "http@b.c", result.Email)
}

func TestRedisUserCacheDecorator_GetByEmailHitSkipsHTTP(t *testing.T) {
	inner := new(MockUserClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "user_cache:user:email:co-1:a@b.c").
		Return(`{"id":"id-1","codeUser":"user-1","email":"a@b.c","status":"ACTIVE"}`, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetByEmail(context.Background(), "a@b.c", "tenant-1", "co-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "a@b.c", result.Email)
	inner.AssertNotCalled(t, "GetByEmail")
}

func TestRedisUserCacheDecorator_GetByEmailMissFallsBackToHTTP(t *testing.T) {
	inner := new(MockUserClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, mock.Anything).Return("", nil)
	inner.On("GetByEmail", mock.Anything, "a@b.c", "tenant-1", "co-1", "corr-1").
		Return(authDto.UserByEmailResponseDTO{Email: "http@b.c"}, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetByEmail(context.Background(), "a@b.c", "tenant-1", "co-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "http@b.c", result.Email)
	inner.AssertExpectations(t)
}
