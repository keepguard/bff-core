package saga

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewInMemorySagaExecutor(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()

	// Act
	executor := NewInMemorySagaExecutor(logger)

	// Assert
	assert.NotNil(t, executor)
	assert.NotNil(t, executor.logger)
	assert.NotNil(t, executor.monitor)
}

func TestInMemorySagaExecutor_Execute_Success(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	// Criar SAGA simples com um step
	step1Executed := false
	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					step1Executed = true
					return nil
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.NoError(t, err)
	assert.True(t, step1Executed)
}

func TestInMemorySagaExecutor_Execute_StepFailure(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	// Criar SAGA com step que falha
	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					return errors.New("step failed")
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step failed")
}

func TestInMemorySagaExecutor_Execute_WithCompensation(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	step1Executed := false
	step1Compensated := false
	step2Executed := false

	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					step1Executed = true
					return nil
				},
				Compensate: func(ctx context.Context, data map[string]interface{}) error {
					step1Compensated = true
					return nil
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
			{
				Name: "Step2",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					step2Executed = true
					return errors.New("step2 failed")
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step failed")
	assert.True(t, step1Executed)
	assert.True(t, step2Executed)
	assert.True(t, step1Compensated)
}

func TestInMemorySagaExecutor_Execute_CompensationFailure(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	step1Executed := false
	step2Executed := false

	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					step1Executed = true
					return nil
				},
				Compensate: func(ctx context.Context, data map[string]interface{}) error {
					return errors.New("compensation failed")
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
			{
				Name: "Step2",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					step2Executed = true
					return errors.New("step2 failed")
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "saga falhou e compensação falhou")
	assert.True(t, step1Executed)
	assert.True(t, step2Executed)
}

func TestInMemorySagaExecutor_Execute_WithRetry(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	attemptCount := 0
	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					attemptCount++
					if attemptCount < 3 {
						return errors.New("temporary failure")
					}
					return nil
				},
				MaxRetries: 3,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 3, attemptCount)
}

func TestInMemorySagaExecutor_Execute_RetryExhausted(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	attemptCount := 0
	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					attemptCount++
					return errors.New("persistent failure")
				},
				MaxRetries: 2,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step failed")
	assert.Equal(t, 2, attemptCount)
}

func TestInMemorySagaExecutor_Execute_ContextCancellation(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx, cancel := context.WithCancel(context.Background())

	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					// Cancelar contexto durante a execução
					cancel()
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
						return nil
					}
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step failed")
}

func TestInMemorySagaExecutor_Execute_Timeout(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					// Simular operação que demora mais que o timeout
					select {
					case <-time.After(2 * time.Second):
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				},
				MaxRetries: 1,
				Timeout:    100 * time.Millisecond,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step failed")
}

func TestInMemorySagaExecutor_Execute_NoSteps(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	saga := InMemorySaga{
		Name:  "TestSaga",
		Steps: []InMemoryStep{},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.NoError(t, err)
}

func TestInMemorySagaExecutor_Execute_MultipleSteps(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	executionOrder := make([]string, 0)
	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					executionOrder = append(executionOrder, "Step1")
					return nil
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
			{
				Name: "Step2",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					executionOrder = append(executionOrder, "Step2")
					return nil
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
			{
				Name: "Step3",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					executionOrder = append(executionOrder, "Step3")
					return nil
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, []string{"Step1", "Step2", "Step3"}, executionOrder)
}

func TestInMemorySagaExecutor_Execute_CompensationOrder(t *testing.T) {
	// Arrange
	logger, _ := zap.NewDevelopment()
	executor := NewInMemorySagaExecutor(logger)
	ctx := context.Background()

	compensationOrder := make([]string, 0)
	saga := InMemorySaga{
		Name: "TestSaga",
		Steps: []InMemoryStep{
			{
				Name: "Step1",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					return nil
				},
				Compensate: func(ctx context.Context, data map[string]interface{}) error {
					compensationOrder = append(compensationOrder, "Step1")
					return nil
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
			{
				Name: "Step2",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					return nil
				},
				Compensate: func(ctx context.Context, data map[string]interface{}) error {
					compensationOrder = append(compensationOrder, "Step2")
					return nil
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
			{
				Name: "Step3",
				Execute: func(ctx context.Context, data map[string]interface{}) error {
					return errors.New("step3 failed")
				},
				MaxRetries: 1,
				Timeout:    time.Second,
			},
		},
	}

	data := map[string]interface{}{
		"test": "data",
	}

	// Act
	err := executor.Execute(ctx, saga, data)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step failed")
	// Compensação deve ser em ordem reversa
	assert.Equal(t, []string{"Step2", "Step1"}, compensationOrder)
}
