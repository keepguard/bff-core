package pkg

import "strings"

// NormalizeRole remove prefixo ROLE_ e padroniza maiúsculas.
func NormalizeRole(role string) string {
	normalized := strings.ToUpper(strings.TrimSpace(role))
	return strings.TrimPrefix(normalized, "ROLE_")
}

// HasAnyRole verifica se o usuário possui qualquer uma das roles pedidas.
func HasAnyRole(have []string, want ...string) bool {
	if len(have) == 0 || len(want) == 0 {
		return false
	}
	wanted := make(map[string]struct{}, len(want))
	for _, role := range want {
		if n := NormalizeRole(role); n != "" {
			wanted[n] = struct{}{}
		}
	}
	for _, role := range have {
		if _, ok := wanted[NormalizeRole(role)]; ok {
			return true
		}
	}
	return false
}

func HasAuthority(have []string, want string) bool {
	wanted := strings.TrimSpace(want)
	if wanted == "" {
		return false
	}
	for _, item := range have {
		if strings.TrimSpace(item) == wanted {
			return true
		}
	}
	return false
}
