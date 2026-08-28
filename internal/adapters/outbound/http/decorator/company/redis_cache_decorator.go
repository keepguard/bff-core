package company

import (
	"context"
	"encoding/json"
	"errors"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const companyCachePrefix = "company_cache:"

// StringValueCache lê strings no Redis (ou em fake de teste).
type StringValueCache interface {
	Get(ctx context.Context, key string) (string, error)
}

type redisStringCache struct {
	client *redis.Client
}

// NewRedisStringCache adapta o cliente Redis. Nil se o client for nil.
func NewRedisStringCache(client *redis.Client) StringValueCache {
	if client == nil {
		return nil
	}
	return &redisStringCache{client: client}
}

func (r *redisStringCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return val, err
}

type redisCacheDecorator struct {
	inner   portsclient.CompanyClient
	cache   StringValueCache
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// NewRedisCacheDecorator consulta company_cache:tenantId no Redis antes do HTTP.
// Não grava cache: a escrita continua no ms-company.
func NewRedisCacheDecorator(
	inner portsclient.CompanyClient,
	cache StringValueCache,
	metricsInstance *metrics.Metrics,
	logger *zap.Logger,
) portsclient.CompanyClient {
	if cache == nil {
		return inner
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &redisCacheDecorator{
		inner:   inner,
		cache:   cache,
		metrics: metricsInstance,
		logger:  logger,
	}
}

func (d *redisCacheDecorator) GetByTenantId(ctx context.Context, tenantId, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	if tenantId != "" {
		if company, ok := d.lookup(ctx, tenantId); ok {
			d.recordHit()
			return company, nil
		}
	}

	d.recordMiss()
	return d.inner.GetByTenantId(ctx, tenantId, correlationID)
}

func (d *redisCacheDecorator) lookup(ctx context.Context, tenantId string) (companyDto.MSCompanyResponseDTO, bool) {
	keys := []string{
		companyCachePrefix + "tenantId:" + tenantId,
		companyCachePrefix + "xapp:" + tenantId,
	}
	for _, key := range keys {
		val, err := d.cache.Get(ctx, key)
		if err != nil {
			d.logger.Warn("Falha ao ler company no Redis; fallback HTTP",
				zap.String("key", key),
				zap.Error(err),
			)
			return companyDto.MSCompanyResponseDTO{}, false
		}
		if val == "" {
			continue
		}
		var company companyDto.MSCompanyResponseDTO
		if err := json.Unmarshal([]byte(val), &company); err != nil || company.ID == "" {
			continue
		}
		return company, true
	}
	return companyDto.MSCompanyResponseDTO{}, false
}

func (d *redisCacheDecorator) recordHit() {
	if d.metrics != nil {
		d.metrics.RecordCacheHit("company-redis", "tenantId")
	}
}

func (d *redisCacheDecorator) recordMiss() {
	if d.metrics != nil {
		d.metrics.RecordCacheMiss("company-redis", "tenantId")
	}
}
