package saga

import (
	"context"
	"time"
)

// Step representa um passo individual da SAGA
type Step struct {
	Name       string
	Execute    func(ctx context.Context, state *State) error
	Compensate func(ctx context.Context, state *State) error
	MaxRetries int
	Timeout    time.Duration
}

// State representa o estado atual da SAGA
type State struct {
	SagaID         string                 `json:"saga_id"`
	CurrentStep    int                    `json:"current_step"`
	CompletedSteps []string               `json:"completed_steps"`
	Status         SagaStatus             `json:"status"`
	Data           map[string]interface{} `json:"data"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// SagaStatus representa o status atual da SAGA
type SagaStatus string

const (
	StatusRunning      SagaStatus = "RUNNING"
	StatusCompleted    SagaStatus = "COMPLETED"
	StatusCompensating SagaStatus = "COMPENSATING"
	StatusCompensated  SagaStatus = "COMPENSATED"
	StatusFailed       SagaStatus = "FAILED"
)

// Saga representa uma transação SAGA completa
type Saga struct {
	State *State
	Steps []Step
}

// StateRepository interface para persistir estado da SAGA
type StateRepository interface {
	Save(ctx context.Context, state *State) error
	Get(ctx context.Context, sagaID string) (*State, error)
	Delete(ctx context.Context, sagaID string) error
}

// NewState cria um novo estado da SAGA
func NewState(sagaID string) *State {
	now := time.Now()
	return &State{
		SagaID:         sagaID,
		CurrentStep:    0,
		CompletedSteps: []string{},
		Status:         StatusRunning,
		Data:           make(map[string]interface{}),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UpdateTimestamp atualiza o timestamp de atualização
func (s *State) UpdateTimestamp() {
	s.UpdatedAt = time.Now()
}

// AddCompletedStep adiciona um step completado
func (s *State) AddCompletedStep(stepName string) {
	s.CompletedSteps = append(s.CompletedSteps, stepName)
	s.UpdateTimestamp()
}

// SetStatus define o status da SAGA
func (s *State) SetStatus(status SagaStatus) {
	s.Status = status
	s.UpdateTimestamp()
}

// IsCompleted verifica se a SAGA foi completada
func (s *State) IsCompleted() bool {
	return s.Status == StatusCompleted
}

// IsFailed verifica se a SAGA falhou
func (s *State) IsFailed() bool {
	return s.Status == StatusFailed
}

// IsCompensated verifica se a SAGA foi compensada
func (s *State) IsCompensated() bool {
	return s.Status == StatusCompensated
}

// IsRunning verifica se a SAGA está rodando
func (s *State) IsRunning() bool {
	return s.Status == StatusRunning
}

// IsCompensating verifica se a SAGA está compensando
func (s *State) IsCompensating() bool {
	return s.Status == StatusCompensating
}
