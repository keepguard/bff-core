package company

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
)

// CacheConfig configuração do cache
type CacheConfig struct {
	TTL             time.Duration
	MaxSize         int
	CleanupInterval time.Duration
}

// cacheEntry representa uma entrada no cache
type cacheEntry struct {
	value     companyDto.MSCompanyResponseDTO
	expiresAt time.Time
}

// cacheDecorator implementa cache para CompanyClient
type cacheDecorator struct {
	inner       portsclient.CompanyClient
	cache       map[string]*cacheEntry
	mutex       sync.RWMutex
	config      CacheConfig
	metrics     *metrics.Metrics
	stopCleanup chan bool
}

// NewCacheDecorator cria um decorator de cache para CompanyClient
func NewCacheDecorator(
	inner portsclient.CompanyClient,
	config CacheConfig,
	metrics *metrics.Metrics,
) portsclient.CompanyClient {
	decorator := &cacheDecorator{
		inner:       inner,
		cache:       make(map[string]*cacheEntry),
		config:      config,
		metrics:     metrics,
		stopCleanup: make(chan bool),
	}

	// Inicia cleanup automático
	go decorator.startCleanup()

	return decorator
}

// GetByTenantId implementa GetByTenantId com cache
func (d *cacheDecorator) GetByTenantId(ctx context.Context, tenantId, correlationID string) (companyDto.MSCompanyResponseDTO, error) {
	// Gera chave do cache
	cacheKey := d.generateCacheKey(tenantId)

	// Tenta buscar no cache
	d.mutex.RLock()
	if entry, exists := d.cache[cacheKey]; exists {
		if time.Now().Before(entry.expiresAt) {
			// Cache hit
			d.mutex.RUnlock()
			d.metrics.RecordCacheHit("company", "tenantId")
			return entry.value, nil
		}
		// Cache expirado, remove
		delete(d.cache, cacheKey)
	}
	d.mutex.RUnlock()

	// Cache miss - busca no serviço
	d.metrics.RecordCacheMiss("company", "tenantId")
	response, err := d.inner.GetByTenantId(ctx, tenantId, correlationID)
	if err != nil {
		return response, err
	}

	// Armazena no cache
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Verifica se não excedeu o tamanho máximo
	if len(d.cache) >= d.config.MaxSize {
		d.evictOldest()
	}

	d.cache[cacheKey] = &cacheEntry{
		value:     response,
		expiresAt: time.Now().Add(d.config.TTL),
	}

	return response, nil
}

// generateCacheKey gera uma chave única para o cache
func (d *cacheDecorator) generateCacheKey(tenantId string) string {
	hash := md5.Sum([]byte(tenantId))
	return fmt.Sprintf("company:tenantId:%x", hash)
}

// evictOldest remove a entrada mais antiga do cache
func (d *cacheDecorator) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range d.cache {
		if oldestKey == "" || entry.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiresAt
		}
	}

	if oldestKey != "" {
		delete(d.cache, oldestKey)
	}
}

// startCleanup inicia o processo de limpeza automática do cache
func (d *cacheDecorator) startCleanup() {
	ticker := time.NewTicker(d.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.cleanup()
		case <-d.stopCleanup:
			return
		}
	}
}

// cleanup remove entradas expiradas do cache
func (d *cacheDecorator) cleanup() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := time.Now()
	for key, entry := range d.cache {
		if now.After(entry.expiresAt) {
			delete(d.cache, key)
		}
	}
}

// Stop para o processo de cleanup
func (d *cacheDecorator) Stop() {
	close(d.stopCleanup)
}
