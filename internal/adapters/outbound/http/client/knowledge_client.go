package client

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

type knowledgeHTTP struct {
	httpClient *resty.Client
	baseURL    string
	logger     *zap.Logger
}

func NewKnowledgeClient(cfg *config.Config, logger *zap.Logger) domainclient.KnowledgeClient {
	timeout := cfg.Services.Knowledge.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	httpClient := resty.New()
	httpClient.SetTimeout(timeout)
	httpClient.SetRetryCount(0)
	return &knowledgeHTTP{
		httpClient: httpClient,
		baseURL:    cfg.Services.Knowledge.BaseURL,
		logger:     logger,
	}
}

func (c *knowledgeHTTP) Ask(ctx context.Context, companyID, bearerToken, correlationID string, body appdto.KnowledgeAskRequest) (appdto.KnowledgeAskResponse, error) {
	if c.baseURL == "" {
		return appdto.KnowledgeAskResponse{}, MapHTTPError(503, []byte(`{"message":"knowledge service indisponível"}`), "knowledge service")
	}
	var out appdto.KnowledgeAskResponse
	req := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&out)
	if bearerToken != "" {
		req.SetHeader("Authorization", bearerToken)
	}
	resp, err := req.Post(c.baseURL + "/api/v1/knowledge/ask")
	if err != nil {
		return appdto.KnowledgeAskResponse{}, MapNetworkError(err, "knowledge service")
	}
	if resp.StatusCode() != 200 {
		return appdto.KnowledgeAskResponse{}, MapHTTPError(resp.StatusCode(), resp.Body(), "knowledge service")
	}
	if out.Sources == nil {
		out.Sources = []appdto.KnowledgeAskSource{}
	}
	return out, nil
}
