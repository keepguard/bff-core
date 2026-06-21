package pkg

import (
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := NewAppError("CODE", "message", http.StatusBadRequest)
	if err.Error() != "message" {
		t.Fatalf("expected 'message', got %s", err.Error())
	}
}

func TestAppError_WithTraceID(t *testing.T) {
	err := NewAppError("CODE", "message", http.StatusBadRequest)
	err = err.WithTraceID("trace123")
	if err.TraceID != "trace123" {
		t.Fatalf("expected trace123, got %s", err.TraceID)
	}
}

func TestAppError_ToResponse(t *testing.T) {
	err := NewAppError("CODE", "message", http.StatusBadRequest, ErrorDetail{Field: "field1", Message: "detail"})
	err = err.WithTraceID("trace123")
	resp := err.ToResponse()
	if resp.Error != "CODE" || resp.Message != "message" || resp.TraceID != "trace123" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if len(resp.Details) != 1 || resp.Details[0].Field != "field1" {
		t.Fatalf("unexpected details: %+v", resp.Details)
	}
}

func TestNewAppError_WithDetails(t *testing.T) {
	err := NewAppError("CODE", "message", http.StatusBadRequest,
		ErrorDetail{Field: "f1", Code: "c1", Message: "m1"},
		ErrorDetail{Field: "f2", Code: "c2", Message: "m2"})
	if len(err.Details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(err.Details))
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		code string
	}{
		{"InvalidRequest", ErrInvalidRequest, "INVALID_REQUEST"},
		{"Unauthorized", ErrUnauthorized, "UNAUTHORIZED"},
		{"Forbidden", ErrForbidden, "FORBIDDEN"},
		{"NotFound", ErrNotFound, "NOT_FOUND"},
		{"Conflict", ErrConflict, "CONFLICT"},
		{"InternalServer", ErrInternalServer, "INTERNAL_ERROR"},
		{"ServiceUnavailable", ErrServiceUnavailable, "SERVICE_UNAVAILABLE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Fatalf("expected code %s, got %s", tt.code, tt.err.Code)
			}
		})
	}
}
