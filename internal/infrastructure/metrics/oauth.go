package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
)

// OAuth token results (closed label set for bff_oauth_token_total{result}).
const (
	OAuthResultCacheHit        = "cache_hit"
	OAuthResultFetched         = "fetched"
	OAuthResultCacheBypass     = "cache_bypass"
	OAuthResultNoClient        = "no_client"
	OAuthResultCompanyMismatch = "company_mismatch"
	OAuthResultError           = "error"
)

var oauthTokenTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bff_oauth_token_total",
		Help: "Per-company service token lookups by result",
	},
	[]string{"result"},
)

var oauthAllowedResults = map[string]struct{}{
	OAuthResultCacheHit: {}, OAuthResultFetched: {}, OAuthResultCacheBypass: {},
	OAuthResultNoClient: {}, OAuthResultCompanyMismatch: {}, OAuthResultError: {},
}

// RecordOAuthToken counts a token lookup. Unknown results collapse into "error" (label set stays closed).
func RecordOAuthToken(result string) {
	if _, ok := oauthAllowedResults[result]; !ok {
		result = OAuthResultError
	}
	oauthTokenTotal.WithLabelValues(result).Inc()
}

// OAuthTokenCount reads the counter (tests).
func OAuthTokenCount(result string) float64 {
	var out dto.Metric
	if err := oauthTokenTotal.WithLabelValues(result).Write(&out); err != nil || out.Counter == nil {
		return 0
	}
	return out.Counter.GetValue()
}
