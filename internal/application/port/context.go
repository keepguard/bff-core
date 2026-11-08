package port

import "context"

type companyIDContextKey struct{}
type userIDContextKey struct{}
type bearerTokenContextKey struct{}
type deviceIDContextKey struct{}

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

// WithDeviceID guarda o X-Device-Id do pedido (aparelho em que a pessoa confirmou o cadastro), que o ms-auth usa para
// não pedir de novo o código de aparelho logo após o cadastro.
func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	if deviceID == "" {
		return ctx
	}
	return context.WithValue(ctx, deviceIDContextKey{}, deviceID)
}

func DeviceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(deviceIDContextKey{}).(string)
	return v
}
