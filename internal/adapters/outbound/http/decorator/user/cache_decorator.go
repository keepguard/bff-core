package user

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	authDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/user"
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
	value     userDto.MSUserResponseDTO
	expiresAt time.Time
}

// cacheDecorator implementa cache para UserClient
type cacheDecorator struct {
	inner       portsclient.UserClient
	cache       map[string]*cacheEntry
	mutex       sync.RWMutex
	config      CacheConfig
	metrics     *metrics.Metrics
	stopCleanup chan bool
}

// NewCacheDecorator cria um decorator de cache para UserClient
func NewCacheDecorator(
	inner portsclient.UserClient,
	config CacheConfig,
	metrics *metrics.Metrics,
) portsclient.UserClient {
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

// GetUserByCodeUser implementa GetUserByCodeUser com cache
func (d *cacheDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	// Gera chave do cache
	cacheKey := d.generateCacheKey("codeUser", codeUser)

	// Tenta buscar no cache
	d.mutex.RLock()
	if entry, exists := d.cache[cacheKey]; exists {
		if time.Now().Before(entry.expiresAt) {
			// Cache hit
			d.mutex.RUnlock()
			d.metrics.RecordCacheHit("user", "codeUser")
			return entry.value, nil
		}
		// Cache expirado, remove
		delete(d.cache, cacheKey)
	}
	d.mutex.RUnlock()

	// Cache miss - busca no serviço
	d.metrics.RecordCacheMiss("user", "codeUser")
	response, err := d.inner.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
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

// CreateUser implementa CreateUser (sem cache)
func (d *cacheDecorator) CreateUser(ctx context.Context, req userDto.MSUserCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserResponseDTO, error) {
	return d.inner.CreateUser(ctx, req, tenantId, correlationID)
}

// GetByEmail implementa GetByEmail (sem cache)
func (d *cacheDecorator) GetByEmail(ctx context.Context, email, tenantId, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	return d.inner.GetByEmail(ctx, email, tenantId, correlationID)
}

// CreateUserNotify implementa CreateUserNotify (sem cache)
func (d *cacheDecorator) CreateUserNotify(ctx context.Context, req userDto.MSUserNotifyCreateRequestDTO, tenantId, correlationID string) (userDto.MSUserNotifyResponseDTO, error) {
	return d.inner.CreateUserNotify(ctx, req, tenantId, correlationID)
}

// InitRegister implementa InitRegister (sem cache)
func (d *cacheDecorator) InitRegister(ctx context.Context, req userDto.MSUserRegisterInitRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterInitResponseDTO, error) {
	return d.inner.InitRegister(ctx, req, tenantId, correlationID)
}

// ConfirmRegister implementa ConfirmRegister (sem cache)
func (d *cacheDecorator) ConfirmRegister(ctx context.Context, req userDto.MSUserRegisterConfirmRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterConfirmResponseDTO, error) {
	return d.inner.ConfirmRegister(ctx, req, tenantId, correlationID)
}

// DeleteUser implementa DeleteUser (sem cache)
func (d *cacheDecorator) DeleteUser(ctx context.Context, userID, tenantId, correlationID string) error {
	return d.inner.DeleteUser(ctx, userID, tenantId, correlationID)
}

// generateCacheKey gera uma chave única para o cache
func (d *cacheDecorator) generateCacheKey(keyType, key string) string {
	hash := md5.Sum([]byte(keyType + ":" + key))
	return fmt.Sprintf("user:%s:%x", keyType, hash)
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

// ResendRegisterToken implementa ResendRegisterToken (sem cache)
func (d *cacheDecorator) ResendRegisterToken(ctx context.Context, req userDto.MSUserRegisterResendRequestDTO, tenantId, correlationID string) (userDto.MSUserRegisterResendResponseDTO, error) {
	return d.inner.ResendRegisterToken(ctx, req, tenantId, correlationID)
}

// Stop para o processo de cleanup
func (d *cacheDecorator) Stop() {
	close(d.stopCleanup)
}
