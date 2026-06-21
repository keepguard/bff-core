package logger

import "go.uber.org/zap"

// Helper functions para campos comuns
func TraceID(traceID string) zap.Field {
	return zap.String("traceId", traceID)
}

func SpanID(spanID string) zap.Field {
	return zap.String("spanId", spanID)
}

func RequestID(requestID string) zap.Field {
	return zap.String("requestId", requestID)
}

func UserID(userID string) zap.Field {
	return zap.String("userId", userID)
}

func Route(route string) zap.Field {
	return zap.String("route", route)
}

func Method(method string) zap.Field {
	return zap.String("method", method)
}

func Status(status int) zap.Field {
	return zap.Int("status", status)
}

func Latency(latency int64) zap.Field {
	return zap.Int64("latencyMs", latency)
}

func Service(service string) zap.Field {
	return zap.String("service", service)
}

func Component(component string) zap.Field {
	return zap.String("component", component)
}

func Environment(env string) zap.Field {
	return zap.String("environment", env)
}

func Version(version string) zap.Field {
	return zap.String("version", version)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func Duration(duration string) zap.Field {
	return zap.String("duration", duration)
}

func IP(ip string) zap.Field {
	return zap.String("ip", ip)
}

func UserAgent(userAgent string) zap.Field {
	return zap.String("userAgent", userAgent)
}

func ResponseSize(size int) zap.Field {
	return zap.Int("responseSize", size)
}

