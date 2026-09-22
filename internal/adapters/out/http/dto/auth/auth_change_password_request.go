package dto

// ChangePasswordMSRequestDTO representa a requisição de alteração de senha para o ms-auth
type ChangePasswordMSRequestDTO struct {
	CodeUser           string `json:"codeUser"`
	CurrentPassword    string `json:"currentPassword"`
	NewPassword        string `json:"newPassword"`
	ConfirmNewPassword string `json:"confirmNewPassword"`
}
