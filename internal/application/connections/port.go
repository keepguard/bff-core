package connections

import "context"

// ConnectionsHealthPort é a porta de entrada do snapshot da tela Conexões.
type ConnectionsHealthPort interface {
	GetSnapshot(ctx context.Context) (Snapshot, error)
}
