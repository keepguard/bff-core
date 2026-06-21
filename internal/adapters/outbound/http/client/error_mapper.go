package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

// MSErrorResponse representa a resposta de erro padrão dos microserviços
type MSErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details string                 `json:"details,omitempty"`
	TraceID string                 `json:"traceId,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"` // Erros de validação
}

// MapHTTPError converte resposta HTTP do MS em HTTPError tipado
func MapHTTPError(statusCode int, responseBody []byte, serviceName string) error {
	// Tenta fazer parse do erro do MS
	var msError MSErrorResponse
	_ = json.Unmarshal(responseBody, &msError)

	// Mensagem padrão se não conseguir fazer parse
	message := msError.Message
	if message == "" {
		message = getDefaultMessageForStatus(statusCode, serviceName)
	}

	// Se houver erros de validação, incluir na mensagem
	if len(msError.Errors) > 0 {
		for field, fieldError := range msError.Errors {
			if errorMsg, ok := fieldError.(string); ok {
				message = fmt.Sprintf("%s - %s: %s", message, field, errorMsg)
			}
		}
	}

	return &appdto.HTTPError{
		Code:    statusCode,
		Message: message,
		Details: string(responseBody),
	}
}

// getDefaultMessageForStatus retorna mensagem padrão por status code
func getDefaultMessageForStatus(statusCode int, serviceName string) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Dados de entrada inválidos"
	case http.StatusUnauthorized:
		return "Não autorizado"
	case http.StatusForbidden:
		return "Acesso negado"
	case http.StatusNotFound:
		return "Recurso não encontrado"
	case http.StatusConflict:
		return "Conflito com recurso existente"
	case http.StatusUnprocessableEntity:
		return "Dados não podem ser processados"
	case http.StatusTooManyRequests:
		return "Muitas requisições"
	case http.StatusInternalServerError:
		return serviceName + " encontrou um erro interno"
	case http.StatusBadGateway:
		return serviceName + " temporariamente indisponível"
	case http.StatusServiceUnavailable:
		return serviceName + " indisponível"
	case http.StatusGatewayTimeout:
		return serviceName + " não respondeu a tempo"
	default:
		return serviceName + " retornou erro " + fmt.Sprintf("%d", statusCode)
	}
}

// MapNetworkError converte erro de rede em HTTPError
func MapNetworkError(err error, serviceName string) error {
	return fmt.Errorf("erro ao comunicar com %s: %w", serviceName, err)
}
