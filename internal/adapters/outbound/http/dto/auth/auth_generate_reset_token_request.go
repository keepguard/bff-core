package dto

// GenerateResetTokenMSRequestDTO representa a requisição de geração de token de reset para o ms-auth
type GenerateResetTokenMSRequestDTO struct {
	CodeUser          string `json:"codeUser"`
	MessageType       string `json:"messageType"`
	CommunicationType string `json:"communicationType"`
	TemplateType      string `json:"templateType"`
}
