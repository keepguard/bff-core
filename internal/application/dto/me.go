package dto

type GetMeQuery struct {
	TenantID      string
	CorrelationID string
	Token         string
	CodeUser      string
}

type MeProfileViewDTO struct {
	Email           string
	PhoneE164       string
	PreferredLocale string
	Timezone        string
	AvatarURL       string
	DisplayHandle   string
	Type            string
	Status          string
	CreatedAt       string
	PersonProfile   *MePersonProfileViewDTO
}

type MePersonProfileViewDTO struct {
	FullName string
}
