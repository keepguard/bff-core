package config

import (
	"os"
	"testing"
)

// Default = previous behavior: W5 and W11 flags off in application.yml and in the defaults.
func TestLoad_SecurityFlagsOffByDefault(t *testing.T) {
	wd, _ := os.Getwd()
	if err := os.Chdir("../../.."); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	t.Setenv("BFF_CORE_ENV", "test-none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.OAuth.PerCompany.Enabled {
		t.Fatal("oauth.per_company.enabled must default to false")
	}
	if cfg.Authz.CompanyScopedCalls.Enabled {
		t.Fatal("authz.company_scoped_calls.enabled must default to false")
	}
	// saga_hardening is ON in application.yml since 16/11 (QA in production); W5/W11 above stay off.
	if !cfg.Registration.SagaHardening.Enabled {
		t.Fatal("registration.saga_hardening.enabled must be true in application.yml")
	}
}

func TestParseCompanyClientIDs(t *testing.T) {
	got := ParseCompanyClientIDs(" A-1 = bff-core-a ,b=bff-core-b,,bad, =x,y= ")
	if len(got) != 2 || got["a-1"] != "bff-core-a" || got["b"] != "bff-core-b" {
		t.Fatalf("unexpected %v", got)
	}
}
