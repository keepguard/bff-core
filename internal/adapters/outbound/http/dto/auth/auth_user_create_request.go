package dto

// AuthUserCreateRequestDTO representa a requisição para criar usuário no ms-auth
type AuthUserCreateRequestDTO struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	IDUserExternal string `json:"id_user_external"`
	CodeUser       string `json:"code_user"`
	CompanyID      string `json:"company_id"`
	CompanyCode    string `json:"company_code"`
	TenantId   string `json:"tenant_id"`
}
