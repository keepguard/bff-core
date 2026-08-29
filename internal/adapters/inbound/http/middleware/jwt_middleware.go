package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// JWTMiddleware implementa middleware JWT com validação local
type JWTMiddleware struct {
	jwtSecret string
	logger    *zap.Logger
}

// NewJWTMiddleware cria um novo middleware JWT
func NewJWTMiddleware(jwtSecret string, logger *zap.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// Middleware retorna o middleware JWT que valida o token LOCALMENTE
func (j *JWTMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := GetCorrelationID(c)
			tenantIdHeader := GetTenantId(c)

			token, err := extractToken(c)
			if err != nil {
				j.logger.Error("Erro ao extrair token JWT",
					zap.String("correlationId", correlationID),
					zap.Error(err),
				)
				return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
					Error:         "UNAUTHORIZED",
					Message:       "Token de autorização não fornecido ou inválido",
					CorrelationID: correlationID,
				})
			}

			c.Set("token", token)

			claims, err := j.validateTokenLocal(token, tenantIdHeader)
			if err != nil {
				j.logger.Error("Token JWT inválido (validação local)",
					zap.String("correlationId", correlationID),
					zap.Error(err),
				)
				return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
					Error:         "UNAUTHORIZED",
					Message:       "Token inválido ou expirado",
					CorrelationID: correlationID,
				})
			}

			if ResolveTenantId(c, claims) == "" {
				j.logger.Warn("tenant_id ausente no JWT e no header",
					zap.String("correlationId", correlationID))
				return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
					Error:         "MISSING_TENANT",
					Message:       "tenant_id do token JWT é obrigatório",
					CorrelationID: correlationID,
				})
			}

			c.Set("claims", claims)

			if claims.CodeUser != "" {
				SetUserID(c, claims.CodeUser)
			} else if claims.Sub != "" {
				SetUserID(c, claims.Sub)
			} else if claims.UserID != "" {
				SetUserID(c, claims.UserID)
			}

			j.logger.Debug("Token validado localmente",
				zap.String("correlationId", correlationID),
				zap.String("userId", GetUserID(c)),
			)

			return next(c)
		}
	}
}

// validateTokenLocal valida JWT localmente usando secret compartilhado
func (j *JWTMiddleware) validateTokenLocal(tokenString, tenantIdHeader string) (*pkg.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inválido: %v", token.Header["alg"])
		}
		return []byte(j.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao parsear token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("claims inválidos")
	}

	if exp, ok := mapClaims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return nil, fmt.Errorf("token expirado")
		}
	}

	claims := &pkg.JWTClaims{}
	if codeUser, ok := mapClaims["codeUser"].(string); ok {
		claims.CodeUser = codeUser
	}
	if sub, ok := mapClaims["sub"].(string); ok {
		claims.Sub = sub
	}
	if username, ok := mapClaims["username"].(string); ok {
		claims.Username = username
	}
	if xApp, ok := mapClaims["tenant_id"].(string); ok {
		claims.TenantId = xApp
	} else if xApp, ok := mapClaims["tenantId"].(string); ok {
		claims.TenantId = xApp
	}
	if userID, ok := mapClaims["userId"].(string); ok {
		claims.UserID = userID
	}
	if email, ok := mapClaims["email"].(string); ok {
		claims.Email = email
	}
	claims.Roles = stringSliceFromClaim(mapClaims["roles"])

	if claims.TenantId != "" && tenantIdHeader != "" && claims.TenantId != tenantIdHeader {
		return nil, fmt.Errorf("tenantId mismatch: token=%s, header=%s", claims.TenantId, tenantIdHeader)
	}

	return claims, nil
}

// extractToken extrai o token JWT da requisição
func extractToken(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Token de autorização não fornecido")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Formato de token inválido")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Token vazio")
	}

	return token, nil
}

func stringSliceFromClaim(raw interface{}) []string {
	switch values := raw.(type) {
	case []string:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if value != "" {
				out = append(out, value)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if text, ok := value.(string); ok && text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

// GetTokenFromContext extrai o token do contexto
func GetTokenFromContext(c echo.Context) string {
	if token := c.Get("token"); token != nil {
		if tokenStr, ok := token.(string); ok {
			return tokenStr
		}
	}
	return ""
}

// GetClaimsFromContext extrai as claims do contexto
func GetClaimsFromContext(c echo.Context) *pkg.JWTClaims {
	if claims := c.Get("claims"); claims != nil {
		if jwtClaims, ok := claims.(*pkg.JWTClaims); ok {
			return jwtClaims
		}
	}
	return nil
}

// ResolveTenantId retorna o tenant do JWT; se ausente, usa o header X-Tenant-Id.
func ResolveTenantId(c echo.Context, claims *pkg.JWTClaims) string {
	if claims != nil && claims.TenantId != "" {
		return claims.TenantId
	}
	if claims == nil {
		if fromCtx := GetClaimsFromContext(c); fromCtx != nil && fromCtx.TenantId != "" {
			return fromCtx.TenantId
		}
	}
	return GetTenantId(c)
}

// GetUserIDFromContext extrai o ID do usuário do contexto
func GetUserIDFromContext(c echo.Context) string {
	if claims := GetClaimsFromContext(c); claims != nil {
		if claims.CodeUser != "" {
			return claims.CodeUser
		}
		if claims.Sub != "" {
			return claims.Sub
		}
		if claims.UserID != "" {
			return claims.UserID
		}
	}
	return GetUserID(c)
}

// GetUserFromContext extrai o usuário do contexto (compatibilidade)
func GetUserFromContext(c echo.Context) jwt.Claims {
	if user := c.Get("user"); user != nil {
		if claims, ok := user.(jwt.Claims); ok {
			return claims
		}
	}
	return nil
}
