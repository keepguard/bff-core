package dto

// MeProfileResponseDTO é o contrato público de GET /users/me.
// Não inclui CPF completo, RG, renda, PEP nem dados de KYC.
type MeProfileResponseDTO struct {
	Email           string              `json:"email"`
	PhoneE164       string              `json:"phoneE164,omitempty"`
	PreferredLocale string              `json:"preferredLocale,omitempty"`
	Timezone        string              `json:"timezone,omitempty"`
	AvatarURL       string              `json:"avatarUrl,omitempty"`
	DisplayHandle   string              `json:"displayHandle,omitempty"`
	Type            string              `json:"type,omitempty"`
	Status          string              `json:"status,omitempty"`
	CreatedAt       string              `json:"createdAt,omitempty"`
	PersonProfile   *MePersonProfileDTO `json:"personProfile,omitempty"`
}

// MePersonProfileDTO nome de exibição da pessoa física. hasCpf/cpfLast4; nunca o CPF cru.
type MePersonProfileDTO struct {
	FullName string `json:"fullName,omitempty"`
	HasCpf   bool   `json:"hasCpf"`
	CpfLast4 string `json:"cpfLast4,omitempty"`
}
