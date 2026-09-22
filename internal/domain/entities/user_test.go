package entities

import (
	"strings"
	"testing"

	"github.com/keepguard/bff-core/internal/domain/valueobjects"
)

func usuarioValido(t *testing.T) *User {
	t.Helper()
	email, phone, userType := dadosValidos(t)
	u, err := NewUser("Fulano", email, phone, "empresa-1", userType)
	if err != nil {
		t.Fatalf("usuário válido foi recusado: %v", err)
	}
	return u
}

func TestNewUser_Valido(t *testing.T) {
	email, phone, userType := dadosValidos(t)

	u, err := NewUser("Fulano", email, phone, "empresa-1", userType)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if u.ID() == "" {
		t.Error("ID não foi gerado")
	}
	if u.Name() != "Fulano" {
		t.Errorf("Name() = %q", u.Name())
	}
	if u.CompanyID() != "empresa-1" {
		t.Errorf("CompanyID() = %q", u.CompanyID())
	}
	if u.Email().Value() != email.Value() {
		t.Errorf("Email() = %v", u.Email().Value())
	}
	if u.Phone().Value() != phone.Value() {
		t.Errorf("Phone() = %v", u.Phone().Value())
	}
	if u.UserType().Value() != userType.Value() {
		t.Errorf("UserType() = %v", u.UserType().Value())
	}
	if u.CreatedAt().IsZero() || u.UpdatedAt().IsZero() {
		t.Error("timestamps não foram preenchidos")
	}
}

// Usuário sem empresa não pode existir: é o que garante o isolamento por
// tenant em todo o resto do sistema.
func TestNewUser_RecusaDadosInvalidos(t *testing.T) {
	email, phone, userType := dadosValidos(t)

	casos := []struct {
		nome       string
		nomeUser   string
		companyID  string
		querNoErro string
	}{
		{"nome vazio", "", "empresa-1", "name"},
		{"sem empresa", "Fulano", "", "companyID"},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			u, err := NewUser(c.nomeUser, email, phone, c.companyID, userType)
			if err == nil {
				t.Fatal("aceitou dados inválidos")
			}
			if u != nil {
				t.Error("devolveu usuário junto com o erro")
			}
			if !strings.Contains(err.Error(), c.querNoErro) {
				t.Errorf("erro = %q, esperava conter %q", err, c.querNoErro)
			}
		})
	}
}

func TestUser_TransicoesDeStatus(t *testing.T) {
	casos := []struct {
		nome string
		muda func(*User)
		quer UserStatus
	}{
		{"activate", func(u *User) { u.Activate() }, UserStatusActive},
		{"deactivate", func(u *User) { u.Deactivate() }, UserStatusInactive},
		{"block", func(u *User) { u.Block() }, UserStatusBlocked},
		{"setPending", func(u *User) { u.SetPending() }, UserStatusPending},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			u := usuarioValido(t)
			antes := u.UpdatedAt()
			c.muda(u)

			if u.Status() != c.quer {
				t.Errorf("status = %v, quer %v", u.Status(), c.quer)
			}
			if u.UpdatedAt().Before(antes) {
				t.Error("updatedAt andou para trás")
			}
		})
	}
}

// Bloquear e reativar tem que funcionar: é o fluxo de suspensão por
// inadimplência e retorno depois do pagamento.
func TestUser_BloqueadoPodeSerReativado(t *testing.T) {
	u := usuarioValido(t)

	u.Block()
	if u.Status() != UserStatusBlocked {
		t.Fatalf("não bloqueou: %v", u.Status())
	}

	u.Activate()
	if u.Status() != UserStatusActive {
		t.Errorf("não reativou: %v", u.Status())
	}
}

func TestUser_UpdateName(t *testing.T) {
	t.Run("nome válido", func(t *testing.T) {
		u := usuarioValido(t)
		if err := u.UpdateName("Novo Nome"); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if u.Name() != "Novo Nome" {
			t.Errorf("Name() = %q", u.Name())
		}
	})

	t.Run("nome vazio é recusado e não altera", func(t *testing.T) {
		u := usuarioValido(t)
		original := u.Name()

		if err := u.UpdateName(""); err == nil {
			t.Fatal("aceitou nome vazio")
		}
		if u.Name() != original {
			t.Errorf("nome mudou para %q apesar do erro", u.Name())
		}
	})
}

func TestUser_UpdateEmail(t *testing.T) {
	u := usuarioValido(t)
	novo, err := valueobjects.NewEmail("outro@exemplo.com")
	if err != nil {
		t.Fatalf("email de teste inválido: %v", err)
	}

	if err := u.UpdateEmail(novo); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if u.Email().Value() != novo.Value() {
		t.Errorf("Email() = %v, quer %v", u.Email().Value(), novo.Value())
	}
}
