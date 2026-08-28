package client

import "context"

type companyIDContextKey struct{}

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
