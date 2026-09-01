package pkg

import "strings"

var errorCodeMessages = map[string]string{
	"USER_NOT_FOUND":      "Usuário não encontrado",
	"INVALID_PASSWORD":    "Senha inválida",
	"INVALID_CREDENTIALS": "Credenciais inválidas",
	"USER_NOT_ACTIVE":     "Usuário não está ativo",
	"EMAIL_NOT_VERIFIED":  "E-mail não verificado",
	"INVALID_TOKEN":       "Token inválido ou expirado",
	"TOKEN_REVOKED":       "Sessão revogada ou expirada. Por favor, realize login novamente.",
	"RESOURCE_NOT_FOUND":  "Recurso não encontrado",
	"ALREADY_EXISTS":      "Recurso já existe",
	"COMPANY_NOT_FOUND":   "Empresa não encontrada",
	"FORBIDDEN":           "Acesso negado",
	"CONFLICT":            "Conflito com recurso existente",
	"NOT_FOUND":           "Recurso não encontrado",
}

var englishDetailMessages = map[string]string{
	"User not found":        "Usuário não encontrado",
	"Invalid password":      "Senha inválida",
	"User is not active":    "Usuário não está ativo",
	"Email not verified":    "E-mail não verificado",
	"Not Found":             "Recurso não encontrado",
	"Conflict":              "Conflito com recurso existente",
	"Bad Request":           "Dados de entrada inválidos",
	"Unauthorized":          "Não autorizado",
	"Forbidden":             "Acesso negado",
	"agent not found":       "Agent não encontrado",
	"data source not found": "Fonte de dados não encontrada",
}

var serviceLabels = map[string]string{
	"collector service":        "serviço de coleta",
	"auth service":             "serviço de autenticação",
	"user service":             "serviço de usuários",
	"company service":          "serviço de empresas",
	"communication service":    "serviço de comunicação",
	"ms-communication service": "serviço de comunicação",
	"user consents service":    "serviço de consentimentos",
	"guardian service":         "serviço Guardian",
	"audit service":            "serviço de auditoria",
	"knowledge service":        "serviço de conhecimento",
	"ms-user":                  "serviço de usuários",
}

// ResolveUserMessage devolve mensagem em português para exibição ao usuário.
func ResolveUserMessage(errorCode, rawMessage string) string {
	raw := strings.TrimSpace(rawMessage)
	if raw != "" {
		if msg, ok := englishDetailMessages[raw]; ok {
			return msg
		}
		return raw
	}

	code := strings.ToUpper(strings.TrimSpace(errorCode))
	if code != "" && code != "UNKNOWN_ERROR" && code != "HTTP_ERROR" {
		if msg, ok := errorCodeMessages[code]; ok {
			return msg
		}
	}
	return ""
}

// LocalizeServiceName traduz identificadores técnicos de serviço para português.
func LocalizeServiceName(serviceName string) string {
	name := strings.TrimSpace(serviceName)
	if label, ok := serviceLabels[name]; ok {
		return label
	}
	return name
}
