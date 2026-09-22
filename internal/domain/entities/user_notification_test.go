package entities

import (
	"strings"
	"testing"
)

func TestNewUserNotification_Valida(t *testing.T) {
	un, err := NewUserNotification("user-1", true, false, true, false)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if un.ID() == "" {
		t.Error("ID não foi gerado")
	}
	if un.UserID() != "user-1" {
		t.Errorf("UserID() = %q", un.UserID())
	}
	if !un.EmailEnabled() {
		t.Error("EmailEnabled() = false, quer true")
	}
	if un.SmsEnabled() {
		t.Error("SmsEnabled() = true, quer false")
	}
	if !un.PushEnabled() {
		t.Error("PushEnabled() = false, quer true")
	}
	if un.WhatsAppEnabled() {
		t.Error("WhatsAppEnabled() = true, quer false")
	}
	if un.CreatedAt().IsZero() || un.UpdatedAt().IsZero() {
		t.Error("timestamps não foram preenchidos")
	}
}

// Preferência sem dono não faz sentido e viraria registro órfão.
func TestNewUserNotification_SemUserIDFalha(t *testing.T) {
	un, err := NewUserNotification("", true, true, true, true)

	if err == nil {
		t.Fatal("aceitou userID vazio")
	}
	if un != nil {
		t.Error("devolveu entidade junto com o erro")
	}
	if !strings.Contains(err.Error(), "userID") {
		t.Errorf("erro = %q, esperava conter 'userID'", err)
	}
}

// UpdatePreferences troca os quatro canais de uma vez — inclusive desligando
// os que estavam ligados.
func TestUserNotification_UpdatePreferences(t *testing.T) {
	un, err := NewUserNotification("user-1", true, true, true, true)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	antes := un.UpdatedAt()

	un.UpdatePreferences(false, true, false, true)

	if un.EmailEnabled() {
		t.Error("email deveria ter sido desligado")
	}
	if !un.SmsEnabled() {
		t.Error("sms deveria seguir ligado")
	}
	if un.PushEnabled() {
		t.Error("push deveria ter sido desligado")
	}
	if !un.WhatsAppEnabled() {
		t.Error("whatsapp deveria seguir ligado")
	}
	if un.UpdatedAt().Before(antes) {
		t.Error("updatedAt andou para trás")
	}
}

// Cada canal liga e desliga de forma independente: mexer em um não pode
// afetar os outros.
func TestUserNotification_CanaisSaoIndependentes(t *testing.T) {
	casos := []struct {
		nome      string
		liga      func(*UserNotification)
		desliga   func(*UserNotification)
		estaAtivo func(*UserNotification) bool
	}{
		{
			"email",
			func(u *UserNotification) { u.EnableEmail() },
			func(u *UserNotification) { u.DisableEmail() },
			func(u *UserNotification) bool { return u.EmailEnabled() },
		},
		{
			"sms",
			func(u *UserNotification) { u.EnableSMS() },
			func(u *UserNotification) { u.DisableSMS() },
			func(u *UserNotification) bool { return u.SmsEnabled() },
		},
		{
			"push",
			func(u *UserNotification) { u.EnablePush() },
			func(u *UserNotification) { u.DisablePush() },
			func(u *UserNotification) bool { return u.PushEnabled() },
		},
		{
			"whatsapp",
			func(u *UserNotification) { u.EnableWhatsApp() },
			func(u *UserNotification) { u.DisableWhatsApp() },
			func(u *UserNotification) bool { return u.WhatsAppEnabled() },
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			// começa com tudo desligado
			un, err := NewUserNotification("user-1", false, false, false, false)
			if err != nil {
				t.Fatalf("setup: %v", err)
			}

			c.liga(un)
			if !c.estaAtivo(un) {
				t.Errorf("%s não ligou", c.nome)
			}

			// os outros três seguem desligados
			ligados := 0
			for _, chk := range []func(*UserNotification) bool{
				func(u *UserNotification) bool { return u.EmailEnabled() },
				func(u *UserNotification) bool { return u.SmsEnabled() },
				func(u *UserNotification) bool { return u.PushEnabled() },
				func(u *UserNotification) bool { return u.WhatsAppEnabled() },
			} {
				if chk(un) {
					ligados++
				}
			}
			if ligados != 1 {
				t.Errorf("ligar %s deixou %d canais ativos, quer 1", c.nome, ligados)
			}

			c.desliga(un)
			if c.estaAtivo(un) {
				t.Errorf("%s não desligou", c.nome)
			}
		})
	}
}
