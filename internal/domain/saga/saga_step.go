package saga

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// StepBuilder facilita a construção de steps
type StepBuilder struct {
	step *Step
}

// NewStepBuilder cria um novo builder de step
func NewStepBuilder(name string) *StepBuilder {
	return &StepBuilder{
		step: &Step{
			Name:       name,
			MaxRetries: 3,
			Timeout:    10 * time.Second,
		},
	}
}

// WithExecute define a função de execução
func (sb *StepBuilder) WithExecute(execute func(ctx context.Context, state *State) error) *StepBuilder {
	sb.step.Execute = execute
	return sb
}

// WithCompensate define a função de compensação
func (sb *StepBuilder) WithCompensate(compensate func(ctx context.Context, state *State) error) *StepBuilder {
	sb.step.Compensate = compensate
	return sb
}

// WithMaxRetries define o número máximo de tentativas
func (sb *StepBuilder) WithMaxRetries(maxRetries int) *StepBuilder {
	sb.step.MaxRetries = maxRetries
	return sb
}

// WithTimeout define o timeout do step
func (sb *StepBuilder) WithTimeout(timeout time.Duration) *StepBuilder {
	sb.step.Timeout = timeout
	return sb
}

// Build constrói o step
func (sb *StepBuilder) Build() Step {
	return *sb.step
}

// StepValidator valida steps antes da execução
type StepValidator struct{}

// NewStepValidator cria um novo validador de steps
func NewStepValidator() *StepValidator {
	return &StepValidator{}
}

// Validate valida um step
func (sv *StepValidator) Validate(step Step) error {
	if step.Name == "" {
		return fmt.Errorf("nome do step não pode ser vazio")
	}

	if step.Execute == nil {
		return fmt.Errorf("função de execução é obrigatória")
	}

	if step.MaxRetries <= 0 {
		return fmt.Errorf("número máximo de tentativas deve ser maior que zero")
	}

	if step.Timeout <= 0 {
		return fmt.Errorf("timeout deve ser maior que zero")
	}

	return nil
}

// ValidateSaga valida todos os steps de uma SAGA
func (sv *StepValidator) ValidateSaga(saga *Saga) error {
	if saga == nil {
		return fmt.Errorf("SAGA não pode ser nil")
	}

	if saga.State == nil {
		return fmt.Errorf("estado da SAGA não pode ser nil")
	}

	if len(saga.Steps) == 0 {
		return fmt.Errorf("SAGA deve ter pelo menos um step")
	}

	for i, step := range saga.Steps {
		if err := sv.Validate(step); err != nil {
			return fmt.Errorf("step %d (%s) inválido: %w", i, step.Name, err)
		}
	}

	return nil
}

// StepExecutor executa steps individuais
type StepExecutor struct {
	logger *zap.Logger
}

// NewStepExecutor cria um novo executor de steps
func NewStepExecutor(logger *zap.Logger) *StepExecutor {
	return &StepExecutor{
		logger: logger,
	}
}

// Execute executa um step individual
func (se *StepExecutor) Execute(ctx context.Context, step Step, state *State) error {
	se.logger.Debug("Executando step",
		zap.String("step_name", step.Name),
		zap.String("saga_id", state.SagaID))

	// Criar contexto com timeout
	stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
	defer cancel()

	// Executar step
	err := step.Execute(stepCtx, state)
	if err != nil {
		se.logger.Error("Step falhou",
			zap.String("step_name", step.Name),
			zap.String("saga_id", state.SagaID),
			zap.Error(err))
		return err
	}

	se.logger.Debug("Step executado com sucesso",
		zap.String("step_name", step.Name),
		zap.String("saga_id", state.SagaID))

	return nil
}

// Compensate executa a compensação de um step
func (se *StepExecutor) Compensate(ctx context.Context, step Step, state *State) error {
	if step.Compensate == nil {
		se.logger.Debug("Step não tem compensação",
			zap.String("step_name", step.Name),
			zap.String("saga_id", state.SagaID))
		return nil
	}

	se.logger.Debug("Compensando step",
		zap.String("step_name", step.Name),
		zap.String("saga_id", state.SagaID))

	// Criar contexto com timeout
	stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
	defer cancel()

	// Executar compensação
	err := step.Compensate(stepCtx, state)
	if err != nil {
		se.logger.Error("Compensação falhou",
			zap.String("step_name", step.Name),
			zap.String("saga_id", state.SagaID),
			zap.Error(err))
		return err
	}

	se.logger.Debug("Step compensado com sucesso",
		zap.String("step_name", step.Name),
		zap.String("saga_id", state.SagaID))

	return nil
}

// StepMetrics coleta métricas de steps
type StepMetrics struct {
	StepName     string
	Duration     time.Duration
	Success      bool
	RetryCount   int
	ErrorMessage string
}

// StepMetricsCollector coleta métricas de execução
type StepMetricsCollector struct {
	metrics []StepMetrics
}

// NewStepMetricsCollector cria um novo coletor de métricas
func NewStepMetricsCollector() *StepMetricsCollector {
	return &StepMetricsCollector{
		metrics: make([]StepMetrics, 0),
	}
}

// RecordStep registra métricas de um step
func (smc *StepMetricsCollector) RecordStep(stepName string, duration time.Duration, success bool, retryCount int, err error) {
	metric := StepMetrics{
		StepName:   stepName,
		Duration:   duration,
		Success:    success,
		RetryCount: retryCount,
	}

	if err != nil {
		metric.ErrorMessage = err.Error()
	}

	smc.metrics = append(smc.metrics, metric)
}

// GetMetrics retorna as métricas coletadas
func (smc *StepMetricsCollector) GetMetrics() []StepMetrics {
	return smc.metrics
}

// GetTotalDuration retorna a duração total de todos os steps
func (smc *StepMetricsCollector) GetTotalDuration() time.Duration {
	var total time.Duration
	for _, metric := range smc.metrics {
		total += metric.Duration
	}
	return total
}

// GetSuccessRate retorna a taxa de sucesso dos steps
func (smc *StepMetricsCollector) GetSuccessRate() float64 {
	if len(smc.metrics) == 0 {
		return 0.0
	}

	successCount := 0
	for _, metric := range smc.metrics {
		if metric.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(smc.metrics))
}
