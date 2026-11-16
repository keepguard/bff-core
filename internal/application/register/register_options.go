package register

import (
	"time"

	"github.com/keepguard/bff-core/internal/application/port"
)

// RegisterOptions reúne o que a saga endurecida do cadastro precisa além dos clientes.
// SagaHardening vem de registration.saga_hardening.enabled: desligado, o comportamento visível é o anterior.
// State (Redis) e Metrics podem ser nil: sem State não há trava nem replay; sem Metrics nada é contado.
type RegisterOptions struct {
	SagaHardening bool
	State         port.RegisterStateStore
	Metrics       port.RegistrationMetrics
}

const (
	// registerLockKeyPrefix e registerResultKeyPrefix seguem o prefixo do produto (kg:<serviço>:<área>:<id>).
	registerLockKeyPrefix   = "kg:bff-core:register:lock:"
	registerResultKeyPrefix = "kg:bff-core:register:result:"

	// registerLockTTL cobre a saga inteira no pior caso (somas dos timeouts dos passos) e expira sozinha se o
	// processo morrer no meio.
	registerLockTTL = 2 * time.Minute
	// registerResultTTL é o prazo em que o mesmo confirm repetido devolve o resultado (sem reenviar e-mail).
	registerResultTTL = 10 * time.Minute
	// registerWaitTimeout é quanto um confirm concorrente espera pelo resultado do que já está rodando.
	registerWaitTimeout  = 30 * time.Second
	registerPollInterval = 250 * time.Millisecond
	// authRecoveryTimeout limita a consulta ao ms-auth quando o passo 5 estoura o prazo (ambíguo).
	authRecoveryTimeout = 5 * time.Second
)

// noopRegistrationMetrics é usado quando nenhuma métrica foi injetada (testes e montagem parcial).
type noopRegistrationMetrics struct{}

func (noopRegistrationMetrics) RecordStep(string, string)         {}
func (noopRegistrationMetrics) RecordCompensation(string, string) {}

func metricsOrNoop(m port.RegistrationMetrics) port.RegistrationMetrics {
	if m == nil {
		return noopRegistrationMetrics{}
	}
	return m
}

// metrics devolve o registrador de métricas do usecase (no-op se não injetado).
func (uc *registerConfirmUseCaseImpl) metrics() port.RegistrationMetrics {
	return metricsOrNoop(uc.opts.Metrics)
}
