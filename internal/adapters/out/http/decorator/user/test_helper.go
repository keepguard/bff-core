package user

import (
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
)

var (
	// sharedMetrics é uma instância compartilhada de métricas para os testes
	sharedMetrics *metrics.Metrics
)

// getTestMetrics retorna uma instância de métricas para testes
func getTestMetrics() *metrics.Metrics {
	if sharedMetrics == nil {
		sharedMetrics = metrics.New()
	}
	return sharedMetrics
}
