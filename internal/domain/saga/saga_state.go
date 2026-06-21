package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// StateManager gerencia o estado da SAGA
type StateManager struct {
	repo   StateRepository
	logger *zap.Logger
}

// NewStateManager cria um novo gerenciador de estado
func NewStateManager(repo StateRepository, logger *zap.Logger) *StateManager {
	return &StateManager{
		repo:   repo,
		logger: logger,
	}
}

// SaveState salva o estado da SAGA
func (sm *StateManager) SaveState(ctx context.Context, state *State) error {
	state.UpdateTimestamp()

	if err := sm.repo.Save(ctx, state); err != nil {
		sm.logger.Error("Falha ao salvar estado da SAGA",
			zap.String("saga_id", state.SagaID),
			zap.Error(err))
		return fmt.Errorf("falha ao salvar estado: %w", err)
	}

	sm.logger.Debug("Estado da SAGA salvo",
		zap.String("saga_id", state.SagaID),
		zap.String("status", string(state.Status)),
		zap.Int("current_step", state.CurrentStep))

	return nil
}

// LoadState carrega o estado da SAGA
func (sm *StateManager) LoadState(ctx context.Context, sagaID string) (*State, error) {
	state, err := sm.repo.Get(ctx, sagaID)
	if err != nil {
		sm.logger.Error("Falha ao carregar estado da SAGA",
			zap.String("saga_id", sagaID),
			zap.Error(err))
		return nil, fmt.Errorf("falha ao carregar estado: %w", err)
	}

	sm.logger.Debug("Estado da SAGA carregado",
		zap.String("saga_id", sagaID),
		zap.String("status", string(state.Status)),
		zap.Int("current_step", state.CurrentStep))

	return state, nil
}

// DeleteState remove o estado da SAGA
func (sm *StateManager) DeleteState(ctx context.Context, sagaID string) error {
	if err := sm.repo.Delete(ctx, sagaID); err != nil {
		sm.logger.Error("Falha ao deletar estado da SAGA",
			zap.String("saga_id", sagaID),
			zap.Error(err))
		return fmt.Errorf("falha ao deletar estado: %w", err)
	}

	sm.logger.Debug("Estado da SAGA deletado",
		zap.String("saga_id", sagaID))

	return nil
}

// ResumeSaga retoma uma SAGA existente
func (sm *StateManager) ResumeSaga(ctx context.Context, sagaID string, steps []Step) (*Saga, error) {
	state, err := sm.LoadState(ctx, sagaID)
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar SAGA para retomada: %w", err)
	}

	// Verificar se a SAGA pode ser retomada
	if state.IsCompleted() {
		return nil, fmt.Errorf("SAGA já foi completada: %s", sagaID)
	}

	if state.IsFailed() {
		return nil, fmt.Errorf("SAGA falhou e não pode ser retomada: %s", sagaID)
	}

	if state.IsCompensated() {
		return nil, fmt.Errorf("SAGA foi compensada e não pode ser retomada: %s", sagaID)
	}

	// Criar SAGA com estado carregado
	saga := &Saga{
		State: state,
		Steps: steps,
	}

	sm.logger.Info("SAGA retomada",
		zap.String("saga_id", sagaID),
		zap.String("status", string(state.Status)),
		zap.Int("current_step", state.CurrentStep),
		zap.Int("completed_steps", len(state.CompletedSteps)))

	return saga, nil
}

// GetSagaStatus retorna o status atual da SAGA
func (sm *StateManager) GetSagaStatus(ctx context.Context, sagaID string) (SagaStatus, error) {
	state, err := sm.LoadState(ctx, sagaID)
	if err != nil {
		return "", fmt.Errorf("falha ao obter status da SAGA: %w", err)
	}

	return state.Status, nil
}

// ListActiveSagas retorna lista de SAGAs ativas (em execução ou compensando)
func (sm *StateManager) ListActiveSagas(ctx context.Context) ([]string, error) {
	// TODO: Implementar busca por SAGAs ativas no repositório
	// Por enquanto, retornar lista vazia
	return []string{}, nil
}

// CleanupExpiredSagas remove SAGAs expiradas
func (sm *StateManager) CleanupExpiredSagas(ctx context.Context, maxAge time.Duration) error {
	// TODO: Implementar limpeza de SAGAs expiradas
	// Por enquanto, apenas log
	sm.logger.Info("Limpeza de SAGAs expiradas não implementada",
		zap.Duration("max_age", maxAge))
	return nil
}

// MarshalState serializa o estado para JSON
func (sm *StateManager) MarshalState(state *State) ([]byte, error) {
	return json.Marshal(state)
}

// UnmarshalState deserializa o estado de JSON
func (sm *StateManager) UnmarshalState(data []byte) (*State, error) {
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("falha ao deserializar estado: %w", err)
	}
	return &state, nil
}

// ValidateState valida se o estado está consistente
func (sm *StateManager) ValidateState(state *State) error {
	if state.SagaID == "" {
		return fmt.Errorf("SagaID não pode ser vazio")
	}

	if state.Status == "" {
		return fmt.Errorf("Status não pode ser vazio")
	}

	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}

	if state.CompletedSteps == nil {
		state.CompletedSteps = []string{}
	}

	// Validar status
	switch state.Status {
	case StatusRunning, StatusCompleted, StatusCompensating, StatusCompensated, StatusFailed:
		// Status válido
	default:
		return fmt.Errorf("Status inválido: %s", state.Status)
	}

	return nil
}
