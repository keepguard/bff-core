package communication

import (
	"context"
	"net/http"
	"time"

	communicationDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/communication"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	portsclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
)

// communicationMetricsDecorator implementa métricas para CommunicationClient
type communicationMetricsDecorator struct {
	inner       portsclient.CommunicationClient
	metrics     *metrics.Metrics
	serviceName string
}

// NewCommunicationMetricsDecorator cria um decorator de métricas para CommunicationClient
func NewCommunicationMetricsDecorator(
	inner portsclient.CommunicationClient,
	metrics *metrics.Metrics,
	serviceName string,
) portsclient.CommunicationClient {
	return &communicationMetricsDecorator{
		inner:       inner,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

// getStatusCodeFromError extrai o status code de um erro
func (d *communicationMetricsDecorator) getStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.Code
	}

	return http.StatusInternalServerError
}

// SendNotification implementa SendNotification com métricas
func (d *communicationMetricsDecorator) SendNotification(ctx context.Context, req portsclient.SendNotificationRequestDTO, xApplication, correlationID string) error {
	start := time.Now()

	err := d.inner.SendNotification(ctx, req, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/notifications/send", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/notifications/send", errorType)
	}

	return err
}

// SendMessage implementa SendMessage com métricas
func (d *communicationMetricsDecorator) SendMessage(ctx context.Context, req communicationDto.SendMessageRequestDTO, xApplication, correlationID string) (communicationDto.SendMessageResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.SendMessage(ctx, req, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/messages/send", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.Message
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/messages/send", errorType)
	}

	return response, err
}
