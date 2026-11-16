package redisstore

import (
	"context"
	"errors"
	"time"

	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/redis/go-redis/v9"
)

// registerStateStore implementa port.RegisterStateStore sobre o Redis. Sem cliente, toda operação falha
// (quem usa segue sem trava, como antes da saga endurecida).
type registerStateStore struct {
	client *redis.Client
}

// NewRegisterStateStore monta o armazenamento da trava e do resultado do confirm de cadastro.
func NewRegisterStateStore(client *redis.Client) port.RegisterStateStore {
	return &registerStateStore{client: client}
}

var errRegisterStoreUnavailable = errors.New("redis indisponível para o estado do cadastro")

func (s *registerStateStore) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if s.client == nil {
		return false, errRegisterStoreUnavailable
	}
	return s.client.SetNX(ctx, key, "1", ttl).Result()
}

func (s *registerStateStore) Unlock(ctx context.Context, key string) error {
	if s.client == nil {
		return errRegisterStoreUnavailable
	}
	return s.client.Del(ctx, key).Err()
}

func (s *registerStateStore) GetResult(ctx context.Context, key string) ([]byte, bool, error) {
	if s.client == nil {
		return nil, false, errRegisterStoreUnavailable
	}
	raw, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

func (s *registerStateStore) SetResult(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if s.client == nil {
		return errRegisterStoreUnavailable
	}
	return s.client.Set(ctx, key, value, ttl).Err()
}
