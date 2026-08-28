package dto

// MSUserCreateRequestDTO representa a requisição para criar um usuário no ms-user
type MSUserCreateRequestDTO struct {
	CompanyID       string             `json:"companyId"`
	Type            string             `json:"type"` // PERSON ou COMPANY
	Email           string             `json:"email"`
	PhoneE164       string             `json:"phoneE164,omitempty"`
	PreferredLocale string             `json:"preferredLocale,omitempty"`
	Timezone        string             `json:"timezone,omitempty"`
	AvatarURL       string             `json:"avatarUrl,omitempty"`
	Status          string             `json:"status,omitempty"`
	PersonProfile   *PersonProfileDTO  `json:"personProfile,omitempty"`
	CompanyProfile  *CompanyProfileDTO `json:"companyProfile,omitempty"`
}

// PersonProfileDTO representa o perfil de pessoa física
type PersonProfileDTO struct {
	FullName      string `json:"fullName"`
	DisplayHandle string `json:"display_handle,omitempty"`
	CPF           string `json:"cpf,omitempty"`
	RG            string `json:"rg,omitempty"`
	RGIssuer      string `json:"rg_issuer,omitempty"`
	RGState       string `json:"rg_state,omitempty"`
	DateOfBirth   string `json:"date_of_birth,omitempty"`
	Gender        string `json:"gender,omitempty"`
	MaritalStatus string `json:"marital_status,omitempty"`
	Nationality   string `json:"nationality,omitempty"`
	BirthCountry  string `json:"birth_country,omitempty"`
	BirthState    string `json:"birth_state,omitempty"`
	BirthCity     string `json:"birth_city,omitempty"`
	MotherName    string `json:"mother_name,omitempty"`
	FatherName    string `json:"father_name,omitempty"`
	PEP           bool   `json:"pep"`
	KYCStatus     string `json:"kyc_status,omitempty"`
	KYCLevel      string `json:"kyc_level,omitempty"`
	Occupation    string `json:"occupation,omitempty"`
	IncomeRange   string `json:"income_range,omitempty"`
}

// CompanyProfileDTO representa o perfil de pessoa jurídica
type CompanyProfileDTO struct {
	CompanyID                     string `json:"company_id"`
	LegalNameSnapshot             string `json:"legal_name_snapshot,omitempty"`
	CNPJSnapshot                  string `json:"cnpj_snapshot,omitempty"`
	StateRegistrationSnapshot     string `json:"state_registration_snapshot,omitempty"`
	MunicipalRegistrationSnapshot string `json:"municipal_registration_snapshot,omitempty"`
	RepresentativeName            string `json:"representative_name,omitempty"`
	RepresentativeCPF             string `json:"representative_cpf,omitempty"`
}
