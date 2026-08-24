package enums

// TemplateType representa os tipos de template suportados
type TemplateType string

const (
	// TemplateTypeAutenticacaoEmailToken representa template de token de autenticação via email
	TemplateTypeAutenticacaoEmailToken TemplateType = "AUTENTICACAO_EMAIL_TOKEN"

	// TemplateTypeAutenticacaoEmailTokenResend representa template de reenvio de token via email
	TemplateTypeAutenticacaoEmailTokenResend TemplateType = "AUTENTICACAO_EMAIL_TOKEN_RESEND"

	// TemplateTypeAutenticacaoSMSToken representa template de token de autenticação via SMS
	TemplateTypeAutenticacaoSMSToken TemplateType = "AUTENTICACAO_SMS_TOKEN"

	// TemplateTypeAutenticacaoWhatsAppToken representa template de token de autenticação via WhatsApp
	TemplateTypeAutenticacaoWhatsAppToken TemplateType = "AUTENTICACAO_WHATSAPP_TOKEN"

	// TemplateTypeCadastroSucesso representa template de sucesso no cadastro
	TemplateTypeCadastroSucesso TemplateType = "CADASTRO_SUCESSO"

	// TemplateTypeRecuperacaoSenha representa template de recuperação de senha
	TemplateTypeRecuperacaoSenha TemplateType = "RECUPERACAO_SENHA"

	// TemplateTypeSenhaAlteradaSucesso representa template de notificação de senha alterada com sucesso
	TemplateTypeSenhaAlteradaSucesso TemplateType = "SENHA_ALTERADA_SUCESSO"

	// TemplateTypeNovoDispositivoAutenticado representa template de alerta de novo dispositivo autenticado
	TemplateTypeNovoDispositivoAutenticado TemplateType = "NOVO_DISPOSITIVO_AUTENTICADO"

	// TemplateTypeNotificacaoGeral representa template de notificação geral
	TemplateTypeNotificacaoGeral TemplateType = "NOTIFICACAO_GERAL"

	// TemplateTypeAlertaSeguranca representa template de alerta de segurança
	TemplateTypeAlertaSeguranca TemplateType = "ALERTA_SEGURANCA"

	// TemplateTypeConfirmacaoAcao representa template de confirmação de ação
	TemplateTypeConfirmacaoAcao TemplateType = "CONFIRMACAO_ACAO"
)

// String retorna a representação em string do TemplateType
func (t TemplateType) String() string {
	return string(t)
}

// IsValid verifica se o TemplateType é válido
func (t TemplateType) IsValid() bool {
	switch t {
	case TemplateTypeAutenticacaoEmailToken, TemplateTypeAutenticacaoEmailTokenResend,
		TemplateTypeAutenticacaoSMSToken, TemplateTypeAutenticacaoWhatsAppToken,
		TemplateTypeCadastroSucesso, TemplateTypeRecuperacaoSenha,
		TemplateTypeSenhaAlteradaSucesso, TemplateTypeNovoDispositivoAutenticado,
		TemplateTypeNotificacaoGeral, TemplateTypeAlertaSeguranca,
		TemplateTypeConfirmacaoAcao:
		return true
	default:
		return false
	}
}
