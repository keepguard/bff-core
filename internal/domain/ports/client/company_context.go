package client

import "context"

type companyIDContextKey struct{}
type userIDContextKey struct{}
type bearerTokenContextKey struct{}

// WithCompanyID anexa o companyId resolvido ao contexto da chamada interna.
func WithCompanyID(ctx context.Context, companyID string) context.Context {
	if companyID == "" {
		return ctx
	}
	return context.WithValue(ctx, companyIDContextKey{}, companyID)
}

// CompanyIDFromContext extrai o companyId anexado ao contexto.
func CompanyIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(companyIDContextKey{}).(string)
	return v
}

// WithUserID anexa o usuário autenticado para auditoria no collector.
func WithUserID(ctx context.Context, userID string) context.Context {
	if userID == "" {
		return ctx
	}
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIDContextKey{}).(string)
	return v
}

// WithBearerToken anexa o JWT inbound para encaminhar a resource servers.
func WithBearerToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, bearerTokenContextKey{}, token)
}

// BearerTokenFromContext extrai o JWT inbound anexado ao contexto.
func BearerTokenFromContext(ctx context.Context) string {
	v, _ := ctx.Value(bearerTokenContextKey{}).(string)
	return v
}
