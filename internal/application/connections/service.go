package connections

import (
	"context"
	"time"

	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

// Service orquestra cache Redis + probes.
type Service struct {
	cfg     config.ConnectionsHealthConfig
	store   *Store
	targets []Target
	logger  *zap.Logger
}

func NewService(cfg config.ConnectionsHealthConfig, store *Store, logger *zap.Logger) *Service {
	if cfg.SnapshotTTL <= 0 {
		cfg.SnapshotTTL = 60 * time.Second
	}
	if cfg.LockTTL <= 0 {
		cfg.LockTTL = 5 * time.Second
	}
	if cfg.ProbeTimeout <= 0 {
		cfg.ProbeTimeout = 500 * time.Millisecond
	}
	return &Service{
		cfg:     cfg,
		store:   store,
		targets: Targets(cfg),
		logger:  logger,
	}
}

// WithTargets restringe os probes (testes e overrides pontuais).
func (s *Service) WithTargets(targets []Target) *Service {
	clone := *s
	clone.targets = targets
	return &clone
}

// GetSnapshot devolve cache ou coleta um snapshot novo.
func (s *Service) GetSnapshot(ctx context.Context) (Snapshot, error) {
	if snap, ok := s.fromCache(ctx); ok {
		return snap, nil
	}

	gotLock, err := s.store.TryLock(ctx, s.cfg.LockTTL)
	if err != nil {
		s.logger.Warn("Falha ao obter lock do snapshot de conexões", zap.Error(err))
		gotLock = true
	}
	if !gotLock {
		if snap := s.waitForCache(ctx); snap != nil {
			return *snap, nil
		}
	} else {
		defer s.store.Unlock(ctx)
	}

	return s.refresh(ctx), nil
}

func (s *Service) fromCache(ctx context.Context) (Snapshot, bool) {
	snap, ttl, err := s.store.Get(ctx)
	if err != nil {
		s.logger.Warn("Falha ao ler snapshot de conexões", zap.Error(err))
		return Snapshot{}, false
	}
	if snap == nil || ttl <= 0 {
		return Snapshot{}, false
	}
	return decorateCached(*snap, ttl), true
}

func (s *Service) waitForCache(ctx context.Context) *Snapshot {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(100 * time.Millisecond):
		}
		if snap, ok := s.fromCache(ctx); ok {
			return &snap
		}
	}
	return nil
}

func (s *Service) refresh(ctx context.Context) Snapshot {
	checkedAt := time.Now().UTC()
	services := ProbeAll(ctx, s.targets, s.cfg.ProbeTimeout)
	snap := Snapshot{
		CheckedAt:  checkedAt,
		ExpiresAt:  checkedAt.Add(s.cfg.SnapshotTTL),
		TTLSeconds: int64(s.cfg.SnapshotTTL / time.Second),
		Cached:     false,
		Services:   services,
	}
	if err := s.store.Set(ctx, snap, s.cfg.SnapshotTTL); err != nil {
		s.logger.Warn("Falha ao gravar snapshot de conexões", zap.Error(err))
	}
	return snap
}

func decorateCached(snap Snapshot, ttl time.Duration) Snapshot {
	now := time.Now().UTC()
	seconds := int64(ttl / time.Second)
	if seconds < 0 {
		seconds = 0
	}
	snap.Cached = true
	snap.TTLSeconds = seconds
	snap.ExpiresAt = now.Add(ttl)
	return snap
}
