package watchlist

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/db"
)

func pgTestService(t *testing.T) (*Service, string, string) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userID := uuid.NewString()
	if _, err := pgDB.Pool.Exec(ctx, `
		INSERT INTO users (id, username) VALUES ($1, $2)
	`, userID, "pgwatch_"+uuid.NewString()[:8]); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	title := "PgWatch_" + uuid.NewString()[:8]
	if _, err := pgDB.Pool.Exec(ctx, `
		INSERT INTO pages (namespace_id, title, slug) VALUES (0, $1, $2)
	`, title, title); err != nil {
		t.Fatalf("seed page: %v", err)
	}
	return NewService(NewPostgresWatchlistStore(pgDB.Pool)), userID, "0:" + title
}

func TestPostgresWatchlistWorkflow(t *testing.T) {
	svc, userID, pageKey := pgTestService(t)

	if err := svc.Add(userID, pageKey, true); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !svc.IsWatching(userID, pageKey) {
		t.Fatal("IsWatching must be true after Add")
	}
	items := svc.List(userID)
	if len(items) != 1 || !items[0].NotifyEmail {
		t.Fatalf("List wrong: %+v", items)
	}
	if watchers := svc.Watchers(pageKey); len(watchers) != 1 || watchers[0] != userID {
		t.Fatalf("Watchers wrong: %+v", watchers)
	}
	// Idempotent re-add.
	if err := svc.Add(userID, pageKey, false); err != nil {
		t.Fatalf("re-Add: %v", err)
	}
	if err := svc.Remove(userID, pageKey); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if svc.IsWatching(userID, pageKey) {
		t.Fatal("IsWatching must be false after Remove")
	}
	// Unknown page and user surface cleanly.
	if err := svc.Add(userID, "0:Missing_"+uuid.NewString()[:8], false); err == nil {
		t.Fatal("expected error for missing page")
	}
	if err := svc.Add("not-a-uuid", pageKey, false); err == nil {
		t.Fatal("expected error for bad user id")
	}
}
