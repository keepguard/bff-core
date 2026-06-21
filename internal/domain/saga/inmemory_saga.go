package saga

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// InMemoryStep representa um passo individual da SAGA em memória
type InMemoryStep struct {
	Name       string
	Execute    func(ctx context.Context, data map[string]interface{}) error
	Compensate func(ctx context.Context, data map[string]interface{}) error
	MaxRetries int
	Timeout    time.Duration
}

// InMemorySaga representa uma transação SAGA em memória
type InMemorySaga struct {
	Name  string
	Steps []InMemoryStep
}

// InMemorySagaExecutor executa SAGAs em memória
type InMemorySagaExecutor struct {
	logger  *zap.Logger
	monitor *SagaMonitor
}

// NewInMemorySagaExecutor cria um novo executor de SAGA em memória
func NewInMemorySagaExecutor(logger *zap.Logger) *InMemorySagaExecutor {
	return &InMemorySagaExecutor{
		logger:  logger,
		monitor: NewSagaMonitor(logger),
	}
}

// Execute executa uma SAGA em memória
func (e *InMemorySagaExecutor) Execute(ctx context.Context, saga InMemorySaga, data map[string]interface{}) error {
	return e.monitor.MonitorSagaExecution(saga.Name, func() error {
		e.logger.Info("Iniciando execução da SAGA em memória",
			zap.String("saga_name", saga.Name),
			zap.Int("total_steps", len(saga.Steps)))

		executedSteps := make([]InMemoryStep, 0)

		// Executar cada step sequencialmente
		for i, step := range saga.Steps {
			e.logger.Info("Executando step",
				zap.String("saga_name", saga.Name),
				zap.String("step_name", step.Name),
				zap.Int("step_index", i))

			// Executar step com retry
			if err := e.executeWithRetry(ctx, step, data); err != nil {
				e.logger.Error("Step falhou após retries",
					zap.String("saga_name", saga.Name),
					zap.String("step_name", step.Name),
					zap.Error(err))

				// Executar compensação
				if compErr := e.compensate(ctx, executedSteps, data); compErr != nil {
					e.logger.Error("Compensação falhou",
						zap.String("saga_name", saga.Name),
						zap.Error(compErr))
					return fmt.Errorf("saga falhou e compensação falhou: %w", compErr)
				}

				// Preservar o tipo do erro original (HTTPError, AppError, etc)
				return err
			}

			// Step executado com sucesso
			executedSteps = append(executedSteps, step)
			e.logger.Info("Step executado com sucesso",
				zap.String("saga_name", saga.Name),
				zap.String("step_name", step.Name))
		}

		e.logger.Info("SAGA completada com sucesso",
			zap.String("saga_name", saga.Name),
			zap.Int("completed_steps", len(executedSteps)))

		return nil
	})
}

// executeWithRetry executa um step com retry e backoff exponencial
func (e *InMemorySagaExecutor) executeWithRetry(ctx context.Context, step InMemoryStep, data map[string]interface{}) error {
	backoff := 1 * time.Second
	maxBackoff := 8 * time.Second
	var lastErr error

	for attempt := 1; attempt <= step.MaxRetries; attempt++ {
		e.logger.Debug("Tentativa de execução",
			zap.String("step_name", step.Name),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", step.MaxRetries))

		// Criar contexto com timeout para o step
		stepCtx := ctx
		if step.Timeout > 0 {
			var cancel context.CancelFunc
			stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
			defer cancel()
		}

		// Executar step com monitoramento
		err := e.monitor.MonitorStepExecution("RegisterConfirmSaga", step.Name, func() error {
			return step.Execute(stepCtx, data)
		})

		if err == nil {
			e.logger.Debug("Step executado com sucesso",
				zap.String("step_name", step.Name),
				zap.Int("attempt", attempt))
			return nil
		}

		// Armazenar o último erro para preservar o tipo original
		lastErr = err

		e.logger.Warn("Step falhou, tentando novamente",
			zap.String("step_name", step.Name),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", step.MaxRetries),
			zap.Error(err))

		// Registrar retry
		e.monitor.RecordRetry("RegisterConfirmSaga", step.Name, attempt)

		// Se não é a última tentativa, aguardar antes de tentar novamente
		if attempt < step.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				// Continuar para próxima tentativa
			}

			// Aumentar backoff exponencialmente, mas com limite máximo
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	// Preservar o tipo do erro original (HTTPError, AppError, etc)
	// Se chegou aqui, significa que todas as tentativas falharam
	// Retornamos o último erro que ocorreu
	return lastErr
}

// compensate executa a compensação em ordem reversa
func (e *InMemorySagaExecutor) compensate(ctx context.Context, executedSteps []InMemoryStep, data map[string]interface{}) error {
	e.logger.Info("Iniciando compensação",
		zap.Int("steps_to_compensate", len(executedSteps)))

	// Compensar em ordem reversa (do último executado até o primeiro)
	for i := len(executedSteps) - 1; i >= 0; i-- {
		step := executedSteps[i]

		// Se o step não tem compensação, pular
		if step.Compensate == nil {
			e.logger.Debug("Step não precisa de compensação",
				zap.String("step_name", step.Name))
			continue
		}

		e.logger.Info("Compensando step",
			zap.String("step_name", step.Name))

		// Criar contexto com timeout para a compensação
		stepCtx := ctx
		if step.Timeout > 0 {
			var cancel context.CancelFunc
			stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
			defer cancel()
		}

		// Executar compensação
		if err := step.Compensate(stepCtx, data); err != nil {
			e.logger.Error("Compensação falhou",
				zap.String("step_name", step.Name),
				zap.Error(err))
			return fmt.Errorf("compensação falhou no step %s: %w", step.Name, err)
		}

		// Registrar compensação executada
		e.monitor.RecordCompensation("RegisterConfirmSaga", step.Name)

		e.logger.Info("Step compensado com sucesso",
			zap.String("step_name", step.Name))
	}

	e.logger.Info("Compensação completa com sucesso")
	return nil
}
