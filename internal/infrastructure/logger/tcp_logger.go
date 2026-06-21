package logger

import (
	"encoding/json"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// tcpLogger implementa Logger enviando logs via TCP
type tcpLogger struct {
	conn    net.Conn
	logger  *zap.Logger
	service string
}

// NewTCPLogger cria um novo logger TCP
func NewTCPLogger(service, tcpAddr string) (Logger, error) {
	// Conecta ao Filebeat via TCP
	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		// Se não conseguir conectar, usa logger padrão e loga o warning
		baseLogger, logErr := New("info", "json")
		if logErr != nil {
			return nil, logErr
		}

		// Loga o warning sobre a falha na conexão TCP
		baseLogger.Warn("Não foi possível conectar ao Filebeat, usando logger padrão",
			zap.String("address", tcpAddr),
			zap.String("service", service),
			zap.Error(err),
		)

		return baseLogger, nil
	}

	// Configura logger básico
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &tcpLogger{
		conn:    conn,
		logger:  logger,
		service: service,
	}, nil
}

// sendToTCP envia log via TCP para o Filebeat
func (l *tcpLogger) sendToTCP(level, msg string, fields map[string]interface{}) {
	// Estrutura de log otimizada para Loki/Grafana
	logEntry := map[string]interface{}{
		"@timestamp":  time.Now().Format(time.RFC3339),
		"timestamp":   time.Now().Format(time.RFC3339),
		"level":       level,
		"message":     msg,
		"service":     l.service,
		"component":   "bff-core",
		"environment": getEnvOrDefault("ENV", "local"),
		"version":     "1.0.0",
	}

	// Adiciona campos adicionais
	for k, v := range fields {
		logEntry[k] = v
	}

	// Converte para JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		l.logger.Error("Erro ao serializar log", zap.Error(err))
		return
	}

	// Envia via TCP
	_, err = l.conn.Write(append(jsonData, '\n'))
	if err != nil {
		l.logger.Error("Erro ao enviar log via TCP", zap.Error(err))
		// Tenta reconectar
		l.reconnect()
	}
}

// getEnvOrDefault retorna valor da variável de ambiente ou valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// reconnect tenta reconectar ao Filebeat
func (l *tcpLogger) reconnect() {
	if l.conn != nil {
		l.conn.Close()
	}

	conn, err := net.Dial("tcp", "localhost:5000")
	if err != nil {
		l.logger.Error("Erro ao reconectar ao Filebeat", zap.Error(err))
		return
	}
	l.conn = conn
}

// Debug logs a message at DebugLevel
func (l *tcpLogger) Debug(msg string, fields ...zap.Field) {
	fieldMap := l.fieldsToMap(fields)
	l.sendToTCP("debug", msg, fieldMap)
	l.logger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel
func (l *tcpLogger) Info(msg string, fields ...zap.Field) {
	fieldMap := l.fieldsToMap(fields)
	l.sendToTCP("info", msg, fieldMap)
	l.logger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel
func (l *tcpLogger) Warn(msg string, fields ...zap.Field) {
	fieldMap := l.fieldsToMap(fields)
	l.sendToTCP("warn", msg, fieldMap)
	l.logger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel
func (l *tcpLogger) Error(msg string, fields ...zap.Field) {
	fieldMap := l.fieldsToMap(fields)
	l.sendToTCP("error", msg, fieldMap)
	l.logger.Error(msg, fields...)
}

// Fatal logs a message at FatalLevel
func (l *tcpLogger) Fatal(msg string, fields ...zap.Field) {
	fieldMap := l.fieldsToMap(fields)
	l.sendToTCP("fatal", msg, fieldMap)
	l.logger.Fatal(msg, fields...)
}

// With creates a child logger and adds structured context to it
func (l *tcpLogger) With(fields ...zap.Field) Logger {
	return &tcpLogger{
		conn:    l.conn,
		logger:  l.logger.With(fields...),
		service: l.service,
	}
}

// Sync flushes any buffered log entries
func (l *tcpLogger) Sync() error {
	if l.conn != nil {
		l.conn.Close()
	}
	return l.logger.Sync()
}

// fieldsToMap converte zap.Field para map[string]interface{}
func (l *tcpLogger) fieldsToMap(fields []zap.Field) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		result[field.Key] = field.Interface
	}
	return result
}
