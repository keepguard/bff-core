package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	client "github.com/keepguard/bff-core/internal/application/port"
	auditport "github.com/keepguard/bff-core/internal/domain/ports/audit"
	"github.com/keepguard/bff-core/internal/pkg"
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
			denied := status == http.StatusForbidden || status == http.StatusUnauthorized
			if !privilegedAuditRead(path) && domainCoveredByMS(path) && !denied {
				return err
			}
			outcome := "SUCCESS"
			if status >= 400 {
				outcome = "FAILURE"
			}
			if denied {
				outcome = "DENIED"
			}
			codeUser, tenantID, companyID, deviceID := auditIdentity(c)
			event := auditport.Event{
				EventID:       newAuditUUID(),
				OccurredAt:    time.Now().UTC().Format(time.RFC3339),
				SchemaVersion: 1,
				SourceService: sourceService,
				CorrelationID: GetCorrelationID(c),
				RequestID:     c.Response().Header().Get(echo.HeaderXRequestID),
				TenantID:      tenantID,
				CompanyID:     companyID,
				Actor: auditport.Actor{
					Type:     actorType(codeUser),
					CodeUser: codeUser,
					ClientIP: c.RealIP(),
					DeviceID: deviceID,
				},
				Action:   mapAuditAction(c.Request().Method, path),
				Resource: auditport.Resource{Type: "HTTP", ID: path},
				Outcome:  outcome,
				Metadata: map[string]any{"method": c.Request().Method, "status": status},
			}
			if strings.Contains(path, "/core/audits/") || strings.HasPrefix(path, "/api/v1/audits/") {
				event.Action = "AUDIT_READ"
				event.Resource = auditport.Resource{Type: "AUDIT_EVENT", ID: strings.TrimSpace(c.Param("eventId"))}
			}
			if strings.Contains(path, "/oauth/clients/") && c.Request().Method == http.MethodGet {
				event.Action = "OAUTH_CLIENT_READ"
				event.Resource = auditport.Resource{Type: "OAUTH_CLIENT", ID: strings.TrimSpace(c.Param("id"))}
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
	if method == http.MethodOptions || method == http.MethodHead {
		return true
	}
	if method == http.MethodGet {
		return !privilegedAuditRead(path)
	}
	return false
}

func privilegedAuditRead(path string) bool {
	// Apenas leitura de detalhe específico (/core/audits/:eventId ou /api/v1/audits/:eventId) é privilegiada.
	// Listagens gerais (/core/audits) são ignoradas para evitar auto-auditoria recursiva com o auto-refresh do frontend.
	if strings.Contains(path, "/core/audits/") || strings.HasPrefix(path, "/api/v1/audits/") {
		return true
	}
	if strings.Contains(path, "/core/oauth/clients/") && !strings.Contains(path, "/service-roles") {
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
	case strings.Contains(path, "/core/collector"):
		return true
	case strings.Contains(path, "/core/oauth"):
		return true
	case strings.Contains(path, "/core/guardian"):
		return true
	case strings.Contains(path, "/core/knowledge/ask"):
		return true
	case strings.Contains(path, "/core/llm"):
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
	case strings.Contains(path, "/core/audits/") || strings.HasPrefix(path, "/api/v1/audits/"):
		return "AUDIT_READ"
	default:
		return method + "_" + strings.Trim(path, "/")
	}
}

func auditIdentity(c echo.Context) (codeUser, tenantID, companyID, deviceID string) {
	claims := GetClaimsFromContext(c)
	if claims == nil {
		authHeader := c.Request().Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			claims, _ = pkg.ExtractAllClaims(authHeader)
		}
	}
	if claims != nil {
		codeUser = firstNonBlank(claims.CodeUser, claims.Sub, claims.UserID)
		tenantID = strings.TrimSpace(claims.TenantId)
		deviceID = strings.TrimSpace(claims.DeviceID)
	}
	if codeUser == "" {
		codeUser = strings.TrimSpace(GetUserIDFromContext(c))
	}
	if tenantID == "" {
		tenantID = ResolveTenantId(c, claims)
	}
	companyID = client.CompanyIDFromContext(c.Request().Context())
	if companyID == "" {
		companyID = strings.TrimSpace(c.Request().Header.Get("X-Company-Id"))
	}
	if deviceID == "" {
		deviceID = strings.TrimSpace(c.Request().Header.Get("X-Device-Id"))
	}
	return codeUser, tenantID, companyID, deviceID
}

func firstNonBlank(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
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
