package dto

// SendMessageResponseDTO representa a resposta do ms-communication
type SendMessageResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
