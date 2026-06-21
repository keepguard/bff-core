package saga

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// SagaMetrics contém todas as métricas relacionadas ao SAGA
type SagaMetrics struct {
	// Contadores simples
	sagaExecutionsTotal    int64
	sagaSuccessTotal       int64
	sagaFailuresTotal      int64
	sagaCompensationsTotal int64
	activeSagasCount       int64

	// Contadores por step
	stepExecutions map[string]int64
	stepSuccess    map[string]int64
	stepFailures   map[string]int64
	stepRetries    map[string]int64

	// Tempos de execução
	sagaDurations []time.Duration
	stepDurations map[string][]time.Duration

	// Mutex para thread safety
	mu sync.RWMutex
}

// NewSagaMetrics cria uma nova instância de métricas SAGA
func NewSagaMetrics() *SagaMetrics {
	return &SagaMetrics{
		stepExecutions: make(map[string]int64),
		stepSuccess:    make(map[string]int64),
		stepFailures:   make(map[string]int64),
		stepRetries:    make(map[string]int64),
		stepDurations:  make(map[string][]time.Duration),
		sagaDurations:  make([]time.Duration, 0),
	}
}

// RecordSagaStart registra o início de uma execução de SAGA
func (m *SagaMetrics) RecordSagaStart(sagaName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sagaExecutionsTotal++
	m.activeSagasCount++
}

// RecordSagaEnd registra o fim de uma execução de SAGA
func (m *SagaMetrics) RecordSagaEnd(sagaName, status string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeSagasCount--
	m.sagaDurations = append(m.sagaDurations, duration)

	switch status {
	case "success":
		m.sagaSuccessTotal++
	case "failure":
		m.sagaFailuresTotal++
	case "compensated":
		m.sagaCompensationsTotal++
	}
}

// RecordStepStart registra o início de um step
func (m *SagaMetrics) RecordStepStart(sagaName, stepName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sagaName + "." + stepName
	m.stepExecutions[key]++
}

// RecordStepEnd registra o fim de um step
func (m *SagaMetrics) RecordStepEnd(sagaName, stepName, status string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := sagaName + "." + stepName
	m.stepDurations[key] = append(m.stepDurations[key], duration)

	switch status {
	case "success":
		m.stepSuccess[key]++
	case "failure":
		m.stepFailures[key]++
	}
}

// RecordStepRetry registra uma tentativa de retry de step
func (m *SagaMetrics) RecordStepRetry(sagaName, stepName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sagaName + "." + stepName
	m.stepRetries[key]++
}

// GetStats retorna estatísticas das métricas
func (m *SagaMetrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	successRate := float64(0)
	if m.sagaExecutionsTotal > 0 {
		successRate = float64(m.sagaSuccessTotal) / float64(m.sagaExecutionsTotal) * 100
	}

	// Calcular tempo médio de execução
	var avgDuration time.Duration
	if len(m.sagaDurations) > 0 {
		var total time.Duration
		for _, d := range m.sagaDurations {
			total += d
		}
		avgDuration = total / time.Duration(len(m.sagaDurations))
	}

	return map[string]interface{}{
		"total_executions":    m.sagaExecutionsTotal,
		"success_total":       m.sagaSuccessTotal,
		"failures_total":      m.sagaFailuresTotal,
		"compensations_total": m.sagaCompensationsTotal,
		"active_sagas":        m.activeSagasCount,
		"success_rate":        successRate,
		"avg_duration":        avgDuration.String(),
		"step_executions":     m.stepExecutions,
		"step_success":        m.stepSuccess,
		"step_failures":       m.stepFailures,
		"step_retries":        m.stepRetries,
	}
}

// Reset reseta todas as métricas
func (m *SagaMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sagaExecutionsTotal = 0
	m.sagaSuccessTotal = 0
	m.sagaFailuresTotal = 0
	m.sagaCompensationsTotal = 0
	m.activeSagasCount = 0

	m.stepExecutions = make(map[string]int64)
	m.stepSuccess = make(map[string]int64)
	m.stepFailures = make(map[string]int64)
	m.stepRetries = make(map[string]int64)
	m.stepDurations = make(map[string][]time.Duration)
	m.sagaDurations = make([]time.Duration, 0)
}

// SagaMonitor monitor de SAGA com métricas e logs
type SagaMonitor struct {
	metrics *SagaMetrics
	logger  *zap.Logger
}

// NewSagaMonitor cria um novo monitor de SAGA
func NewSagaMonitor(logger *zap.Logger) *SagaMonitor {
	return &SagaMonitor{
		metrics: NewSagaMetrics(),
		logger:  logger,
	}
}

// MonitorSagaExecution monitora uma execução completa de SAGA
func (sm *SagaMonitor) MonitorSagaExecution(sagaName string, fn func() error) error {
	start := time.Now()
	sm.metrics.RecordSagaStart(sagaName)

	sm.logger.Info("SAGA execution started",
		zap.String("saga_name", sagaName),
		zap.Time("start_time", start))

	err := fn()
	duration := time.Since(start)

	var status string
	if err != nil {
		status = "failure"
		sm.logger.Error("SAGA execution failed",
			zap.String("saga_name", sagaName),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		status = "success"
		sm.logger.Info("SAGA execution completed successfully",
			zap.String("saga_name", sagaName),
			zap.Duration("duration", duration))
	}

	sm.metrics.RecordSagaEnd(sagaName, status, duration)
	return err
}

// MonitorStepExecution monitora a execução de um step individual
func (sm *SagaMonitor) MonitorStepExecution(sagaName, stepName string, fn func() error) error {
	start := time.Now()
	sm.metrics.RecordStepStart(sagaName, stepName)

	err := fn()
	duration := time.Since(start)

	var status string
	if err != nil {
		status = "failure"
		sm.logger.Error("SAGA step failed",
			zap.String("saga_name", sagaName),
			zap.String("step_name", stepName),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		status = "success"
		sm.logger.Debug("SAGA step completed successfully",
			zap.String("saga_name", sagaName),
			zap.String("step_name", stepName),
			zap.Duration("duration", duration))
	}

	sm.metrics.RecordStepEnd(sagaName, stepName, status, duration)
	return err
}

// RecordCompensation registra uma compensação executada
func (sm *SagaMonitor) RecordCompensation(sagaName, stepName string) {
	sm.metrics.mu.Lock()
	sm.metrics.sagaCompensationsTotal++
	sm.metrics.mu.Unlock()

	sm.logger.Warn("SAGA compensation executed",
		zap.String("saga_name", sagaName),
		zap.String("step_name", stepName))
}

// RecordRetry registra uma tentativa de retry
func (sm *SagaMonitor) RecordRetry(sagaName, stepName string, attempt int) {
	sm.metrics.RecordStepRetry(sagaName, stepName)
	sm.logger.Warn("SAGA step retry",
		zap.String("saga_name", sagaName),
		zap.String("step_name", stepName),
		zap.Int("attempt", attempt))
}
