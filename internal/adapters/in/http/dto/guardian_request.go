package dto

type GuardianExecuteActionRequestDTO struct {
	SuggestionID string `json:"suggestionId"`
	Confirmation string `json:"confirmation,omitempty"`
}

type GuardianRecipientUpsertRequestDTO struct {
	Email   string `json:"email"`
	Label   string `json:"label,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}
