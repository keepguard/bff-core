package port

import "context"

type companyIDContextKey struct{}
type userIDContextKey struct{}
type bearerTokenContextKey struct{}

func WithCompanyID(ctx context.Context, companyID string) context.Context {
	if companyID == "" {
		return ctx
	}
	return context.WithValue(ctx, companyIDContextKey{}, companyID)
}

func CompanyIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(companyIDContextKey{}).(string)
	return v
}

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

func WithBearerToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, bearerTokenContextKey{}, token)
}

func BearerTokenFromContext(ctx context.Context) string {
	v, _ := ctx.Value(bearerTokenContextKey{}).(string)
	return v
}
