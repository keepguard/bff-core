package http

import (
	"strings"

	libauth "github.com/keepguard/lib-go-common/auth"
)

// jwksKeys liga a validação EdDSA quando BFF_CORE_JWT_JWKS_URL existe (chaves
// públicas do ms-auth). Vazio = só HS256, como antes.
func jwksKeys(url string) []libauth.KeySource {
	if strings.TrimSpace(url) == "" {
		return nil
	}
	return []libauth.KeySource{libauth.NewJWKSClient(strings.TrimSpace(url))}
}
