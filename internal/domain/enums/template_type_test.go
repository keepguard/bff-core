package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTemplateType_String testa a conversão para string
func TestTemplateType_String(t *testing.T) {
	tests := []struct {
		name         string
		templateType TemplateType
		expected     string
	}{
		{"Autenticacao Email Token", TemplateTypeAutenticacaoEmailToken, "AUTENTICACAO_EMAIL_TOKEN"},
		{"Autenticacao SMS Token", TemplateTypeAutenticacaoSMSToken, "AUTENTICACAO_SMS_TOKEN"},
		{"Autenticacao WhatsApp Token", TemplateTypeAutenticacaoWhatsAppToken, "AUTENTICACAO_WHATSAPP_TOKEN"},
		{"Cadastro Sucesso", TemplateTypeCadastroSucesso, "CADASTRO_SUCESSO"},
		{"Recuperacao Senha", TemplateTypeRecuperacaoSenha, "RECUPERACAO_SENHA"},
		{"Notificacao Geral", TemplateTypeNotificacaoGeral, "NOTIFICACAO_GERAL"},
		{"Alerta Seguranca", TemplateTypeAlertaSeguranca, "ALERTA_SEGURANCA"},
		{"Confirmacao Acao", TemplateTypeConfirmacaoAcao, "CONFIRMACAO_ACAO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.templateType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTemplateType_IsValid testa a validação dos tipos
func TestTemplateType_IsValid(t *testing.T) {
	tests := []struct {
		name         string
		templateType TemplateType
		expected     bool
	}{
		{"Autenticacao Email Token válido", TemplateTypeAutenticacaoEmailToken, true},
		{"Autenticacao SMS Token válido", TemplateTypeAutenticacaoSMSToken, true},
		{"Autenticacao WhatsApp Token válido", TemplateTypeAutenticacaoWhatsAppToken, true},
		{"Cadastro Sucesso válido", TemplateTypeCadastroSucesso, true},
		{"Recuperacao Senha válido", TemplateTypeRecuperacaoSenha, true},
		{"Notificacao Geral válido", TemplateTypeNotificacaoGeral, true},
		{"Alerta Seguranca válido", TemplateTypeAlertaSeguranca, true},
		{"Confirmacao Acao válido", TemplateTypeConfirmacaoAcao, true},
		{"Tipo inválido", TemplateType("INVALID"), false},
		{"Tipo vazio", TemplateType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.templateType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}
