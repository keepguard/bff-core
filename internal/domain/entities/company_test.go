package entities

import (
	"strings"
	"testing"

	"github.com/keepguard/bff-core/internal/domain/valueobjects"
)

func cnpjValido(t *testing.T) valueobjects.CNPJ {
	t.Helper()
	c, err := valueobjects.NewCNPJ("11222333000181")
	if err != nil {
		t.Fatalf("CNPJ de teste inválido: %v", err)
	}
	return c
}

func empresaValida(t *testing.T) *Company {
	t.Helper()
	c, err := NewCompany("tenant-1", "Empresa", "Empresa Ltda", cnpjValido(t))
	if err != nil {
		t.Fatalf("empresa válida foi recusada: %v", err)
	}
	return c
}

func TestNewCompany_Valida(t *testing.T) {
	cnpj := cnpjValido(t)

	c, err := NewCompany("tenant-1", "Empresa", "Empresa Ltda", cnpj)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if c.ID() == "" {
		t.Error("ID não foi gerado")
	}
	if c.TenantId() != "tenant-1" {
		t.Errorf("TenantId() = %q", c.TenantId())
	}
	if c.Name() != "Empresa" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.LegalName() != "Empresa Ltda" {
		t.Errorf("LegalName() = %q", c.LegalName())
	}
	if c.CNPJ().Value() != cnpj.Value() {
		t.Errorf("CNPJ() = %v", c.CNPJ().Value())
	}
	// empresa nasce ativa
	if c.Status() != CompanyStatusActive {
		t.Errorf("status inicial = %v, quer active", c.Status())
	}
	if c.CreatedAt().IsZero() || c.UpdatedAt().IsZero() {
		t.Error("timestamps não foram preenchidos")
	}
}

// Sem tenantId não há isolamento entre clientes — é o campo que separa os
// dados de uma empresa dos da outra.
func TestNewCompany_RecusaDadosInvalidos(t *testing.T) {
	cnpj := cnpjValido(t)

	casos := []struct {
		nome       string
		tenantId   string
		nomeEmp    string
		legalName  string
		querNoErro string
	}{
		{"sem tenantId", "", "Empresa", "Empresa Ltda", "tenantId"},
		{"nome vazio", "tenant-1", "", "Empresa Ltda", "name"},
		{"razão social vazia", "tenant-1", "Empresa", "", "legalName"},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			emp, err := NewCompany(c.tenantId, c.nomeEmp, c.legalName, cnpj)
			if err == nil {
				t.Fatal("aceitou dados inválidos")
			}
			if emp != nil {
				t.Error("devolveu empresa junto com o erro")
			}
			if !strings.Contains(err.Error(), c.querNoErro) {
				t.Errorf("erro = %q, esperava conter %q", err, c.querNoErro)
			}
		})
	}
}

func TestCompany_TransicoesDeStatus(t *testing.T) {
	casos := []struct {
		nome string
		muda func(*Company)
		quer CompanyStatus
	}{
		{"activate", func(c *Company) { c.Activate() }, CompanyStatusActive},
		{"deactivate", func(c *Company) { c.Deactivate() }, CompanyStatusInactive},
		{"block", func(c *Company) { c.Block() }, CompanyStatusBlocked},
		{"setPending", func(c *Company) { c.SetPending() }, CompanyStatusPending},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			emp := empresaValida(t)
			antes := emp.UpdatedAt()
			c.muda(emp)

			if emp.Status() != c.quer {
				t.Errorf("status = %v, quer %v", emp.Status(), c.quer)
			}
			if emp.UpdatedAt().Before(antes) {
				t.Error("updatedAt andou para trás")
			}
		})
	}
}

func TestCompany_UpdateName(t *testing.T) {
	t.Run("nome válido", func(t *testing.T) {
		c := empresaValida(t)
		if err := c.UpdateName("Outro Nome"); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if c.Name() != "Outro Nome" {
			t.Errorf("Name() = %q", c.Name())
		}
	})

	t.Run("vazio é recusado e não altera", func(t *testing.T) {
		c := empresaValida(t)
		original := c.Name()
		if err := c.UpdateName(""); err == nil {
			t.Fatal("aceitou nome vazio")
		}
		if c.Name() != original {
			t.Errorf("nome mudou para %q apesar do erro", c.Name())
		}
	})
}

func TestCompany_UpdateLegalName(t *testing.T) {
	t.Run("razão social válida", func(t *testing.T) {
		c := empresaValida(t)
		if err := c.UpdateLegalName("Outra Ltda"); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if c.LegalName() != "Outra Ltda" {
			t.Errorf("LegalName() = %q", c.LegalName())
		}
	})

	t.Run("vazia é recusada e não altera", func(t *testing.T) {
		c := empresaValida(t)
		original := c.LegalName()
		if err := c.UpdateLegalName(""); err == nil {
			t.Fatal("aceitou razão social vazia")
		}
		if c.LegalName() != original {
			t.Errorf("razão social mudou para %q apesar do erro", c.LegalName())
		}
	})
}

func TestCompany_UpdateCNPJ(t *testing.T) {
	c := empresaValida(t)
	novo, err := valueobjects.NewCNPJ("33445566000186")
	if err != nil {
		t.Fatalf("CNPJ de teste inválido: %v", err)
	}

	if err := c.UpdateCNPJ(novo); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if c.CNPJ().Value() != novo.Value() {
		t.Errorf("CNPJ() = %v, quer %v", c.CNPJ().Value(), novo.Value())
	}
}
