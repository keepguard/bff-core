package rabbitmq

import (
	"encoding/json"
	"testing"

	outport "github.com/keepguard/bff-core/internal/application/port/out"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// msCommunicationMessage é cópia do rabbitMessage do consumidor
// (ms-communication-go/internal/adapters/in/rabbitmq/messagesend/message_send_consumer.go).
// O consumidor descarta a mensagem quando companyId, recipient ou
// communicationType vêm vazios.
type msCommunicationMessage struct {
	CompanyID         string         `json:"companyId"`
	MessageType       string         `json:"messageType"`
	Recipient         string         `json:"recipient"`
	TemplateType      string         `json:"templateType"`
	CommunicationType string         `json:"communicationType"`
	CodeUser          string         `json:"codeUser"`
	Variables         map[string]any `json:"variables"`
	XCorrelationID    string         `json:"xCorrelationId"`
}

func TestMessageDTO_ContratoComConsumidorDoMsCommunication(t *testing.T) {
	published := outport.MessageDTO{
		CompanyID:         "f7fc7350-b9fc-4e54-9c58-ac9385b23ae4",
		XCorrelationID:    "corr-1",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "AUTENTICACAO_EMAIL_TOKEN",
		Recipient:         "user@example.com",
		CodeUser:          "session-1",
		Variables:         map[string]interface{}{"token": "123456"},
	}

	body, err := json.Marshal(published)
	require.NoError(t, err)

	var consumed msCommunicationMessage
	require.NoError(t, json.Unmarshal(body, &consumed))

	assert.Equal(t, published.CompanyID, consumed.CompanyID)
	assert.Equal(t, "user@example.com", consumed.Recipient)
	assert.Equal(t, "EMAIL", consumed.CommunicationType)
	assert.Equal(t, "AUTENTICACAO_EMAIL_TOKEN", consumed.TemplateType)
	assert.Equal(t, "123456", consumed.Variables["token"])

	var raw map[string]any
	require.NoError(t, json.Unmarshal(body, &raw))
	assert.NotContains(t, raw, "tenantId", "o consumidor não lê tenantId")
}
