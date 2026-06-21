package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger interface para logging
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}

// zapLogger implementa Logger usando zap
type zapLogger struct {
	logger *zap.Logger
}

// New cria um novo logger
func New(level, format string) (Logger, error) {
	var config zap.Config

	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Configura o nível de log
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Configura o encoder para logs estruturados
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// Configura o output para stdout (para Logstash)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{logger: logger}, nil
}

// NewNop cria um logger que não faz nada (para testes)
func NewNop() Logger {
	return &zapLogger{logger: zap.NewNop()}
}

// Debug logs a message at DebugLevel
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a message at FatalLevel
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// With creates a child logger and adds structured context to it
func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// Sync flushes any buffered log entries
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// GetZapLogger retorna o *zap.Logger subjacente (para casos onde é necessário passar diretamente)
func (l *zapLogger) GetZapLogger() *zap.Logger {
	return l.logger
}

