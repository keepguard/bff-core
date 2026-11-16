package redisstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sem cliente, toda operação falha de forma explícita (nunca pânico).
func TestRegisterStateStore_NilClientFailsExplicitly(t *testing.T) {
	s := NewRegisterStateStore(nil)
	ctx := context.Background()

	_, err := s.TryLock(ctx, "k", time.Second)
	assert.Error(t, err)
	assert.Error(t, s.Unlock(ctx, "k"))
	_, _, err = s.GetResult(ctx, "k")
	assert.Error(t, err)
	assert.Error(t, s.SetResult(ctx, "k", []byte("v"), time.Second))
}

// Contra Redis real (local e descartável): SETNX com prazo, soltar, guardar e ler resultado.
// Rodar: BFF_CORE_TEST_REDIS_ADDR=127.0.0.1:6390 go test ./internal/adapters/out/redis/ -run RegisterStateStore
func TestRegisterStateStore_RedisSemantics(t *testing.T) {
	addr := os.Getenv("BFF_CORE_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("defina BFF_CORE_TEST_REDIS_ADDR para rodar contra um Redis local")
	}
	client := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	defer client.Close()
	ctx := context.Background()
	require.NoError(t, client.Ping(ctx).Err())

	s := NewRegisterStateStore(client)
	base := fmt.Sprintf("kg:bff-core:test:register:%d", time.Now().UnixNano())
	lockKey, resKey := base+":lock", base+":result"
	t.Cleanup(func() { client.Del(context.Background(), lockKey, resKey) })

	got, err := s.TryLock(ctx, lockKey, 5*time.Second)
	require.NoError(t, err)
	assert.True(t, got, "primeira trava deve ser concedida")

	got, err = s.TryLock(ctx, lockKey, 5*time.Second)
	require.NoError(t, err)
	assert.False(t, got, "segunda trava com a primeira presa deve falhar")

	require.NoError(t, s.Unlock(ctx, lockKey))
	got, err = s.TryLock(ctx, lockKey, 5*time.Second)
	require.NoError(t, err)
	assert.True(t, got, "depois de soltar, a trava volta a ser concedida")

	_, found, err := s.GetResult(ctx, resKey)
	require.NoError(t, err)
	assert.False(t, found)

	require.NoError(t, s.SetResult(ctx, resKey, []byte("sealed"), 10*time.Second))
	val, found, err := s.GetResult(ctx, resKey)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, []byte("sealed"), val)

	ttl, err := client.TTL(ctx, resKey).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0), "resultado precisa de prazo")
	assert.LessOrEqual(t, ttl, 10*time.Second)

	lockTTL, err := client.TTL(ctx, lockKey).Result()
	require.NoError(t, err)
	assert.Greater(t, lockTTL, time.Duration(0), "trava precisa de prazo")
}
