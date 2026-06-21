package dto

// GenerateResetTokenMSResponseDTO representa a resposta de geração de token de reset do ms-auth
type GenerateResetTokenMSResponseDTO struct {
	CodeUser          string `json:"codeUser"`
	MessageType       string `json:"messageType"`
	CommunicationType string `json:"communicationType"`
	TemplateType      string `json:"templateType"`
	Token             string `json:"token"`
	ExpiresInSeconds  int64  `json:"expiresInSeconds"`
}
