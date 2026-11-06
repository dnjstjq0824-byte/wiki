package wiki

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/db"
)

// pgTestService connects to the test database (migrated) or skips when
// TREEWIKI_TEST_DATABASE_URL is unset. CI provides a postgres service.
func pgTestService(t *testing.T) *WikiService {
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
	return NewWikiService(NewPostgresPageStore(pgDB.Pool))
}

func pgTitle(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, strings.ReplaceAll(uuid.NewString()[:8], "-", "x"))
}

func TestPostgresWikiLifecycle(t *testing.T) {
	svc := pgTestService(t)
	title := pgTitle("PgLife")

	created, err := svc.CreatePage(NSMain, title, "hello [[PgTarget]] [[Category:PgCat]]", "init", "pg-author", "REGISTERED")
	if err != nil {
		t.Fatalf("CreatePage: %v", err)
	}
	if created.Content != "hello [[PgTarget]] [[Category:PgCat]]" {
		t.Fatalf("content mismatch: %q", created.Content)
	}
	if created.AuthorID != "pg-author" {
		t.Fatalf("author mismatch: %q", created.AuthorID)
	}
	if created.CurrentRevID == "" {
		t.Fatal("current revision must be set")
	}

	// Duplicate create must fail like memory.
	if _, err := svc.CreatePage(NSMain, title, "x", "dup", "u", "REGISTERED"); err != ErrPageExists {
		t.Fatalf("expected ErrPageExists, got %v", err)
	}

	// Update with optimistic concurrency.
	updated, err := svc.UpdatePage(NSMain, title, "hello v2", "second", "pg-author2", "REGISTERED", created.CurrentRevID, false)
	if err != nil {
		t.Fatalf("UpdatePage: %v", err)
	}
	if updated.Content != "hello v2" || updated.AuthorID != "pg-author2" {
		t.Fatalf("update mismatch: %+v", updated)
	}
	if _, err := svc.UpdatePage(NSMain, title, "stale", "s", "u", "REGISTERED", created.CurrentRevID, false); err != ErrRevisionConflict {
		t.Fatalf("expected conflict, got %v", err)
	}

	// History oldest-first with parent linkage.
	revs := svc.GetRevisions(NSMain, title)
	if len(revs) != 2 {
		t.Fatalf("expected 2 revisions, got %d", len(revs))
	}
	if revs[0].ParentRevisionID != "" || revs[1].ParentRevisionID != revs[0].ID {
		t.Fatalf("parent linkage broken: %+v", revs)
	}

	// Backlinks and categories resolve through SQL relations.
	svc.CreatePage(NSMain, title+"_Linker", "see [["+title+"]]", "init", "u", "REGISTERED")
	found := false
	for _, p := range svc.GetBacklinks(NSMain, title) {
		if p.Title == title+"_Linker" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected linker backlink")
	}

	// Revert restores v1 content as a new revision.
	reverted, err := svc.RevertToRevision(NSMain, title, revs[0].ID, "pg-admin", "SYSOP")
	if err != nil {
		t.Fatalf("RevertToRevision: %v", err)
	}
	if reverted.Content != revs[0].Content {
		t.Fatalf("revert content mismatch: %q vs %q", reverted.Content, revs[0].Content)
	}

	// Move + delete + undelete.
	moved := title + "_Moved"
	if err := svc.MovePage(NSMain, title, NSMain, moved, "u", "REGISTERED", false); err != nil {
		t.Fatalf("MovePage: %v", err)
	}
	if _, err := svc.GetPage(NSMain, moved); err != nil {
		t.Fatalf("moved page missing: %v", err)
	}
	if err := svc.DeletePage(NSMain, moved, "REGISTERED"); err != nil {
		t.Fatalf("DeletePage: %v", err)
	}
	if _, err := svc.GetPage(NSMain, moved); err != ErrPageDeleted {
		t.Fatalf("expected deleted, got %v", err)
	}
	if err := svc.UndeletePage(NSMain, moved, "SYSOP"); err != nil {
		t.Fatalf("UndeletePage: %v", err)
	}
	if _, err := svc.GetPage(NSMain, moved); err != nil {
		t.Fatalf("undelete failed: %v", err)
	}
	// The revert restored v1 categories; the move preserved them by page UUID.
	catFound := false
	for _, p := range svc.GetCategoryMembers("PgCat") {
		if p.Title == moved {
			catFound = true
		}
	}
	if !catFound {
		t.Fatal("expected moved page in PgCat members")
	}
}

func TestPostgresWikiCountsRandomAndProtection(t *testing.T) {
	svc := pgTestService(t)
	before := svc.CountAll()
	a := pgTitle("PgCountA")
	b := pgTitle("PgCountB")
	if _, err := svc.CreatePage(NSMain, a, "a", "init", "u", "REGISTERED"); err != nil {
		t.Fatalf("create A: %v", err)
	}
	if _, err := svc.CreatePage(NSMain, b, "b", "init", "u", "REGISTERED"); err != nil {
		t.Fatalf("create B: %v", err)
	}
	if got := svc.CountAll(); got != before+2 {
		t.Fatalf("CountAll delta: before=%d got=%d", before, got)
	}
	rnd, err := svc.RandomPage(nil)
	if err != nil || rnd.Title == "" {
		t.Fatalf("RandomPage: %v %+v", err, rnd)
	}
	// Invalid protection level must surface cleanly.
	if err := svc.SetProtection(NSMain, a, "nope-not-real", "SYSOP"); err == nil {
		t.Fatal("expected protection validation error")
	}
	// Timestamps are real.
	got, _ := svc.GetPage(NSMain, a)
	if time.Since(got.CreatedAt) > 10*time.Minute {
		t.Fatalf("suspicious created_at: %v", got.CreatedAt)
	}
}
