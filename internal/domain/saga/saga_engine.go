package saga

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SagaEngine implementa a execução de SAGAs
type SagaEngine struct {
	stateRepo StateRepository
	logger    *zap.Logger
}

// NewSagaEngine cria uma nova instância do SagaEngine
func NewSagaEngine(stateRepo StateRepository, logger *zap.Logger) *SagaEngine {
	return &SagaEngine{
		stateRepo: stateRepo,
		logger:    logger,
	}
}

// Run executa uma SAGA completa
func (e *SagaEngine) Run(ctx context.Context, saga *Saga) error {
	e.logger.Info("Iniciando execução da SAGA",
		zap.String("saga_id", saga.State.SagaID),
		zap.Int("total_steps", len(saga.Steps)))

	// Salvar estado inicial
	if err := e.stateRepo.Save(ctx, saga.State); err != nil {
		return fmt.Errorf("falha ao salvar estado inicial: %w", err)
	}

	// Executar cada step sequencialmente
	for i, step := range saga.Steps {
		e.logger.Info("Executando step",
			zap.String("saga_id", saga.State.SagaID),
			zap.String("step_name", step.Name),
			zap.Int("step_index", i))

		// Atualizar step atual
		saga.State.CurrentStep = i
		if err := e.stateRepo.Save(ctx, saga.State); err != nil {
			return fmt.Errorf("falha ao salvar estado antes do step %s: %w", step.Name, err)
		}

		// Executar step com retry
		if err := e.executeWithRetry(ctx, step, saga.State); err != nil {
			e.logger.Error("Step falhou após retries",
				zap.String("saga_id", saga.State.SagaID),
				zap.String("step_name", step.Name),
				zap.Error(err))

			// Iniciar compensação
			saga.State.SetStatus(StatusCompensating)
			if err := e.stateRepo.Save(ctx, saga.State); err != nil {
				e.logger.Error("Falha ao salvar estado de compensação",
					zap.String("saga_id", saga.State.SagaID),
					zap.Error(err))
			}

			// Executar compensação
			if compErr := e.compensate(ctx, saga, i); compErr != nil {
				// Compensação falhou - situação crítica
				saga.State.SetStatus(StatusFailed)
				if err := e.stateRepo.Save(ctx, saga.State); err != nil {
					e.logger.Error("Falha ao salvar estado de falha",
						zap.String("saga_id", saga.State.SagaID),
						zap.Error(err))
				}

				e.alertOps("SAGA_COMPENSATION_FAILED", saga.State.SagaID, compErr)
				return fmt.Errorf("compensação falhou: %w", compErr)
			}

			saga.State.SetStatus(StatusCompensated)
			if err := e.stateRepo.Save(ctx, saga.State); err != nil {
				e.logger.Error("Falha ao salvar estado compensado",
					zap.String("saga_id", saga.State.SagaID),
					zap.Error(err))
			}

			// Preservar o tipo do erro original (HTTPError, AppError, etc)
			return err
		}

		// Step executado com sucesso
		saga.State.AddCompletedStep(step.Name)
		if err := e.stateRepo.Save(ctx, saga.State); err != nil {
			e.logger.Error("Falha ao salvar estado após step",
				zap.String("saga_id", saga.State.SagaID),
				zap.String("step_name", step.Name),
				zap.Error(err))
		}

		e.logger.Info("Step executado com sucesso",
			zap.String("saga_id", saga.State.SagaID),
			zap.String("step_name", step.Name))
	}

	// Todos os steps completados
	saga.State.SetStatus(StatusCompleted)
	if err := e.stateRepo.Save(ctx, saga.State); err != nil {
		e.logger.Error("Falha ao salvar estado final",
			zap.String("saga_id", saga.State.SagaID),
			zap.Error(err))
	}

	e.logger.Info("SAGA completada com sucesso",
		zap.String("saga_id", saga.State.SagaID),
		zap.Int("completed_steps", len(saga.State.CompletedSteps)))

	return nil
}

// executeWithRetry executa um step com retry e backoff exponencial
func (e *SagaEngine) executeWithRetry(ctx context.Context, step Step, state *State) error {
	backoff := 1 * time.Second
	maxBackoff := 8 * time.Second
	var lastErr error

	for attempt := 1; attempt <= step.MaxRetries; attempt++ {
		e.logger.Debug("Tentativa de execução",
			zap.String("step_name", step.Name),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", step.MaxRetries))

		// Criar contexto com timeout para o step
		stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
		err := step.Execute(stepCtx, state)
		cancel()

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
func (e *SagaEngine) compensate(ctx context.Context, saga *Saga, failedStepIndex int) error {
	e.logger.Info("Iniciando compensação",
		zap.String("saga_id", saga.State.SagaID),
		zap.Int("failed_step_index", failedStepIndex),
		zap.Int("steps_to_compensate", failedStepIndex))

	// Compensar em ordem reversa (do último executado até o primeiro)
	for i := failedStepIndex - 1; i >= 0; i-- {
		step := saga.Steps[i]

		// Se o step não tem compensação (ex: consultas GET), pular
		if step.Compensate == nil {
			e.logger.Debug("Step não precisa de compensação",
				zap.String("saga_id", saga.State.SagaID),
				zap.String("step_name", step.Name))
			continue
		}

		e.logger.Info("Compensando step",
			zap.String("saga_id", saga.State.SagaID),
			zap.String("step_name", step.Name),
			zap.Int("step_index", i))

		// Criar step de compensação
		compensateStep := Step{
			Name:       step.Name + "_compensate",
			Execute:    step.Compensate,
			MaxRetries: step.MaxRetries,
			Timeout:    step.Timeout,
		}

		// Executar compensação com retry
		if err := e.executeWithRetry(ctx, compensateStep, saga.State); err != nil {
			e.logger.Error("Compensação falhou",
				zap.String("saga_id", saga.State.SagaID),
				zap.String("step_name", step.Name),
				zap.Error(err))

			// Compensação falhou - situação crítica
			e.alertOps("SAGA_STEP_COMPENSATION_FAILED", saga.State.SagaID, err)
			return fmt.Errorf("compensação falhou no step %s: %w", step.Name, err)
		}

		e.logger.Info("Step compensado com sucesso",
			zap.String("saga_id", saga.State.SagaID),
			zap.String("step_name", step.Name))
	}

	e.logger.Info("Compensação completa com sucesso",
		zap.String("saga_id", saga.State.SagaID))

	return nil
}

// alertOps envia alerta para operações (implementação básica)
func (e *SagaEngine) alertOps(alertType, sagaID string, err error) {
	e.logger.Error("ALERTA CRÍTICO - SAGA",
		zap.String("alert_type", alertType),
		zap.String("saga_id", sagaID),
		zap.Error(err))

	// TODO: Implementar integração com sistema de alertas (Slack, PagerDuty, etc.)
	// Por enquanto, apenas log estruturado
}
