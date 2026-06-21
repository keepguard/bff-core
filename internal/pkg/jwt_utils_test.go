package pkg

import (
	"testing"
)

func TestExtractCodeUserFromToken_Success(t *testing.T) {
	// Token válido com codeUser no payload (formato: header.payload.signature)
	// payload base64: {"codeUser":"test123","sub":"user1"}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2RlVXNlciI6InRlc3QxMjMiLCJzdWIiOiJ1c2VyMSJ9.signature"

	codeUser, err := ExtractCodeUserFromToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if codeUser != "test123" {
		t.Fatalf("expected test123, got %s", codeUser)
	}
}

func TestExtractCodeUserFromToken_WithBearerPrefix(t *testing.T) {
	token := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2RlVXNlciI6InRlc3QxMjMifQ.sig"

	codeUser, err := ExtractCodeUserFromToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if codeUser != "test123" {
		t.Fatalf("expected test123, got %s", codeUser)
	}
}

func TestExtractCodeUserFromToken_InvalidFormat(t *testing.T) {
	token := "invalid"

	_, err := ExtractCodeUserFromToken(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractCodeUserFromToken_InvalidBase64(t *testing.T) {
	token := "header.invalid_base64!@#.signature"

	_, err := ExtractCodeUserFromToken(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractCodeUserFromToken_InvalidJSON(t *testing.T) {
	token := "header.aW52YWxpZGpzb24.signature"

	_, err := ExtractCodeUserFromToken(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractCodeUserFromToken_NoCodeUser(t *testing.T) {
	// payload: {"other":"value"}
	token := "header.eyJvdGhlciI6InZhbHVlIn0.signature"

	_, err := ExtractCodeUserFromToken(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractCodeUserFromToken_FallbackToSub(t *testing.T) {
	// payload: {"sub":"user123"}
	token := "header.eyJzdWIiOiJ1c2VyMTIzIn0.signature"

	codeUser, err := ExtractCodeUserFromToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if codeUser != "user123" {
		t.Fatalf("expected user123, got %s", codeUser)
	}
}

func TestExtractAllClaims_Success(t *testing.T) {
	// payload: {"codeUser":"c1","email":"j@e.com"}
	token := "header.eyJjb2RlVXNlciI6ImMxIiwiZW1haWwiOiJqQGUuY29tIn0.signature"

	claims, err := ExtractAllClaims(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.CodeUser != "c1" || claims.Email != "j@e.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestExtractAllClaims_InvalidFormat(t *testing.T) {
	token := "invalid"

	_, err := ExtractAllClaims(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
