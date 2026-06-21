package dto

// ResetPasswordMSRequestDTO representa a requisição de reset de senha para o ms-auth
type ResetPasswordMSRequestDTO struct {
	CodeUser           string `json:"codeUser"`
	ResetToken         string `json:"resetToken"`
	NewPassword        string `json:"newPassword"`
	ConfirmNewPassword string `json:"confirmNewPassword"`
	MessageType        string `json:"messageType"`
	TemplateType       string `json:"templateType"`
}
