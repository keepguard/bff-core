package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	auditport "github.com/keepguard/bff-core/internal/domain/ports/audit"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type recPub struct {
	events []auditport.Event
}

func (r *recPub) Publish(_ context.Context, event auditport.Event) {
	r.events = append(r.events, event)
}

func (r *recPub) Close() error { return nil }

func TestMapAuditActionRegister(t *testing.T) {
	require.Equal(t, "REGISTER_INIT", mapAuditAction(http.MethodPost, "/api/v1/register/init"))
	require.Equal(t, "REGISTER_CONFIRM", mapAuditAction(http.MethodPost, "/api/v1/register/confirm"))
	require.Equal(t, "ACCEPT_CONSENTS", mapAuditAction(http.MethodPost, "/api/v1/user-consents/accept-batch"))
}

func TestAuditMiddlewareEmitsRegisterWithCorrelationID(t *testing.T) {
	rec := &recPub{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", nil)
	req.Header.Set("X-Correlation-ID", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	c.SetPath("/api/v1/register/init")

	mw := AuditMiddleware(rec, "bff-core")
	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	require.NoError(t, err)
	require.Len(t, rec.events, 1)
	require.Equal(t, "REGISTER_INIT", rec.events[0].Action)
	require.Equal(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", rec.events[0].CorrelationID)
	require.Equal(t, "SUCCESS", rec.events[0].Outcome)
	require.Equal(t, "bff-core", rec.events[0].SourceService)
}
