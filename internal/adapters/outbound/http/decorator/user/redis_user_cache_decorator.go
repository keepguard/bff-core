package user

import (
	"context"
	"encoding/json"
	"strings"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
	companydecorator "github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/company"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

const (
	userByCodeCachePrefix  = "user_cache:user:codeuser:"
	userByEmailCachePrefix = "user_cache:user:email:"
)

type redisUserCacheDecorator struct {
	portsclient.UserClient
	cache   companydecorator.StringValueCache
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// NewRedisUserCacheDecorator lê user_cache:user:codeuser:{codeUser} (gravado pelo ms-user) antes do HTTP.
func NewRedisUserCacheDecorator(
	inner portsclient.UserClient,
	cache companydecorator.StringValueCache,
	metricsInstance *metrics.Metrics,
	logger *zap.Logger,
) portsclient.UserClient {
	if cache == nil {
		return inner
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &redisUserCacheDecorator{
		UserClient: inner,
		cache:      cache,
		metrics:    metricsInstance,
		logger:     logger,
	}
}

func (d *redisUserCacheDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	if codeUser != "" {
		if user, ok := d.lookup(ctx, codeUser); ok {
			d.recordHit("codeUser")
			return user, nil
		}
	}

	d.recordMiss("codeUser")
	return d.UserClient.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
}

func (d *redisUserCacheDecorator) GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	if email != "" && companyId != "" {
		if user, ok := d.lookupByEmail(ctx, companyId, email); ok {
			d.recordHit("email")
			return user, nil
		}
	}

	d.recordMiss("email")
	return d.UserClient.GetByEmail(ctx, email, tenantId, companyId, correlationID)
}

func (d *redisUserCacheDecorator) lookup(ctx context.Context, codeUser string) (userDto.MSUserResponseDTO, bool) {
	key := userByCodeCachePrefix + strings.ToLower(strings.TrimSpace(codeUser))
	val, err := d.cache.Get(ctx, key)
	if err != nil {
		d.logger.Warn("Falha ao ler usuário no Redis; fallback HTTP",
			zap.String("key", key),
			zap.Error(err),
		)
		return userDto.MSUserResponseDTO{}, false
	}
	if val == "" {
		return userDto.MSUserResponseDTO{}, false
	}

	var user userDto.MSUserResponseDTO
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return userDto.MSUserResponseDTO{}, false
	}
	if user.ID == "" && user.CodeUser == "" && user.Email == "" {
		return userDto.MSUserResponseDTO{}, false
	}

	if companyID := portsclient.CompanyIDFromContext(ctx); companyID != "" && user.CompanyID != "" && !strings.EqualFold(user.CompanyID, companyID) {
		return userDto.MSUserResponseDTO{}, false
	}

	return user, true
}

func (d *redisUserCacheDecorator) lookupByEmail(ctx context.Context, companyId, email string) (authDto.UserByEmailResponseDTO, bool) {
	key := userByEmailCachePrefix + companyId + ":" + strings.ToLower(strings.TrimSpace(email))
	val, err := d.cache.Get(ctx, key)
	if err != nil {
		d.logger.Warn("Falha ao ler usuário por email no Redis; fallback HTTP",
			zap.String("key", key),
			zap.Error(err),
		)
		return authDto.UserByEmailResponseDTO{}, false
	}
	if val == "" {
		return authDto.UserByEmailResponseDTO{}, false
	}

	var user authDto.UserByEmailResponseDTO
	if err := json.Unmarshal([]byte(val), &user); err != nil || (user.Email == "" && user.CodeUser == "") {
		return authDto.UserByEmailResponseDTO{}, false
	}
	return user, true
}

func (d *redisUserCacheDecorator) recordHit(keyPattern string) {
	if d.metrics != nil {
		d.metrics.RecordCacheHit("user-redis", keyPattern)
	}
}

func (d *redisUserCacheDecorator) recordMiss(keyPattern string) {
	if d.metrics != nil {
		d.metrics.RecordCacheMiss("user-redis", keyPattern)
	}
}
