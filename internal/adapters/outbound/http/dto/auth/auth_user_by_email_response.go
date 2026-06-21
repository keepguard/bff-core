package dto

// UserByEmailResponseDTO representa a resposta de busca de usuário por email do ms-auth
type UserByEmailResponseDTO struct {
	ID            string `json:"id"`
	CodeUser      string `json:"codeUser"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"emailVerified"`
}
