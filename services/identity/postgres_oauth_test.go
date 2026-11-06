package identity

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/db"
)

// pgOAuthService connects to the test database (migrated) or skips when
// TREEWIKI_TEST_DATABASE_URL is unset. CI provides a postgres service.
func pgOAuthService(t *testing.T) *IdentityService {
	t.Helper()
	url := os.Getenv("TREEWIKI_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TREEWIKI_TEST_DATABASE_URL not set; skipping postgres integration test")
	}
	pgDB, err := db.Connect(url)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(func() { pgDB.Close() })
	if err := pgDB.MigrateAll("../../database/migrations"); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return NewIdentityService(NewPostgresUserStore(pgDB.Pool), "test-secret")
}

func pgOAuthUsername(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, strings.ReplaceAll(uuid.NewString()[:8], "-", "x"))
}

func TestPostgresOAuthLinkRoundtrip(t *testing.T) {
	svc := pgOAuthService(t)
	name := pgOAuthUsername("OAuthLink")

	registered, err := svc.Register(RegisterRequest{
		Username: name,
		Email:    name + "@example.com",
		Password: "Str0ngPass!x",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := svc.store.GetByOAuthLink("google", "pg-oauth-1"); err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound for missing link, got %v", err)
	}

	if err := svc.store.CreateOAuthLink(registered.User.ID, "google", "pg-oauth-1"); err != nil {
		t.Fatalf("CreateOAuthLink: %v", err)
	}
	// Idempotent for the same owner.
	if err := svc.store.CreateOAuthLink(registered.User.ID, "google", "pg-oauth-1"); err != nil {
		t.Fatalf("re-link same owner: %v", err)
	}

	linked, err := svc.store.GetByOAuthLink("google", "pg-oauth-1")
	if err != nil {
		t.Fatalf("GetByOAuthLink: %v", err)
	}
	if linked.ID != registered.User.ID || linked.Username != name {
		t.Fatalf("link mismatch: %+v", linked)
	}

	// Same provider identity cannot be claimed by another account.
	other := pgOAuthUsername("OAuthOther")
	registeredOther, err := svc.Register(RegisterRequest{
		Username: other,
		Email:    other + "@example.com",
		Password: "Str0ngPass!x",
	})
	if err != nil {
		t.Fatalf("Register other: %v", err)
	}
	if err := svc.store.CreateOAuthLink(registeredOther.User.ID, "google", "pg-oauth-1"); err != ErrOAuthLinkTaken {
		t.Fatalf("expected ErrOAuthLinkTaken, got %v", err)
	}

	if err := svc.store.CreateOAuthLink("missing-user", "google", "pg-oauth-2"); err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound for missing user, got %v", err)
	}
}
