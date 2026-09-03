package client

import "context"

// ServiceTokenClient emite JWT OAuth de serviço (client_credentials) por company.
type ServiceTokenClient interface {
	GetToken(ctx context.Context, companyID string) (string, error)
}
