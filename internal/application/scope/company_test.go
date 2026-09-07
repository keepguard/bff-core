package scope

import (
	"context"
	"errors"
	"net/http"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
)

type stubCompany struct {
	id  string
	err error
}

func (s stubCompany) GetByTenantId(_ context.Context, _, _ string) (appdto.MSCompanyResponseDTO, error) {
	if s.err != nil {
		return appdto.MSCompanyResponseDTO{}, s.err
	}
	return appdto.MSCompanyResponseDTO{ID: s.id}, nil
}

func TestResolveCompany_FromContext(t *testing.T) {
	id, err := ResolveCompany(context.Background(), nil, "company-ctx", "", "corr")
	if err != nil {
		t.Fatal(err)
	}
	if id != "company-ctx" {
		t.Fatalf("got %q", id)
	}
}

func TestResolveCompany_MissingTenant(t *testing.T) {
	_, err := ResolveCompany(context.Background(), stubCompany{id: "c1"}, "", "  ", "corr")
	appErr, ok := err.(*pkg.AppError)
	if !ok || appErr.Code != "MISSING_TENANT" || appErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestResolveCompany_NilCompanies(t *testing.T) {
	_, err := ResolveCompany(context.Background(), nil, "", "tenant-1", "corr")
	appErr, ok := err.(*pkg.AppError)
	if !ok || appErr.Code != "SERVICE_UNAVAILABLE" || appErr.Message != "Gestão de agents indisponível" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestResolveCompany_NotFound(t *testing.T) {
	_, err := ResolveCompany(context.Background(), stubCompany{}, "", "tenant-1", "corr")
	appErr, ok := err.(*pkg.AppError)
	if !ok || appErr.Code != "COMPANY_NOT_FOUND" || appErr.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestResolveCompany_GetByTenantIdError(t *testing.T) {
	want := errors.New("upstream")
	_, err := ResolveCompany(context.Background(), stubCompany{err: want}, "", "tenant-1", "corr")
	if err != want {
		t.Fatalf("expected upstream error, got %#v", err)
	}
}

func TestResolveCompany_Success(t *testing.T) {
	id, err := ResolveCompany(context.Background(), stubCompany{id: "company-1"}, "", "tenant-1", "corr")
	if err != nil {
		t.Fatal(err)
	}
	if id != "company-1" {
		t.Fatalf("got %q", id)
	}
}

func TestUnavailable(t *testing.T) {
	err := Unavailable("Serviço de conhecimento indisponível")
	if err.Code != "SERVICE_UNAVAILABLE" || err.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected: %#v", err)
	}
}
