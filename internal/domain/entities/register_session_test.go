package entities

import (
	"strings"
	"testing"
	"time"

	"github.com/keepguard/bff-core/internal/domain/valueobjects"
)

func dadosValidos(t *testing.T) (valueobjects.Email, valueobjects.Phone, valueobjects.UserType) {
	t.Helper()
	email, err := valueobjects.NewEmail("pessoa@exemplo.com")
	if err != nil {
		t.Fatalf("email de teste inválido: %v", err)
	}
	phone, err := valueobjects.NewPhone("11999998888")
	if err != nil {
		t.Fatalf("telefone de teste inválido: %v", err)
	}
	userType, err := valueobjects.NewUserType("PERSON")
	if err != nil {
		t.Fatalf("userType de teste inválido: %v", err)
	}
	return email, phone, userType
}

func sessaoValida(t *testing.T) *RegisterSession {
	t.Helper()
	email, phone, userType := dadosValidos(t)
	rs, err := NewRegisterSession(email, "Fulano", "senha123", phone,
		true, "v1", true, "v1", userType)
	if err != nil {
		t.Fatalf("sessão válida foi recusada: %v", err)
	}
	return rs
}

func TestNewRegisterSession_Valida(t *testing.T) {
	rs := sessaoValida(t)

	if rs.ID() == "" {
		t.Error("ID não foi gerado")
	}
	if rs.Token() == "" {
		t.Error("token não foi gerado")
	}
	if !rs.IsPending() {
		t.Errorf("status inicial = %v, quer pending", rs.Status())
	}
	if rs.ConfirmedAt() != nil {
		t.Error("sessão nova não pode ter confirmedAt")
	}
	// a janela de confirmação é de 24h
	if d := time.Until(rs.ExpiresAt()); d < 23*time.Hour || d > 25*time.Hour {
		t.Errorf("expira em %v, esperava ~24h", d)
	}
}

// Dois cadastros seguidos não podem sair com o mesmo token: o token é o que
// autentica a confirmação por e-mail.
func TestNewRegisterSession_TokenEIDSaoUnicos(t *testing.T) {
	a := sessaoValida(t)
	b := sessaoValida(t)

	if a.Token() == b.Token() {
		t.Error("duas sessões saíram com o mesmo token")
	}
	if a.ID() == b.ID() {
		t.Error("duas sessões saíram com o mesmo ID")
	}
}

// Aceite de termos e de privacidade é exigência legal — não pode passar
// sessão sem os dois.
func TestNewRegisterSession_RecusaDadosInvalidos(t *testing.T) {
	email, phone, userType := dadosValidos(t)

	casos := []struct {
		nome            string
		nome_           string
		password        string
		termsAccepted   bool
		termsVersion    string
		privacyAccepted bool
		privacyVersion  string
		querNoErro      string
	}{
		{"nome vazio", "", "senha123", true, "v1", true, "v1", "name"},
		{"senha vazia", "Fulano", "", true, "v1", true, "v1", "password"},
		{"termos não aceitos", "Fulano", "senha123", false, "v1", true, "v1", "terms"},
		{"privacidade não aceita", "Fulano", "senha123", true, "v1", false, "v1", "privacy"},
		{"versão dos termos vazia", "Fulano", "senha123", true, "", true, "v1", "termsVersion"},
		{"versão da privacidade vazia", "Fulano", "senha123", true, "v1", true, "", "privacyVersion"},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rs, err := NewRegisterSession(email, c.nome_, c.password, phone,
				c.termsAccepted, c.termsVersion, c.privacyAccepted, c.privacyVersion, userType)
			if err == nil {
				t.Fatal("aceitou dados inválidos")
			}
			if rs != nil {
				t.Error("devolveu sessão junto com o erro")
			}
			if !strings.Contains(err.Error(), c.querNoErro) {
				t.Errorf("erro = %q, esperava conter %q", err, c.querNoErro)
			}
		})
	}
}

