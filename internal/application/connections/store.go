package connections

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	snapshotKey = "bff-core:connections-health:snapshot"
	lockKey     = "bff-core:connections-health:lock"
)

// Snapshot é o payload cacheado da tela Conexões.
type Snapshot struct {
	CheckedAt  time.Time       `json:"checkedAt"`
	ExpiresAt  time.Time       `json:"expiresAt"`
	TTLSeconds int64           `json:"ttlSeconds"`
	Cached     bool            `json:"cached"`
	Services   []ServiceStatus `json:"services"`
}

// Store persiste o snapshot compartilhado.
type Store struct {
	redis *redis.Client
}

func NewStore(client *redis.Client) *Store {
	return &Store{redis: client}
}

func (s *Store) available() bool {
	return s != nil && s.redis != nil
}

// Get devolve o snapshot e o TTL restante.
func (s *Store) Get(ctx context.Context) (*Snapshot, time.Duration, error) {
	if !s.available() {
		return nil, 0, nil
	}
	raw, err := s.redis.Get(ctx, snapshotKey).Bytes()
	if err == redis.Nil {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	ttl, err := s.redis.TTL(ctx, snapshotKey).Result()
	if err != nil {
		ttl = 0
	}
	var snap Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return nil, 0, err
	}
	return &snap, ttl, nil
}

// Set grava o snapshot com TTL.
func (s *Store) Set(ctx context.Context, snap Snapshot, ttl time.Duration) error {
	if !s.available() {
		return nil
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, snapshotKey, raw, ttl).Err()
}

// TryLock tenta o lock de refresh (anti-stampede).
func (s *Store) TryLock(ctx context.Context, ttl time.Duration) (bool, error) {
	if !s.available() {
		return true, nil
	}
	return s.redis.SetNX(ctx, lockKey, "1", ttl).Result()
}

// Unlock libera o lock de refresh.
func (s *Store) Unlock(ctx context.Context) {
	if !s.available() {
		return
	}
	_ = s.redis.Del(ctx, lockKey).Err()
}
