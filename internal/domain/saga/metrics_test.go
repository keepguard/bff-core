package saga

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSagaMetrics(t *testing.T) {
	// Act
	metrics := NewSagaMetrics()

	// Assert
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.sagaExecutionsTotal)
	assert.Equal(t, int64(0), metrics.sagaSuccessTotal)
	assert.Equal(t, int64(0), metrics.sagaFailuresTotal)
	assert.Equal(t, int64(0), metrics.sagaCompensationsTotal)
	assert.Equal(t, int64(0), metrics.activeSagasCount)
	assert.NotNil(t, metrics.stepExecutions)
	assert.NotNil(t, metrics.stepSuccess)
	assert.NotNil(t, metrics.stepFailures)
	assert.NotNil(t, metrics.stepRetries)
	assert.NotNil(t, metrics.stepDurations)
	assert.NotNil(t, metrics.sagaDurations)
}

func TestSagaMetrics_RecordSagaStart(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()

	// Act
	metrics.RecordSagaStart("TestSaga")

	// Assert
	assert.Equal(t, int64(1), metrics.sagaExecutionsTotal)
	assert.Equal(t, int64(1), metrics.activeSagasCount)
}

func TestSagaMetrics_RecordSagaEnd(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()
	metrics.RecordSagaStart("TestSaga")

	// Act
	metrics.RecordSagaEnd("TestSaga", "success", 100*time.Millisecond)

	// Assert
	assert.Equal(t, int64(0), metrics.activeSagasCount)
	assert.Equal(t, int64(1), metrics.sagaSuccessTotal)
	assert.Equal(t, int64(0), metrics.sagaFailuresTotal)
	assert.Equal(t, int64(0), metrics.sagaCompensationsTotal)
	assert.Len(t, metrics.sagaDurations, 1)
	assert.Equal(t, 100*time.Millisecond, metrics.sagaDurations[0])
}

func TestSagaMetrics_RecordSagaEnd_Failure(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()
	metrics.RecordSagaStart("TestSaga")

	// Act
	metrics.RecordSagaEnd("TestSaga", "failure", 200*time.Millisecond)

	// Assert
	assert.Equal(t, int64(0), metrics.activeSagasCount)
	assert.Equal(t, int64(0), metrics.sagaSuccessTotal)
	assert.Equal(t, int64(1), metrics.sagaFailuresTotal)
	assert.Equal(t, int64(0), metrics.sagaCompensationsTotal)
	assert.Len(t, metrics.sagaDurations, 1)
	assert.Equal(t, 200*time.Millisecond, metrics.sagaDurations[0])
}

func TestSagaMetrics_RecordSagaEnd_Compensated(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()
	metrics.RecordSagaStart("TestSaga")

	// Act
	metrics.RecordSagaEnd("TestSaga", "compensated", 300*time.Millisecond)

	// Assert
	assert.Equal(t, int64(0), metrics.activeSagasCount)
	assert.Equal(t, int64(0), metrics.sagaSuccessTotal)
	assert.Equal(t, int64(0), metrics.sagaFailuresTotal)
	assert.Equal(t, int64(1), metrics.sagaCompensationsTotal)
	assert.Len(t, metrics.sagaDurations, 1)
	assert.Equal(t, 300*time.Millisecond, metrics.sagaDurations[0])
}

func TestSagaMetrics_RecordStepStart(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()

	// Act
	metrics.RecordStepStart("TestSaga", "Step1")

	// Assert
	key := "TestSaga.Step1"
	assert.Equal(t, int64(1), metrics.stepExecutions[key])
}

func TestSagaMetrics_RecordStepEnd_Success(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()
	metrics.RecordStepStart("TestSaga", "Step1")

	// Act
	metrics.RecordStepEnd("TestSaga", "Step1", "success", 50*time.Millisecond)

	// Assert
	key := "TestSaga.Step1"
	assert.Equal(t, int64(1), metrics.stepSuccess[key])
	assert.Equal(t, int64(0), metrics.stepFailures[key])
	assert.Len(t, metrics.stepDurations[key], 1)
	assert.Equal(t, 50*time.Millisecond, metrics.stepDurations[key][0])
}

