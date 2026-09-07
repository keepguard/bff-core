package servicetoken

import (
	"context"

	"github.com/keepguard/bff-core/internal/adapters/outbound/http/decorator/observe"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

type decorated struct {
	inner port.ServiceTokenClient
	cfg   observe.Config
}

func New(inner port.ServiceTokenClient, logger *zap.Logger, metrics *metrics.Metrics, serviceName string) port.ServiceTokenClient {
	if inner == nil {
		return nil
	}
	return &decorated{inner: inner, cfg: observe.Config{Logger: logger, Metrics: metrics, ServiceName: serviceName, Retry: observe.DefaultRetry()}}
}

func (d *decorated) GetToken(ctx context.Context, companyID string) (string, error) {
	token, err := observe.Call(d.cfg, "GetToken", "POST", "/oauth/token", companyID, func() (string, error) {
		return d.inner.GetToken(ctx, companyID)
	})
	if err != nil {
		return "", err
	}
	return token, nil
}
