package pkg

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse representa a resposta de erro padronizada
type ErrorResponse struct {
	Error   string        `json:"error"`
	Message string        `json:"message"`
	TraceID string        `json:"traceId,omitempty"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail representa detalhes específicos do erro
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// AppError representa um erro da aplicação
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Details    []ErrorDetail
	TraceID    string
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError cria um novo erro da aplicação
func NewAppError(code, message string, statusCode int, details ...ErrorDetail) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}
}

// WithTraceID adiciona um trace ID ao erro
func (e *AppError) WithTraceID(traceID string) *AppError {
	e.TraceID = traceID
	return e
}

// ToResponse converte o erro para ErrorResponse
func (e *AppError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Error:   e.Code,
		Message: e.Message,
		TraceID: e.TraceID,
		Details: e.Details,
	}
}

// WriteError escreve um erro HTTP padronizado
func WriteError(w http.ResponseWriter, err error, traceID string) {
	var appErr *AppError

	switch e := err.(type) {
	case *AppError:
		appErr = e
	default:
		appErr = NewAppError("INTERNAL_ERROR", "Erro interno do servidor", http.StatusInternalServerError)
	}

	if appErr.TraceID == "" {
		appErr.TraceID = traceID
	}

	response := appErr.ToResponse()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	json.NewEncoder(w).Encode(response)
}

// Erros pré-definidos
var (
	ErrInvalidRequest     = NewAppError("INVALID_REQUEST", "Requisição inválida", http.StatusBadRequest)
	ErrUnauthorized       = NewAppError("UNAUTHORIZED", "Não autorizado", http.StatusUnauthorized)
	ErrForbidden          = NewAppError("FORBIDDEN", "Acesso negado", http.StatusForbidden)
	ErrNotFound           = NewAppError("NOT_FOUND", "Recurso não encontrado", http.StatusNotFound)
	ErrConflict           = NewAppError("CONFLICT", "Conflito de recursos", http.StatusConflict)
	ErrInternalServer     = NewAppError("INTERNAL_ERROR", "Erro interno do servidor", http.StatusInternalServerError)
	ErrServiceUnavailable = NewAppError("SERVICE_UNAVAILABLE", "Serviço indisponível", http.StatusServiceUnavailable)
)

