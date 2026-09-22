package client

import (
	"context"

	domainclient "github.com/keepguard/bff-core/internal/application/port"
)

func companyHeader(ctx context.Context) string {
	return domainclient.CompanyIDFromContext(ctx)
}
