package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	auditport "github.com/keepguard/bff-core/internal/domain/ports/audit"
	"github.com/labstack/echo/v4"
)

func AuditMiddleware(publisher auditport.EventPublisher, sourceService string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if publisher == nil {
				return err
			}
			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}
			if shouldSkipAudit(c.Request().Method, path) {
				return err
			}
			status := c.Response().Status
			outcome := "SUCCESS"
			if status >= 400 {
				outcome = "FAILURE"
			}
			if status == http.StatusForbidden || status == http.StatusUnauthorized {
				outcome = "DENIED"
			}
			if outcome == "SUCCESS" && domainCoveredByMS(path) {
				return err
			}
			event := auditport.Event{
				EventID:       newAuditUUID(),
				OccurredAt:    time.Now().UTC().Format(time.RFC3339),
				SchemaVersion: 1,
				SourceService: sourceService,
				CorrelationID: GetCorrelationID(c),
				RequestID:     c.Response().Header().Get(echo.HeaderXRequestID),
				TenantID:      c.Request().Header.Get("X-Tenant-Id"),
				CompanyID:     c.Request().Header.Get("X-Company-Id"),
				Actor: auditport.Actor{
					Type:     actorType(c.Request().Header.Get("X-User-ID")),
					CodeUser: c.Request().Header.Get("X-User-ID"),
					ClientIP: c.RealIP(),
				},
				Action:   mapAuditAction(c.Request().Method, path),
				Resource: auditport.Resource{Type: "HTTP", ID: path},
				Outcome:  outcome,
				Metadata: map[string]any{"method": c.Request().Method, "status": status},
			}
			publisher.Publish(c.Request().Context(), event)
			return err
		}
	}
}

func shouldSkipAudit(method, path string) bool {
	if path == "/health" || strings.HasPrefix(path, "/swagger") || path == "/metrics" {
		return true
	}
	if method == http.MethodGet || method == http.MethodOptions || method == http.MethodHead {
		return true
	}
	return false
}

func domainCoveredByMS(path string) bool {
	switch {
	case strings.Contains(path, "/register/init"):
		return true
	case strings.Contains(path, "/register/confirm"):
		return true
	case strings.Contains(path, "/register/resend"):
		return true
	case strings.Contains(path, "/user-consents/accept"):
		return true
	default:
		return false
	}
}

func mapAuditAction(method, path string) string {
	switch {
	case strings.Contains(path, "/register/init"):
		return "REGISTER_INIT"
	case strings.Contains(path, "/register/confirm"):
		return "REGISTER_CONFIRM"
	case strings.Contains(path, "/register/resend"):
		return "REGISTER_RESEND"
	case strings.Contains(path, "/user-consents/accept"):
		return "ACCEPT_CONSENTS"
	default:
		return method + "_" + strings.Trim(path, "/")
	}
}

func actorType(codeUser string) string {
	if strings.TrimSpace(codeUser) == "" {
		return "ANONYMOUS"
	}
	return "USER"
}

func newAuditUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b)
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}
