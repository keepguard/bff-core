package dto

type KnowledgeAskRequestDTO struct {
	Question string `json:"question"`
	Context  string `json:"context,omitempty"`
}

type OAuthClientCreateRequestDTO struct {
	ClientID        string `json:"clientId"`
	Description     string `json:"description,omitempty"`
	RoleID          string `json:"roleId"`
	TokenTTLSeconds *int   `json:"tokenTtlSeconds,omitempty"`
}

type OAuthClientUpdateRequestDTO struct {
	Description     string `json:"description,omitempty"`
	RoleID          string `json:"roleId"`
	TokenTTLSeconds *int   `json:"tokenTtlSeconds,omitempty"`
}
