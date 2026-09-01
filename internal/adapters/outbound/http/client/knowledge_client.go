package client

import (
	"context"
	"strconv"
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

func (c *knowledgeHTTP) authReq(ctx context.Context, companyID, bearerToken, correlationID string) *resty.Request {
	req := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Company-Id", companyID).
		SetHeader("X-Correlation-ID", correlationID)
	if bearerToken != "" {
		req.SetHeader("Authorization", bearerToken)
	}
	return req
}

func (c *knowledgeHTTP) GetSnapshot(ctx context.Context, companyID, bearerToken, correlationID, snapshotID string) (appdto.KnowledgeSnapshotDTO, error) {
	if c.baseURL == "" {
		return appdto.KnowledgeSnapshotDTO{}, MapHTTPError(503, []byte(`{"message":"knowledge service indisponível"}`), "knowledge service")
	}
	var out appdto.KnowledgeSnapshotDTO
	resp, err := c.authReq(ctx, companyID, bearerToken, correlationID).
		SetResult(&out).
		Get(c.baseURL + "/api/v1/knowledge/snapshots/" + snapshotID)
	if err != nil {
		return appdto.KnowledgeSnapshotDTO{}, MapNetworkError(err, "knowledge service")
	}
	if resp.StatusCode() != 200 {
		return appdto.KnowledgeSnapshotDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "knowledge service")
	}
	return out, nil
}

func (c *knowledgeHTTP) GetDocumentPreview(ctx context.Context, companyID, bearerToken, correlationID, documentID string) (appdto.KnowledgeDocumentPreviewDTO, error) {
	if c.baseURL == "" {
		return appdto.KnowledgeDocumentPreviewDTO{}, MapHTTPError(503, []byte(`{"message":"knowledge service indisponível"}`), "knowledge service")
	}
	var out appdto.KnowledgeDocumentPreviewDTO
	resp, err := c.authReq(ctx, companyID, bearerToken, correlationID).
		SetResult(&out).
		Get(c.baseURL + "/api/v1/knowledge/documents/" + documentID + "/preview")
	if err != nil {
		return appdto.KnowledgeDocumentPreviewDTO{}, MapNetworkError(err, "knowledge service")
	}
	if resp.StatusCode() != 200 {
		return appdto.KnowledgeDocumentPreviewDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "knowledge service")
	}
	return out, nil
}

func (c *knowledgeHTTP) GetCollectionResults(ctx context.Context, companyID, bearerToken, correlationID, agentID, collectedAt string, windowSeconds int) (appdto.KnowledgeCollectionResultsDTO, error) {
	if c.baseURL == "" {
		return appdto.KnowledgeCollectionResultsDTO{}, MapHTTPError(503, []byte(`{"message":"knowledge service indisponível"}`), "knowledge service")
	}
	if windowSeconds <= 0 {
		windowSeconds = 60
	}
	var out appdto.KnowledgeCollectionResultsDTO
	resp, err := c.authReq(ctx, companyID, bearerToken, correlationID).
		SetQueryParam("agentId", agentID).
		SetQueryParam("collectedAt", collectedAt).
		SetQueryParam("windowSeconds", strconv.Itoa(windowSeconds)).
		SetResult(&out).
		Get(c.baseURL + "/api/v1/knowledge/collection-results")
	if err != nil {
		return appdto.KnowledgeCollectionResultsDTO{}, MapNetworkError(err, "knowledge service")
	}
	if resp.StatusCode() != 200 {
		return appdto.KnowledgeCollectionResultsDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "knowledge service")
	}
	if out.Snapshots == nil {
		out.Snapshots = []appdto.KnowledgeSnapshotDTO{}
	}
	if out.Documents == nil {
		out.Documents = []appdto.KnowledgeDocumentPreviewDTO{}
	}
	return out, nil
}