func TestRegisterSession_Confirm(t *testing.T) {
	rs := sessaoValida(t)

	if err := rs.Confirm(); err != nil {
		t.Fatalf("confirmação recusada: %v", err)
	}
	if !rs.IsConfirmed() {
		t.Errorf("status = %v, quer confirmed", rs.Status())
	}
	if rs.ConfirmedAt() == nil {
		t.Error("confirmedAt não foi preenchido")
	}
	if rs.IsPending() {
		t.Error("sessão confirmada não pode seguir pendente")
	}
}

// Confirmar duas vezes tem que falhar: senão um link de e-mail reenviado
// reconfirmaria um cadastro já cancelado.
func TestRegisterSession_ConfirmDuasVezesFalha(t *testing.T) {
	rs := sessaoValida(t)

	if err := rs.Confirm(); err != nil {
		t.Fatalf("primeira confirmação: %v", err)
	}
	err := rs.Confirm()
	if err == nil {
		t.Fatal("confirmou duas vezes")
	}
	if !strings.Contains(err.Error(), "not pending") {
		t.Errorf("erro = %q, esperava 'not pending'", err)
	}
}

func TestRegisterSession_ConfirmarCanceladaFalha(t *testing.T) {
	rs := sessaoValida(t)
	rs.Cancel()

	if err := rs.Confirm(); err == nil {
		t.Fatal("confirmou uma sessão cancelada")
	}
	if rs.IsConfirmed() {
		t.Error("sessão cancelada acabou confirmada")
	}
}

// Sessão expirada não confirma — é o que impede um link antigo de criar conta.
func TestRegisterSession_ConfirmarExpiradaFalha(t *testing.T) {
	rs := sessaoValida(t)
	rs.expiresAt = time.Now().Add(-1 * time.Minute)

	if !rs.IsExpired() {
		t.Fatal("IsExpired não detectou o vencimento")
	}
	err := rs.Confirm()
	if err == nil {
		t.Fatal("confirmou sessão expirada")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("erro = %q, esperava 'expired'", err)
	}
}

func TestRegisterSession_CancelEMarkAsExpired(t *testing.T) {
	t.Run("cancel", func(t *testing.T) {
		rs := sessaoValida(t)
		rs.Cancel()
		if rs.Status() != RegisterSessionStatusCancelled {
			t.Errorf("status = %v, quer cancelled", rs.Status())
		}
		if rs.IsPending() {
			t.Error("cancelada não pode seguir pendente")
		}
	})

	t.Run("markAsExpired", func(t *testing.T) {
		rs := sessaoValida(t)
		rs.MarkAsExpired()
		if rs.Status() != RegisterSessionStatusExpired {
			t.Errorf("status = %v, quer expired", rs.Status())
		}
	})
}

// Os getters expõem o que foi construído — vale uma passada para garantir que
// nenhum devolve o campo errado.
func TestRegisterSession_Getters(t *testing.T) {
	email, phone, userType := dadosValidos(t)
	rs, err := NewRegisterSession(email, "Fulano", "senha123", phone,
		true, "termos-v2", true, "privacidade-v3", userType)
	if err != nil {
		t.Fatalf("sessão recusada: %v", err)
	}

	if rs.Email().Value() != email.Value() {
		t.Errorf("Email() = %v, quer %v", rs.Email().Value(), email.Value())
	}
	if rs.Name() != "Fulano" {
		t.Errorf("Name() = %q", rs.Name())
	}
	if rs.Phone().Value() != phone.Value() {
		t.Errorf("Phone() = %v", rs.Phone().Value())
	}
	if !rs.TermsAccepted() {
		t.Error("TermsAccepted() = false")
	}
	if rs.TermsVersion() != "termos-v2" {
		t.Errorf("TermsVersion() = %q", rs.TermsVersion())
	}
	if !rs.PrivacyAccepted() {
		t.Error("PrivacyAccepted() = false")
	}
	if rs.PrivacyVersion() != "privacidade-v3" {
		t.Errorf("PrivacyVersion() = %q", rs.PrivacyVersion())
	}
	if rs.UserType().Value() != userType.Value() {
		t.Errorf("UserType() = %v", rs.UserType().Value())
	}
	if rs.CreatedAt().IsZero() {
		t.Error("CreatedAt() zerado")
	}
	if rs.UpdatedAt().IsZero() {
		t.Error("UpdatedAt() zerado")
	}
}
