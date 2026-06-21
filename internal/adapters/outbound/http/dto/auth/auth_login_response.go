package dto

// AuthLoginResponseDTO representa a resposta de login do ms-auth
type AuthLoginResponseDTO struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}
