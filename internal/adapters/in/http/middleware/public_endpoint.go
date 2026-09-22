package http

import (
	"github.com/labstack/echo/v4"
)

// PublicEndpointConfig configuração para endpoints públicos
type PublicEndpointConfig struct {
	// PublicPaths define os caminhos que não requerem autenticação
	PublicPaths []string

	// SkipHeaderValidation se verdadeiro, não valida headers obrigatórios em rotas públicas
	SkipHeaderValidation bool
}

// PublicEndpointMiddleware middleware que marca endpoints como públicos
type PublicEndpointMiddleware struct {
	config PublicEndpointConfig
}

// NewPublicEndpointMiddleware cria um novo middleware de endpoints públicos
func NewPublicEndpointMiddleware(config PublicEndpointConfig) *PublicEndpointMiddleware {
	return &PublicEndpointMiddleware{
		config: config,
	}
}

// MarkAsPublic marca uma rota como pública (não requer autenticação)
func (m *PublicEndpointMiddleware) MarkAsPublic() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Marca o contexto como endpoint público
			c.Set("public_endpoint", true)
			return next(c)
		}
	}
}

// IsPublicEndpoint verifica se um endpoint é público
func IsPublicEndpoint(c echo.Context) bool {
	if publicEndpoint := c.Get("public_endpoint"); publicEndpoint != nil {
		if isPublic, ok := publicEndpoint.(bool); ok {
			return isPublic
		}
	}

	// Também verifica se o path está na lista de paths públicos
	return false
}

// ConditionalJWTMiddleware retorna um middleware JWT que só valida em rotas privadas
func ConditionalJWTMiddleware(jwtMiddleware *JWTMiddleware) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Se é um endpoint público, pula a validação do JWT
			if IsPublicEndpoint(c) {
				return next(c)
			}

			// Caso contrário, aplica o middleware JWT normalmente
			return jwtMiddleware.Middleware()(next)(c)
		}
	}
}

// PublicEndpoint é um helper para marcar rotas como públicas de forma fluente
type PublicEndpoint struct{}

// NewPublicEndpoint cria um novo marcador de endpoint público
func NewPublicEndpoint() *PublicEndpoint {
	return &PublicEndpoint{}
}

// Middleware retorna o middleware que marca a rota como pública
func (p *PublicEndpoint) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("public_endpoint", true)
			return next(c)
		}
	}
}
