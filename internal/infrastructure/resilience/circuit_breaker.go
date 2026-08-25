package resilience

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"

	dto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/sony/gobreaker"
)

// CircuitBreakerManager gerencia circuit breakers para diferentes serviços
type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	metrics  MetricsRecorder
}

// MetricsRecorder interface para registrar métricas do circuit breaker
type MetricsRecorder interface {
	SetCircuitBreakerState(service string, state int)
}

// CircuitBreakerConfig configuração do circuit breaker
type CircuitBreakerConfig struct {
	Name         string
	MaxRequests  uint32
	Interval     time.Duration
	Timeout      time.Duration
	ReadyToTrip  func(counts gobreaker.Counts) bool
	IsSuccessful func(err error) bool
}

// NewCircuitBreakerManager cria um novo gerenciador de circuit breakers
func NewCircuitBreakerManager(metrics MetricsRecorder) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		metrics:  metrics,
	}
}

// DefaultIsSuccessful considera sucesso respostas normais e todos os erros 4xx (negócio/validação/credenciais)
// Apenas falhas reais de infraestrutura (5xx, connection refused, timeouts) contam como falha para o Circuit Breaker.
func DefaultIsSuccessful(err error) bool {
	if err == nil {
		return true
	}

	// 1. Verifica se é um HTTPError com status code 4xx
	var httpErr *dto.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Code >= 400 && httpErr.Code < 500
	}

	// 2. Verifica se é um AppError com status code 4xx
	var appErr *pkg.AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode >= 400 && appErr.StatusCode < 500
	}

	// 3. Verifica erros de conexão de rede ou timeout (estes são falhas de infraestrutura)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return false
	}
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, http.ErrHandlerTimeout) {
		return false
	}

	return false
}

// GetOrCreate obtém ou cria um circuit breaker para um serviço
func (m *CircuitBreakerManager) GetOrCreate(service string, config CircuitBreakerConfig) *gobreaker.CircuitBreaker {
	if breaker, exists := m.breakers[service]; exists {
		return breaker
	}

	isSuccessful := config.IsSuccessful
	if isSuccessful == nil {
		isSuccessful = DefaultIsSuccessful
	}

	settings := gobreaker.Settings{
		Name:         config.Name,
		MaxRequests:  config.MaxRequests,
		Interval:     config.Interval,
		Timeout:      config.Timeout,
		ReadyToTrip:  config.ReadyToTrip,
		IsSuccessful: isSuccessful,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			m.recordStateChange(service, to)
		},
	}

	if settings.ReadyToTrip == nil {
		settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		}
	}

	breaker := gobreaker.NewCircuitBreaker(settings)
	m.breakers[service] = breaker

	return breaker
}

// Execute executa uma função com circuit breaker
func (m *CircuitBreakerManager) Execute(ctx context.Context, service string, fn func() (interface{}, error)) (interface{}, error) {
	breaker := m.breakers[service]
	if breaker == nil {
		return fn()
	}

	return breaker.Execute(func() (interface{}, error) {
		return fn()
	})
}

// recordStateChange registra mudança de estado do circuit breaker
func (m *CircuitBreakerManager) recordStateChange(service string, state gobreaker.State) {
	var cbState int

	switch state {
	case gobreaker.StateClosed:
		cbState = 0 // CircuitBreakerClosed
	case gobreaker.StateOpen:
		cbState = 1 // CircuitBreakerOpen
	case gobreaker.StateHalfOpen:
		cbState = 2 // CircuitBreakerHalfOpen
	}

	if m.metrics != nil {
		m.metrics.SetCircuitBreakerState(service, cbState)
	}
}

// GetState retorna o estado atual do circuit breaker
func (m *CircuitBreakerManager) GetState(service string) gobreaker.State {
	breaker := m.breakers[service]
	if breaker == nil {
		return gobreaker.StateClosed
	}
	return breaker.State()
}

// Reset reseta o circuit breaker de um serviço
func (m *CircuitBreakerManager) Reset(service string) error {
	breaker := m.breakers[service]
	if breaker == nil {
		return fmt.Errorf("circuit breaker for service %s not found", service)
	}

	// O gobreaker não tem método Reset, então não fazemos nada
	// O circuit breaker se resetará automaticamente quando as chamadas começarem a ter sucesso
	return nil
}

// GetCounts retorna as contagens do circuit breaker
func (m *CircuitBreakerManager) GetCounts(service string) gobreaker.Counts {
	breaker := m.breakers[service]
	if breaker == nil {
		return gobreaker.Counts{}
	}
	return breaker.Counts()
}

// DefaultConfig retorna uma configuração padrão para circuit breaker
func DefaultConfig(service string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:        service,
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	}
}

