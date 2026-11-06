package dto

// UserAccessMSResponseDTO é a resposta de GET /api/v1/users/code-user/:codeUser/access.
type UserAccessMSResponseDTO struct {
	CodeUser    string   `json:"codeUser"`
	CompanyID   string   `json:"companyId"`
	Active      bool     `json:"active"`
	Roles       []string `json:"roles"`
	Authorities []string `json:"authorities"`
}
