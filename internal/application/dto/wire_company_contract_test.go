package dto

import (
	"encoding/json"
	"testing"
)

// Contrato com o ms-company-go: a resposta (e o JSON do cache Redis) ganhou
// documentType, document e product, e a empresa pode ser PF (cnpj ""). O
// decoder ignora campo novo e não pode quebrar com data null.
func TestMSCompanyResponseDecodesIndividualCompany(t *testing.T) {
	raw := `{"id":"2b7e1c1e-7a52-4f5b-9a3e-1f2d3c4b5a69","codeCompany":"f7fc7350-b9fc-4e54-9c58-ac9385b23ae4",
		"tenantId":"0d9a4c3e-2b1a-4f6e-8d7c-5b4a3c2d1e0f","name":"Achadinhos","legalName":"Maria da Silva",
		"cnpj":"","documentType":"CPF","document":"12345678909","product":"achadinhos",
		"stateRegistration":null,"municipalRegistration":null,"address":null,"contacts":null,"representatives":null,
		"bankAccount":null,"taxRegime":"PESSOA_FISICA","cnaes":null,"mfaChannels":[],"ein":null,"status":"ACTIVE",
		"createdAt":"2026-10-10T12:00:00.123456","updatedAt":null}`

	var c MSCompanyResponseDTO
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatalf("empresa PF não decodificou: %v", err)
	}
	if c.Name != "Achadinhos" || c.Status != "ACTIVE" || c.CNPJ != "" {
		t.Fatalf("campos errados: %+v", c)
	}
	if c.CreatedAt.Year() != 2026 {
		t.Fatalf("createdAt não lido: %v", c.CreatedAt)
	}
	if !c.UpdatedAt.IsZero() {
		t.Fatalf("updatedAt null deve virar data zero, veio %v", c.UpdatedAt)
	}
}

func TestCustomTimeRejectsNonString(t *testing.T) {
	var ct CustomTime
	if err := json.Unmarshal([]byte(`1`), &ct); err == nil {
		t.Fatal("número de 1 byte deveria dar erro, não panic")
	}
}
