package dto

import (
	"time"
)

// CustomTime é um tipo customizado para fazer parsing de datas sem timezone
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implementa o unmarshal customizado para datas
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1] // Remove aspas

	// Tenta fazer parse com diferentes formatos
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05.99999",
		"2006-01-02T15:04:05.9999",
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05.99",
		"2006-01-02T15:04:05.9",
		"2006-01-02T15:04:05",
	}

	var err error
	for _, format := range formats {
		ct.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return err
}

// CompanyMfaChannelDTO representa a configuração de canal de MFA da empresa
type CompanyMfaChannelDTO struct {
	ID       string `json:"id"`
	Channel  string `json:"channel"`
	Required bool   `json:"required"`
	Enabled  bool   `json:"enabled"`
}

// MSCompanyResponseDTO representa a resposta com dados da empresa do ms-company
type MSCompanyResponseDTO struct {
	ID          string                 `json:"id"`
	CodeCompany string                 `json:"codeCompany"`
	Name        string                 `json:"name"`
	LegalName   string                 `json:"legalName"`
	CNPJ        string                 `json:"cnpj"`
	Status      string                 `json:"status"`
	MfaChannels []CompanyMfaChannelDTO `json:"mfaChannels"`
	CreatedAt   CustomTime             `json:"createdAt"`
	UpdatedAt   CustomTime             `json:"updatedAt"`
	TenantId    string                 `json:"tenantId"`
}
