package dto

// AuthRegisterLoginRequestDTO representa a requisição para login após registro com senha criptografada
type AuthRegisterLoginRequestDTO struct {
	Username     string `json:"username" validate:"required"`
	PasswordHash string `json:"passwordHash" validate:"required"`
	TenantId string `json:"tenantId" validate:"required"`
}
