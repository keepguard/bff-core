package collector

import (
	"context"
	"errors"
	"net/http"
	"strings"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/application/scope"
	"github.com/keepguard/bff-core/internal/pkg"
)

var errKnowledgeServiceTokenMissing = errors.New("token de serviço knowledge indisponível")

func (s *service) GetExecutionPayloads(ctx context.Context, query GetExecutionPayloadsQuery) ([]appdto.ExecutionPayloadItemDTO, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return nil, err
	}
	if s.knowledge == nil {
		return nil, scope.Unavailable("Serviço de conhecimento indisponível")
	}
	executionID := strings.TrimSpace(query.ExecutionID)
	if executionID == "" {
		return nil, pkg.NewAppError("BAD_REQUEST", "executionId é obrigatório", http.StatusBadRequest)
	}
	execution, err := s.collector.GetExecution(ctx, companyID, executionID, query.CorrelationID)
	if err != nil {
		return nil, err
	}
	items, err := s.loadExecutionPayloads(ctx, companyID, query.CorrelationID, execution)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []appdto.ExecutionPayloadItemDTO{}
	}
	return items, nil
}

func (s *service) loadExecutionPayloads(
	ctx context.Context, companyID, correlationID string, execution appdto.CollectorExecutionRaw,
) ([]appdto.ExecutionPayloadItemDTO, error) {
	bearer, tokErr := knowledgeServiceBearer(ctx, s.tokens, companyID)
	if tokErr != nil {
		return nil, tokErr
	}
	refs := parsePayloadRefs(execution.Metadata)
	if len(refs) > 0 {
		items := make([]appdto.ExecutionPayloadItemDTO, 0, len(refs))
		for _, ref := range refs {
			item, err := s.loadPayloadRef(ctx, companyID, bearer, correlationID, ref)
			if err != nil {
				if isNotFound(err) {
					continue
				}
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	}
	if strings.TrimSpace(execution.AgentID) == "" || strings.TrimSpace(execution.StartedAt) == "" {
		return []appdto.ExecutionPayloadItemDTO{}, nil
	}
	results, err := s.knowledge.GetCollectionResults(
		ctx, companyID, bearer, correlationID, execution.AgentID, execution.StartedAt, 60,
	)
	if err != nil {
		return nil, err
	}
	items := make([]appdto.ExecutionPayloadItemDTO, 0, len(results.Snapshots)+len(results.Documents))
	for _, snapshot := range results.Snapshots {
		items = append(items, snapshotToPayloadItem(snapshot))
	}
	for _, document := range results.Documents {
		items = append(items, documentToPayloadItem(document))
	}
	return items, nil
}

type payloadRef struct {
	Kind string
	ID   string
}

func parsePayloadRefs(metadata map[string]any) []payloadRef {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata["payload_refs"]
	if !ok || raw == nil {
		return nil
	}
	var rows []any
	switch typed := raw.(type) {
	case []any:
		rows = typed
	case []map[string]any:
		for _, item := range typed {
			rows = append(rows, item)
		}
	default:
		return nil
	}
	refs := make([]payloadRef, 0, len(rows))
	for _, row := range rows {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		kind, _ := item["kind"].(string)
		id, _ := item["id"].(string)
		kind = strings.TrimSpace(strings.ToLower(kind))
		id = strings.TrimSpace(id)
		if (kind != "snapshot" && kind != "document") || id == "" {
			continue
		}
		refs = append(refs, payloadRef{Kind: kind, ID: id})
	}
	return refs
}

func (s *service) loadPayloadRef(
	ctx context.Context, companyID, bearer, correlationID string, ref payloadRef,
) (appdto.ExecutionPayloadItemDTO, error) {
	switch ref.Kind {
	case "snapshot":
		snapshot, err := s.knowledge.GetSnapshot(ctx, companyID, bearer, correlationID, ref.ID)
		if err != nil {
			return appdto.ExecutionPayloadItemDTO{}, err
		}
		return snapshotToPayloadItem(snapshot), nil
	case "document":
		document, err := s.knowledge.GetDocumentPreview(ctx, companyID, bearer, correlationID, ref.ID)
		if err != nil {
			return appdto.ExecutionPayloadItemDTO{}, err
		}
		return documentToPayloadItem(document), nil
	default:
		return appdto.ExecutionPayloadItemDTO{}, nil
	}
}

func snapshotToPayloadItem(snapshot appdto.KnowledgeSnapshotDTO) appdto.ExecutionPayloadItemDTO {
	return appdto.ExecutionPayloadItemDTO{
		Kind:        "snapshot",
		ID:          snapshot.ID,
		ContentType: "application/json",
		Payload:     snapshot.Payload,
		Metadata: map[string]any{
			"collectorType": snapshot.CollectorType,
			"entityHint":    snapshot.EntityHint,
			"collectedAt":   snapshot.CollectedAt,
			"schema":        snapshot.Schema,
			"sourceUrl":     snapshot.SourceURL,
		},
	}
}

func documentToPayloadItem(document appdto.KnowledgeDocumentPreviewDTO) appdto.ExecutionPayloadItemDTO {
	return appdto.ExecutionPayloadItemDTO{
		Kind:        "document",
		ID:          document.ID,
		ContentType: document.ContentType,
		FileName:    document.FileName,
		PreviewText: document.PreviewText,
		Metadata: map[string]any{
			"entityHint":       document.EntityHint,
			"dataSource":       document.DataSource,
			"sourceKey":        document.SourceKey,
			"collectedAt":      document.CollectedAt,
			"status":           document.Status,
			"previewAvailable": document.PreviewAvailable,
			"message":          document.Message,
		},
	}
}

func isNotFound(err error) bool {
	httpErr, ok := err.(*appdto.HTTPError)
	return ok && httpErr.Code == http.StatusNotFound
}

func knowledgeServiceBearer(ctx context.Context, tokens port.ServiceTokenClient, companyID string) (string, error) {
	if tokens == nil {
		return "", errKnowledgeServiceTokenMissing
	}
	return tokens.GetToken(ctx, companyID)
}
