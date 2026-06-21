package handler

import (
	"encoding/json"
	"net/http"

	"github.com/keepguard/bff-core/internal/domain/saga"
	"go.uber.org/zap"
)

// DashboardHandler manipula requisições do dashboard de monitoramento
type DashboardHandler struct {
	logger      *zap.Logger
	sagaMetrics *saga.SagaMetrics
}

// NewDashboardHandler cria um novo handler de dashboard
func NewDashboardHandler(
	logger *zap.Logger,
	sagaMetrics *saga.SagaMetrics,
) *DashboardHandler {
	return &DashboardHandler{
		logger:      logger,
		sagaMetrics: sagaMetrics,
	}
}

// Dashboard serve a página principal do dashboard
func (dh *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SAGA Monitor Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f7fa;
            color: #333;
            line-height: 1.6;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 2rem;
            text-align: center;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        
        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 0.5rem;
        }
        
        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }
        
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 2rem;
            margin-bottom: 2rem;
        }
        
        .card {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            border-left: 4px solid #667eea;
            transition: transform 0.2s ease;
        }
        
        .card:hover {
            transform: translateY(-2px);
        }
        
        .card h3 {
            color: #667eea;
            margin-bottom: 1rem;
            font-size: 1.3rem;
        }
        
        .status {
            display: inline-block;
            padding: 0.3rem 0.8rem;
            border-radius: 20px;
            font-size: 0.9rem;
            font-weight: 600;
            text-transform: uppercase;
        }
        
        .status.healthy {
            background: #d4edda;
            color: #155724;
        }
        
        .status.unhealthy {
            background: #f8d7da;
            color: #721c24;
        }
        
        .status.warning {
            background: #fff3cd;
            color: #856404;
        }
        
        .metric {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.8rem 0;
            border-bottom: 1px solid #eee;
        }
        
        .metric:last-child {
            border-bottom: none;
        }
        
        .metric-value {
            font-weight: 600;
            font-size: 1.1rem;
        }
        
        .metric-label {
            color: #666;
        }
        
        .refresh-btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 0.8rem 1.5rem;
            border-radius: 8px;
            font-size: 1rem;
            cursor: pointer;
            transition: background 0.2s ease;
            margin-bottom: 2rem;
        }
        
        .refresh-btn:hover {
            background: #5a6fd8;
        }
        
        .alert-item {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 8px;
            padding: 1rem;
            margin-bottom: 1rem;
        }
        
        .alert-item.critical {
            background: #f8d7da;
            border-color: #f5c6cb;
        }
        
        .alert-item.warning {
            background: #fff3cd;
            border-color: #ffeaa7;
        }
        
        .alert-title {
            font-weight: 600;
            margin-bottom: 0.5rem;
        }
        
        .alert-time {
            font-size: 0.9rem;
            color: #666;
        }
        
        .footer {
            text-align: center;
            padding: 2rem;
            color: #666;
            border-top: 1px solid #eee;
            margin-top: 2rem;
        }
        
        .loading {
            text-align: center;
            padding: 2rem;
            color: #666;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #667eea;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 2s linear infinite;
            margin: 0 auto;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 SAGA Monitor Dashboard</h1>
        <p>Monitoramento em tempo real do sistema de transações distribuídas</p>
    </div>
    
    <div class="container">
        <button class="refresh-btn" onclick="refreshDashboard()">🔄 Atualizar Dashboard</button>
        
        <div class="grid">
            <div class="card">
                <h3>📊 Métricas SAGA</h3>
                <div class="metric">
                    <span class="metric-label">Execuções Totais:</span>
                    <span class="metric-value" id="total-executions">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Taxa de Sucesso:</span>
                    <span class="metric-value" id="success-rate">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Compensações:</span>
                    <span class="metric-value" id="compensations">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">SAGAs Ativas:</span>
                    <span class="metric-value" id="active-sagas">-</span>
                </div>
            </div>
            
            <div class="card">
                <h3>🏥 Status dos Serviços</h3>
                <div id="services-status">
                    <div class="loading">
                        <div class="spinner"></div>
                        <p>Carregando status dos serviços...</p>
                    </div>
                </div>
            </div>
            
            <div class="card">
                <h3>⚠️ Alertas Recentes</h3>
                <div id="recent-alerts">
                    <div class="loading">
                        <div class="spinner"></div>
                        <p>Carregando alertas...</p>
                    </div>
                </div>
            </div>
            
            <div class="card">
                <h3>📈 Performance</h3>
                <div class="metric">
                    <span class="metric-label">Tempo Médio SAGA:</span>
                    <span class="metric-value" id="avg-saga-time">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Tempo Médio Step:</span>
                    <span class="metric-value" id="avg-step-time">-</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Retries por Step:</span>
                    <span class="metric-value" id="step-retries">-</span>
                </div>
            </div>
        </div>
    </div>
    
    <div class="footer">
        <p>SAGA Monitor Dashboard - KeepGuard Template</p>
        <p>Última atualização: <span id="last-update">-</span></p>
    </div>
    
    <script>
        async function refreshDashboard() {
            try {
                // Atualizar métricas SAGA
                const metricsResponse = await fetch('/api/metrics');
                const metricsData = await metricsResponse.json();
                updateMetrics(metricsData);
                
                // Atualizar outros componentes
                updateServicesStatus();
                updateAlerts();
                
                // Atualizar timestamp
                document.getElementById('last-update').textContent = new Date().toLocaleString('pt-BR');
                
            } catch (error) {
                console.error('Erro ao atualizar dashboard:', error);
            }
        }
        
        function updateServicesStatus() {
            const container = document.getElementById('services-status');
            container.innerHTML = '<p style="color: #28a745;">✅ Status dos serviços será implementado futuramente</p>';
        }
        
        function updateMetrics(metrics) {
            document.getElementById('total-executions').textContent = metrics.total_executions || '-';
            document.getElementById('success-rate').textContent = metrics.success_rate || '-';
            document.getElementById('compensations').textContent = metrics.compensations || '-';
            document.getElementById('active-sagas').textContent = metrics.active_sagas || '-';
            document.getElementById('avg-saga-time').textContent = metrics.avg_saga_time || '-';
            document.getElementById('avg-step-time').textContent = metrics.avg_step_time || '-';
            document.getElementById('step-retries').textContent = metrics.step_retries || '-';
        }
        
        function updateAlerts() {
            const container = document.getElementById('recent-alerts');
            container.innerHTML = '<p style="color: #28a745;">✅ Sistema de alertas será implementado futuramente</p>';
        }
        
        // Atualizar dashboard a cada 30 segundos
        setInterval(refreshDashboard, 30000);
        
        // Carregar dashboard inicial
        refreshDashboard();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// MetricsAPI retorna métricas em formato JSON
func (dh *DashboardHandler) MetricsAPI(w http.ResponseWriter, r *http.Request) {
	stats := dh.sagaMetrics.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		dh.logger.Error("Failed to encode metrics response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
