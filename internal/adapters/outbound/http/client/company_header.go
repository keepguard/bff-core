package client

import (
	"context"

	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
)

func companyHeader(ctx context.Context) string {
	return domainclient.CompanyIDFromContext(ctx)
}
