package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator interface para validação
type Validator interface {
	ValidateStruct(s interface{}) error
	ValidateVar(field interface{}, tag string) error
}

// customValidator implementa Validator usando go-playground/validator
type customValidator struct {
	validator *validator.Validate
}

// NewValidator cria um novo validador
func NewValidator() Validator {
	v := validator.New()

	// Registra validações customizadas
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("email", validateEmail)
	v.RegisterValidation("password", validatePassword)

	return &customValidator{validator: v}
}

// ValidateStruct valida uma struct
func (cv *customValidator) ValidateStruct(s interface{}) error {
	return cv.validator.Struct(s)
}

// ValidateVar valida uma variável individual
func (cv *customValidator) ValidateVar(field interface{}, tag string) error {
	return cv.validator.Var(field, tag)
}

// validateUsername valida formato de username
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Não pode estar vazio
	if strings.TrimSpace(username) == "" {
		return false
	}

	// Deve ter entre 3 e 50 caracteres
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	// Apenas letras, números, underscore e hífen
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// validateEmail valida formato de email
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// Não pode estar vazio
	if strings.TrimSpace(email) == "" {
		return false
	}

	// Deve conter @
	if !strings.Contains(email, "@") {
		return false
	}

	// Deve ter pelo menos um ponto após o @
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if !strings.Contains(parts[1], ".") {
		return false
	}

	// Validação básica de formato
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}

	return true
}

// validatePassword valida formato de password
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Não pode estar vazio
	if strings.TrimSpace(password) == "" {
		return false
	}

	// Deve ter pelo menos 8 caracteres
	if len(password) < 8 {
		return false
	}

	// Deve ter no máximo 128 caracteres
	if len(password) > 128 {
		return false
	}

	return true
}

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implementa a interface error
func (ve ValidationError) Error() string {
	return ve.Message
}

// ValidationErrors representa múltiplos erros de validação
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implementa a interface error
func (ve ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, err.Message)
	}

	return strings.Join(messages, "; ")
}

// FormatValidationErrors formata erros de validação do validator
func FormatValidationErrors(err error) ValidationErrors {
	var validationErrors ValidationErrors

	if err == nil {
		return validationErrors
	}

	// Converte erros do validator para ValidationError
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationError := ValidationError{
				Field: fieldError.Field(),
				Tag:   fieldError.Tag(),
				Value: fmt.Sprintf("%v", fieldError.Value()),
			}

			// Gera mensagem customizada baseada na tag
			validationError.Message = getValidationMessage(fieldError)
			validationErrors.Errors = append(validationErrors.Errors, validationError)
		}
	} else {
		// Se não for um ValidationErrors, cria um erro genérico
		validationErrors.Errors = append(validationErrors.Errors, ValidationError{
			Field:   "unknown",
			Tag:     "unknown",
			Value:   "",
			Message: err.Error(),
		})
	}

	return validationErrors
}

// getValidationMessage retorna mensagem customizada para cada tag de validação
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s é obrigatório", fe.Field())
	case "min":
		return fmt.Sprintf("%s deve ter pelo menos %s caracteres", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s deve ter no máximo %s caracteres", fe.Field(), fe.Param())
	case "email":
		return fmt.Sprintf("%s deve ser um email válido", fe.Field())
	case "username":
		return fmt.Sprintf("%s deve conter apenas letras, números, underscore e hífen (3-50 caracteres)", fe.Field())
	case "password":
		return fmt.Sprintf("%s deve ter entre 8 e 128 caracteres", fe.Field())
	case "len":
		return fmt.Sprintf("%s deve ter exatamente %s caracteres", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s é inválido", fe.Field())
	}
}

