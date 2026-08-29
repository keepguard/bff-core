package pkg

import "testing"

func TestNormalizeRole(t *testing.T) {
	if got := NormalizeRole("ROLE_ADMIN"); got != "ADMIN" {
		t.Fatalf("expected ADMIN, got %s", got)
	}
	if got := NormalizeRole(" system "); got != "SYSTEM" {
		t.Fatalf("expected SYSTEM, got %s", got)
	}
}

func TestHasAnyRole(t *testing.T) {
	if !HasAnyRole([]string{"ROLE_USER", "ADMIN"}, "SYSTEM", "ADMIN") {
		t.Fatal("expected ADMIN to match")
	}
	if HasAnyRole([]string{"USER"}, "ADMIN", "SYSTEM") {
		t.Fatal("USER must not match admin roles")
	}
	if !HasAnyRole([]string{"ROLE_SYSTEM"}, "ADMIN", "SYSTEM") {
		t.Fatal("expected SYSTEM to match")
	}
}

func TestHasAuthority(t *testing.T) {
	if !HasAuthority([]string{"user:block", "audit:read"}, "audit:read") {
		t.Fatal("expected audit:read")
	}
	if HasAuthority([]string{"user:block"}, "audit:read") {
		t.Fatal("must not match missing authority")
	}
}
