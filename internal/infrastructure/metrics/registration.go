package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Contadores únicos do pacote (promauto registra uma vez por processo).
var (
	registrationStepTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "registration_step_total",
			Help: "Passos da saga de cadastro por resultado",
		},
		[]string{"step", "result"},
	)
	registrationCompensationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "registration_compensation_total",
			Help: "Compensações da saga de cadastro por resultado",
		},
		[]string{"step", "result"},
	)
)

// RegistrationMetrics implementa port.RegistrationMetrics com os contadores acima.
type RegistrationMetrics struct{}

// NewRegistrationMetrics devolve o registrador das métricas de cadastro.
func NewRegistrationMetrics() *RegistrationMetrics {
	return &RegistrationMetrics{}
}

// RecordStep conta um passo da saga de cadastro.
func (*RegistrationMetrics) RecordStep(step, result string) {
	registrationStepTotal.WithLabelValues(step, result).Inc()
}

// RecordCompensation conta uma compensação da saga de cadastro.
func (*RegistrationMetrics) RecordCompensation(step, result string) {
	registrationCompensationTotal.WithLabelValues(step, result).Inc()
}
