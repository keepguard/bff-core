package connections

import (
	"strings"

	"github.com/keepguard/bff-core/internal/infrastructure/config"
)

// Target é um serviço a ser verificado na tela Conexões.
type Target struct {
	ID            string
	Name          string
	Description   string
	Group         string
	Endpoint      string
	URL           string
	TreatAuthAsUp bool
}

type catalogItem struct {
	ID            string
	Name          string
	Description   string
	Group         string
	Endpoint      string
	DefaultURL    string
	TreatAuthAsUp bool
}

var catalog = []catalogItem{
	{ID: "front-keepguard-core", Name: "Frontend", Description: "Painel KeepGuard servido pelo nginx.", Group: "gateway", Endpoint: "GET /healthz", DefaultURL: "http://localhost:5173/healthz"},
	{ID: "bff-auth", Name: "BFF Auth", Description: "Autenticação, sessões e tokens de acesso.", Group: "gateway", Endpoint: "GET /health", DefaultURL: "http://localhost:8381/health"},
	{ID: "bff-core", Name: "BFF Core", Description: "Cadastro, consentimentos e serviços de núcleo.", Group: "gateway", Endpoint: "GET /health", DefaultURL: "http://127.0.0.1:8382/health"},
	{ID: "ms-auth", Name: "MS Auth", Description: "Identidade, roles e ciclo de vida da conta.", Group: "microservice", Endpoint: "GET /actuator/health/liveness", DefaultURL: "http://localhost:8081/actuator/health/liveness"},
	{ID: "ms-communication", Name: "MS Communication", Description: "Notificações e canais de comunicação.", Group: "microservice", Endpoint: "GET /actuator/health/liveness", DefaultURL: "http://localhost:8082/actuator/health/liveness"},
	{ID: "ms-company", Name: "MS Company", Description: "Tenants, empresas e provisionamento.", Group: "microservice", Endpoint: "GET /actuator/health/liveness", DefaultURL: "http://localhost:8083/actuator/health/liveness"},
	{ID: "ms-user", Name: "MS User", Description: "Perfil e cadastro de usuários.", Group: "microservice", Endpoint: "GET /actuator/health/liveness", DefaultURL: "http://localhost:8085/actuator/health/liveness"},
	{ID: "ms-user-consents", Name: "MS User Consents", Description: "Consentimentos e documentos LGPD.", Group: "microservice", Endpoint: "GET /actuator/health/liveness", DefaultURL: "http://localhost:8086/actuator/health/liveness"},
	{ID: "srv-email-sender", Name: "SRV Email Sender", Description: "Worker de envio de e-mail.", Group: "worker", Endpoint: "GET /health", DefaultURL: "http://localhost:8601/health"},
	{ID: "srv-token-manager", Name: "SRV Token Manager", Description: "Gestão de tokens OAuth de provedores.", Group: "worker", Endpoint: "GET /health", DefaultURL: "http://localhost:8700/health"},
	{ID: "srv-sms-sender", Name: "SRV SMS Sender", Description: "Worker de envio assíncrono de SMS.", Group: "worker", Endpoint: "GET /health", DefaultURL: "http://localhost:8610/health"},
	{ID: "mock-sms-gateway", Name: "Mock SMS Gateway", Description: "Gateway simulado de SMS (ambiente local).", Group: "worker", Endpoint: "GET /health", DefaultURL: "http://localhost:8089/health"},
	{ID: "minio", Name: "MinIO", Description: "Object storage de avatares, documentos e consents.", Group: "infra", Endpoint: "GET /minio/health/live", DefaultURL: "http://localhost:9000/minio/health/live"},
	{ID: "rabbitmq", Name: "RabbitMQ", Description: "Filas de notificação e workers.", Group: "infra", Endpoint: "GET :15672/api/health/checks/alarms", DefaultURL: "http://localhost:15672/api/health/checks/alarms", TreatAuthAsUp: true},
	{ID: "prometheus", Name: "Prometheus", Description: "Coleta de métricas da stack.", Group: "infra", Endpoint: "GET /-/healthy", DefaultURL: "http://localhost:9095/-/healthy"},
	{ID: "grafana", Name: "Grafana", Description: "Dashboards de observabilidade.", Group: "infra", Endpoint: "GET /api/health", DefaultURL: "http://localhost:3001/api/health"},
}

// Targets monta a lista efetiva com URLs do config (ou default local).
func Targets(cfg config.ConnectionsHealthConfig) []Target {
	out := make([]Target, 0, len(catalog))
	for _, item := range catalog {
		url := item.DefaultURL
		if override := urlOverride(cfg.URLs, item.ID); override != "" {
			url = override
		}
		out = append(out, Target{
			ID:            item.ID,
			Name:          item.Name,
			Description:   item.Description,
			Group:         item.Group,
			Endpoint:      item.Endpoint,
			URL:           url,
			TreatAuthAsUp: item.TreatAuthAsUp,
		})
	}
	return out
}

func urlOverride(urls map[string]string, id string) string {
	if urls == nil {
		return ""
	}
	if value := strings.TrimSpace(urls[id]); value != "" {
		return value
	}
	return strings.TrimSpace(urls[strings.ReplaceAll(id, "-", "_")])
}
