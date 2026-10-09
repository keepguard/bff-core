package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
)

// MSErrorResponse representa a resposta de erro padrão dos microserviços
type MSErrorResponse struct {
	Error         string                 `json:"error"`
	Message       string                 `json:"message"`
	Details       string                 `json:"details,omitempty"`
	CorrelationID string                 `json:"correlationId,omitempty"`
	TraceID       string                 `json:"traceId,omitempty"`
	Errors        map[string]interface{} `json:"errors,omitempty"` // Erros de validação
}

// MapHTTPError converte resposta HTTP do MS em HTTPError tipado
func MapHTTPError(statusCode int, responseBody []byte, serviceName string) error {
	var msError MSErrorResponse
	_ = json.Unmarshal(responseBody, &msError)

	serviceLabel := pkg.LocalizeServiceName(serviceName)
	message := resolveUpstreamMessage(msError, statusCode, serviceLabel)

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

func resolveUpstreamMessage(msError MSErrorResponse, statusCode int, serviceLabel string) string {
	message := strings.TrimSpace(msError.Message)
	errorField := strings.TrimSpace(msError.Error)

	if message != "" {
		if translated := pkg.ResolveUserMessage("", message); translated != "" {
			return translated
		}
	}

	if errorField != "" {
		if translated := pkg.ResolveUserMessage(errorField, ""); translated != "" {
			return translated
		}
		if message == "" && !isTechnicalErrorCode(errorField) {
			if translated := pkg.ResolveUserMessage("", errorField); translated != "" {
				return translated
			}
		}
	}

	return getDefaultMessageForStatus(statusCode, serviceLabel)
}

func isTechnicalErrorCode(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r == ' ' {
			return false
		}
	}
	return true
}

// getDefaultMessageForStatus retorna mensagem padrão por status code
func getDefaultMessageForStatus(statusCode int, serviceName string) string {
	serviceLabel := pkg.LocalizeServiceName(serviceName)
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
		return serviceLabel + " encontrou um erro interno"
	case http.StatusBadGateway:
		return serviceLabel + " temporariamente indisponível"
	case http.StatusServiceUnavailable:
		return serviceLabel + " indisponível"
	case http.StatusGatewayTimeout:
		return serviceLabel + " não respondeu a tempo"
	default:
		return serviceLabel + " retornou erro " + fmt.Sprintf("%d", statusCode)
	}
}

// MapNetworkError converte erro de rede em HTTPError
func MapNetworkError(err error, serviceName string) error {
	return fmt.Errorf("erro ao comunicar com %s: %w", pkg.LocalizeServiceName(serviceName), err)
}
