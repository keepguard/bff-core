package http

import (
	"time"

	"github.com/labstack/echo/v4"
)

// Middleware define a interface para middlewares HTTP
type Middleware interface {
	RequestIDMiddleware() echo.MiddlewareFunc
	CorrelationIDMiddleware() echo.MiddlewareFunc
	LoggingMiddleware() echo.MiddlewareFunc
	RecoveryMiddleware() echo.MiddlewareFunc
	CORSMiddleware() echo.MiddlewareFunc
	SecurityMiddleware() echo.MiddlewareFunc
	MetricsMiddleware() echo.MiddlewareFunc
	TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc
}
