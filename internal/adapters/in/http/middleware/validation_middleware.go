package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/keepguard/bff-core/internal/infrastructure/validation"
	"github.com/labstack/echo/v4"
)

// ValidationMiddleware middleware de validação
func ValidationMiddleware(v validation.Validator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("validator", v)
			return next(c)
		}
	}
}

// CustomValidator implementa echo.Validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator cria um novo validador
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// Validate valida uma struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