func TestSagaMetrics_RecordStepEnd_Failure(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()
	metrics.RecordStepStart("TestSaga", "Step1")

	// Act
	metrics.RecordStepEnd("TestSaga", "Step1", "failure", 75*time.Millisecond)

	// Assert
	key := "TestSaga.Step1"
	assert.Equal(t, int64(0), metrics.stepSuccess[key])
	assert.Equal(t, int64(1), metrics.stepFailures[key])
	assert.Len(t, metrics.stepDurations[key], 1)
	assert.Equal(t, 75*time.Millisecond, metrics.stepDurations[key][0])
}

func TestSagaMetrics_RecordStepRetry(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()

	// Act
	metrics.RecordStepRetry("TestSaga", "Step1")

	// Assert
	key := "TestSaga.Step1"
	assert.Equal(t, int64(1), metrics.stepRetries[key])
}

func TestSagaMetrics_GetStats(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()

	// Simular algumas execuções
	metrics.RecordSagaStart("TestSaga")
	metrics.RecordSagaEnd("TestSaga", "success", 100*time.Millisecond)

	metrics.RecordSagaStart("TestSaga")
	metrics.RecordSagaEnd("TestSaga", "failure", 200*time.Millisecond)

	metrics.RecordStepStart("TestSaga", "Step1")
	metrics.RecordStepEnd("TestSaga", "Step1", "success", 50*time.Millisecond)
	metrics.RecordStepRetry("TestSaga", "Step1")

	// Act
	stats := metrics.GetStats()

	// Assert
	assert.NotNil(t, stats)
	assert.Equal(t, int64(2), stats["total_executions"])
	assert.Equal(t, int64(1), stats["success_total"])
	assert.Equal(t, int64(1), stats["failures_total"])
	assert.Equal(t, int64(0), stats["compensations_total"])
	assert.Equal(t, int64(0), stats["active_sagas"])
	assert.Equal(t, float64(50), stats["success_rate"]) // 50% success rate

	stepExecutions := stats["step_executions"].(map[string]int64)
	assert.Equal(t, int64(1), stepExecutions["TestSaga.Step1"])

	stepSuccess := stats["step_success"].(map[string]int64)
	assert.Equal(t, int64(1), stepSuccess["TestSaga.Step1"])

	stepRetries := stats["step_retries"].(map[string]int64)
	assert.Equal(t, int64(1), stepRetries["TestSaga.Step1"])
}

func TestSagaMetrics_Reset(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()

	// Adicionar alguns dados
	metrics.RecordSagaStart("TestSaga")
	metrics.RecordSagaEnd("TestSaga", "success", 100*time.Millisecond)
	metrics.RecordStepStart("TestSaga", "Step1")
	metrics.RecordStepEnd("TestSaga", "Step1", "success", 50*time.Millisecond)

	// Act
	metrics.Reset()

	// Assert
	assert.Equal(t, int64(0), metrics.sagaExecutionsTotal)
	assert.Equal(t, int64(0), metrics.sagaSuccessTotal)
	assert.Equal(t, int64(0), metrics.sagaFailuresTotal)
	assert.Equal(t, int64(0), metrics.sagaCompensationsTotal)
	assert.Equal(t, int64(0), metrics.activeSagasCount)
	assert.Len(t, metrics.stepExecutions, 0)
	assert.Len(t, metrics.stepSuccess, 0)
	assert.Len(t, metrics.stepFailures, 0)
	assert.Len(t, metrics.stepRetries, 0)
	assert.Len(t, metrics.stepDurations, 0)
	assert.Len(t, metrics.sagaDurations, 0)
}

func TestNewSagaMonitor(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()

	// Act
	monitor := NewSagaMonitor(logger)

	// Assert
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.metrics)
	assert.NotNil(t, monitor.logger)
}

