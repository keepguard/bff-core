package client

import (
	"errors"
	"net/http"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
)

func TestMapHTTPError_WithMSResponse(t *testing.T) {
	// Arrange
	body := []byte(`{"error":"VALIDATION_ERROR","message":"Token inválido","details":"Token expirado"}`)
	serviceName := "ms-user"

	// Act
	err := MapHTTPError(400, body, serviceName)

	// Assert
	assert.IsType(t, &appdto.HTTPError{}, err)
	httpErr := err.(*appdto.HTTPError)
	assert.Equal(t, 400, httpErr.Code)
	assert.Equal(t, "Token inválido", httpErr.Message)
	assert.Equal(t, string(body), httpErr.Details)
}

func TestMapHTTPError_DefaultMessage(t *testing.T) {
	// Arrange
	body := []byte(`{"error":"UNKNOWN"}`)
	serviceName := "ms-user"

	// Act
	err := MapHTTPError(400, body, serviceName)

	// Assert
	httpErr := err.(*appdto.HTTPError)
	assert.Equal(t, 400, httpErr.Code)
	assert.Equal(t, "Dados de entrada inválidos", httpErr.Message)
	assert.Equal(t, string(body), httpErr.Details)
}

func TestMapHTTPError_EmptyBody(t *testing.T) {
	// Arrange
	body := []byte(``)
	serviceName := "ms-user"

	// Act
	err := MapHTTPError(404, body, serviceName)

	// Assert
	httpErr := err.(*appdto.HTTPError)
	assert.Equal(t, 404, httpErr.Code)
	assert.Equal(t, "Recurso não encontrado", httpErr.Message)
	assert.Equal(t, "", httpErr.Details)
}

func TestMapHTTPError_MaxAttemptsExceeded(t *testing.T) {
	// Arrange - Simula resposta do ms-user quando max tentativas é excedido
	body := []byte(`{"error":"VALIDATION_ERROR","message":"Número máximo de tentativas excedido. Por favor, inicie o registro novamente.","details":"","traceId":"test-123"}`)
	serviceName := "ms-user"

	// Act
	err := MapHTTPError(400, body, serviceName)

	// Assert
	httpErr := err.(*appdto.HTTPError)
	assert.Equal(t, 400, httpErr.Code)
	assert.Equal(t, "Número máximo de tentativas excedido. Por favor, inicie o registro novamente.", httpErr.Message)
	assert.Equal(t, string(body), httpErr.Details)
}

func TestMapHTTPError_InvalidToken(t *testing.T) {
	// Arrange - Simula resposta do ms-user quando token é inválido
	body := []byte(`{"error":"VALIDATION_ERROR","message":"Token inválido. Tentativas restantes: 2","details":"","traceId":"test-123"}`)
	serviceName := "ms-user"

	// Act
	err := MapHTTPError(400, body, serviceName)

	// Assert
	httpErr := err.(*appdto.HTTPError)
	assert.Equal(t, 400, httpErr.Code)
	assert.Equal(t, "Token inválido. Tentativas restantes: 2", httpErr.Message)
	assert.Equal(t, string(body), httpErr.Details)
}

func TestGetDefaultMessageForStatus_4xx(t *testing.T) {
	testCases := []struct {
		statusCode int
		expected   string
	}{
		{http.StatusBadRequest, "Dados de entrada inválidos"},
		{http.StatusUnauthorized, "Não autorizado"},
		{http.StatusForbidden, "Acesso negado"},
		{http.StatusNotFound, "Recurso não encontrado"},
		{http.StatusConflict, "Conflito com recurso existente"},
		{http.StatusUnprocessableEntity, "Dados não podem ser processados"},
		{http.StatusTooManyRequests, "Muitas requisições"},
	}

	for _, tc := range testCases {
		t.Run(http.StatusText(tc.statusCode), func(t *testing.T) {
			// Act
			message := getDefaultMessageForStatus(tc.statusCode, "ms-user")

			// Assert
			assert.Equal(t, tc.expected, message)
		})
	}
}

func TestGetDefaultMessageForStatus_5xx(t *testing.T) {
	testCases := []struct {
		statusCode int
		expected   string
	}{
		{http.StatusInternalServerError, "ms-user encontrou um erro interno"},
		{http.StatusBadGateway, "ms-user temporariamente indisponível"},
		{http.StatusServiceUnavailable, "ms-user indisponível"},
		{http.StatusGatewayTimeout, "ms-user não respondeu a tempo"},
	}

	for _, tc := range testCases {
		t.Run(http.StatusText(tc.statusCode), func(t *testing.T) {
			// Act
			message := getDefaultMessageForStatus(tc.statusCode, "ms-user")

			// Assert
			assert.Equal(t, tc.expected, message)
		})
	}
}

func TestGetDefaultMessageForStatus_Unknown(t *testing.T) {
	// Act
	message := getDefaultMessageForStatus(999, "ms-user")

	// Assert
	assert.Equal(t, "ms-user retornou erro 999", message)
}

func TestMapNetworkError(t *testing.T) {
	// Arrange
	originalErr := errors.New("connection refused")
	serviceName := "ms-user"

	// Act
	err := MapNetworkError(originalErr, serviceName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao comunicar com ms-user")
	assert.Contains(t, err.Error(), "connection refused")
	assert.ErrorIs(t, err, originalErr)
}
