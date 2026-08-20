package messaging

import "context"

// MessagePublisher interface para publicação de mensagens
type MessagePublisher interface {
	// PublishMessage publica uma mensagem na fila
	PublishMessage(ctx context.Context, message MessageDTO) error
	// Close fecha a conexão com o broker de mensageria
	Close() error
}

// MessageDTO representa uma mensagem a ser enviada
type MessageDTO struct {
	TenantId      string                 `json:"tenantId"`
	XCorrelationID    string                 `json:"xCorrelationId"`
	MessageType       string                 `json:"messageType"`
	CommunicationType string                 `json:"communicationType"`
	TemplateType      string                 `json:"templateType"`
	Recipient         string                 `json:"recipient"`
	Subject           string                 `json:"subject,omitempty"`
	Content           string                 `json:"content,omitempty"`
	CodeUser          string                 `json:"codeUser,omitempty"`
	Variables         map[string]interface{} `json:"variables"`
}