func TestSagaMonitor_MonitorSagaExecution_Success(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	monitor := NewSagaMonitor(logger)

	// Act
	err := monitor.MonitorSagaExecution("TestSaga", func() error {
		return nil
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), monitor.metrics.sagaExecutionsTotal)
	assert.Equal(t, int64(1), monitor.metrics.sagaSuccessTotal)
	assert.Equal(t, int64(0), monitor.metrics.sagaFailuresTotal)
}

func TestSagaMonitor_MonitorSagaExecution_Failure(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	monitor := NewSagaMonitor(logger)

	// Act
	err := monitor.MonitorSagaExecution("TestSaga", func() error {
		return assert.AnError
	})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, int64(1), monitor.metrics.sagaExecutionsTotal)
	assert.Equal(t, int64(0), monitor.metrics.sagaSuccessTotal)
	assert.Equal(t, int64(1), monitor.metrics.sagaFailuresTotal)
}

func TestSagaMonitor_MonitorStepExecution_Success(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	monitor := NewSagaMonitor(logger)

	// Act
	err := monitor.MonitorStepExecution("TestSaga", "Step1", func() error {
		return nil
	})

	// Assert
	assert.NoError(t, err)
	key := "TestSaga.Step1"
	assert.Equal(t, int64(1), monitor.metrics.stepExecutions[key])
	assert.Equal(t, int64(1), monitor.metrics.stepSuccess[key])
	assert.Equal(t, int64(0), monitor.metrics.stepFailures[key])
}

func TestSagaMonitor_MonitorStepExecution_Failure(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	monitor := NewSagaMonitor(logger)

	// Act
	err := monitor.MonitorStepExecution("TestSaga", "Step1", func() error {
		return assert.AnError
	})

	// Assert
	assert.Error(t, err)
	key := "TestSaga.Step1"
	assert.Equal(t, int64(1), monitor.metrics.stepExecutions[key])
	assert.Equal(t, int64(0), monitor.metrics.stepSuccess[key])
	assert.Equal(t, int64(1), monitor.metrics.stepFailures[key])
}

func TestSagaMonitor_RecordCompensation(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	monitor := NewSagaMonitor(logger)

	// Act
	monitor.RecordCompensation("TestSaga", "Step1")

	// Assert
	assert.Equal(t, int64(1), monitor.metrics.sagaCompensationsTotal)
}

func TestSagaMonitor_RecordRetry(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	monitor := NewSagaMonitor(logger)

	// Act
	monitor.RecordRetry("TestSaga", "Step1", 2)

	// Assert
	key := "TestSaga.Step1"
	assert.Equal(t, int64(1), monitor.metrics.stepRetries[key])
}

func TestSagaMetrics_ConcurrentAccess(t *testing.T) {
	// Arrange
	metrics := NewSagaMetrics()

	// Simular acesso concorrente
	done := make(chan bool, 100)

	for i := 0; i < 50; i++ {
		go func() {
			metrics.RecordSagaStart("TestSaga")
			metrics.RecordSagaEnd("TestSaga", "success", 100*time.Millisecond)
			done <- true
		}()
	}

	for i := 0; i < 50; i++ {
		go func() {
			metrics.RecordStepStart("TestSaga", "Step1")
			metrics.RecordStepEnd("TestSaga", "Step1", "success", 50*time.Millisecond)
			done <- true
		}()
	}

	// Aguardar todas as goroutines terminarem
	for i := 0; i < 100; i++ {
		<-done
	}

	// Assert
	assert.Equal(t, int64(50), metrics.sagaExecutionsTotal)
	assert.Equal(t, int64(50), metrics.sagaSuccessTotal)
	assert.Equal(t, int64(0), metrics.activeSagasCount)

	key := "TestSaga.Step1"
	assert.Equal(t, int64(50), metrics.stepExecutions[key])
	assert.Equal(t, int64(50), metrics.stepSuccess[key])
}
