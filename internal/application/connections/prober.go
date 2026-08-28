package connections

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ServiceStatus é o resultado de um probe.
type ServiceStatus struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Group         string `json:"group"`
	Endpoint      string `json:"endpoint"`
	Status        string `json:"status"`
	LatencyMs     int64  `json:"latencyMs"`
	HTTPStatus    int    `json:"httpStatus,omitempty"`
	TreatAuthAsUp bool   `json:"-"`
}

// ProbeAll dispara os healths em paralelo com timeout curto.
func ProbeAll(ctx context.Context, targets []Target, timeout time.Duration) []ServiceStatus {
	if timeout <= 0 {
		timeout = 500 * time.Millisecond
	}
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	results := make([]ServiceStatus, len(targets))
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Add(1)
		go func(idx int, t Target) {
			defer wg.Done()
			results[idx] = probeOne(ctx, client, t)
		}(i, target)
	}
	wg.Wait()
	return results
}

func probeOne(ctx context.Context, client *http.Client, target Target) ServiceStatus {
	item := ServiceStatus{
		ID:            target.ID,
		Name:          target.Name,
		Description:   target.Description,
		Group:         target.Group,
		Endpoint:      target.Endpoint,
		Status:        "unhealthy",
		TreatAuthAsUp: target.TreatAuthAsUp,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
	if err != nil {
		return item
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "keepguard-bff-core-connections-health")

	started := time.Now()
	resp, err := client.Do(req)
	item.LatencyMs = time.Since(started).Milliseconds()
	if err != nil {
		return item
	}
	defer resp.Body.Close()
	item.HTTPStatus = resp.StatusCode
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	ok := resp.StatusCode >= 200 && resp.StatusCode < 400
	if target.TreatAuthAsUp && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		ok = true
	}
	if ok && strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var payload struct {
			Status string `json:"status"`
		}
		if json.Unmarshal(body, &payload) == nil && strings.EqualFold(payload.Status, "DOWN") {
			ok = false
		}
	}
	if ok {
		item.Status = "healthy"
	}
	return item
}
