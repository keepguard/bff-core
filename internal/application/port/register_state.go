package port

import (
	"context"
	"time"
)

// RegisterStateStore guarda o estado do confirm de cadastro no Redis: a trava da saga e o resultado já
// concluído (para o retry do mesmo confirm). Só Redis; nada em memória do processo. Chaves com prazo.
type RegisterStateStore interface {
	// TryLock grava a trava se ela não existe (SETNX com prazo). Devolve false se já está presa.
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	// Unlock solta a trava.
	Unlock(ctx context.Context, key string) error
	// GetResult devolve o valor guardado e true, ou false se não há resultado (ou expirou).
	GetResult(ctx context.Context, key string) ([]byte, bool, error)
	// SetResult guarda o valor com prazo.
	SetResult(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

// RegistrationMetrics registra os passos e as compensações da saga de cadastro.
// result: ok | error | ambiguous_recovered | not_found | replayed | in_progress | store_error (conforme o passo).
type RegistrationMetrics interface {
	RecordStep(step, result string)
	RecordCompensation(step, result string)
}
