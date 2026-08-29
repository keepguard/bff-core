package dto

type PaginatedGuardianIncidents struct {
	Content       []map[string]any `json:"content"`
	Page          int              `json:"page"`
	Size          int              `json:"size"`
	TotalElements int64            `json:"totalElements"`
	TotalPages    int              `json:"totalPages"`
}

type GuardianExecuteActionRequest struct {
	SuggestionID string `json:"suggestionId"`
	Confirmation string `json:"confirmation,omitempty"`
}

type GuardianRecipientUpsertRequest struct {
	Email   string `json:"email"`
	Label   string `json:"label,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}
