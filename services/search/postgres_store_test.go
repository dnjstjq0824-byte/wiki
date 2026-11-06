package search

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/db"
)

func pgTestBackends(t *testing.T) (*PostgresSearchBackend, *PostgresMediaStore, string) {
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
	tag := uuid.NewString()[:8]
	title := "PgSearchDoc_" + tag
	content := "the quick pgsearchable-" + tag + " wombat jumps over lexical fences"
	var pageID string
	if err := pgDB.Pool.QueryRow(ctx, `
		INSERT INTO pages (namespace_id, title, slug) VALUES (0, $1, $2)
		RETURNING id
	`, title, title).Scan(&pageID); err != nil {
		t.Fatalf("seed page: %v", err)
	}
	var revID string
	if err := pgDB.Pool.QueryRow(ctx, `
		INSERT INTO revisions (page_id, author_name, summary, byte_length, sha256_hash)
		VALUES ($1, 'pg-seed', 'seed', $2, '0')
		RETURNING id
	`, pageID, len(content)).Scan(&revID); err != nil {
		t.Fatalf("seed revision: %v", err)
	}
	if _, err := pgDB.Pool.Exec(ctx, `
		INSERT INTO revision_text (revision_id, raw_wikitext) VALUES ($1, $2)
	`, revID, content); err != nil {
		t.Fatalf("seed text: %v", err)
	}
	if _, err := pgDB.Pool.Exec(ctx, `
		UPDATE pages
		SET current_revision_id = $1,
		    search_vector = to_tsvector('simple', title || ' ' || $2)
		WHERE id = $3
	`, revID, content, pageID); err != nil {
		t.Fatalf("link revision: %v", err)
	}
	return NewPostgresSearchBackend(pgDB.Pool), NewPostgresMediaStore(pgDB.Pool), tag
}

func TestPostgresSearchFaceted(t *testing.T) {
	backend, _, tag := pgTestBackends(t)

	res := backend.SearchFaceted("pgsearchable-"+tag, -1, "", "", "")
	if len(res) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(res))
	}
	if res[0].Document.Title == "" || res[0].Snippet == "" {
		t.Fatalf("hit lacks title/snippet: %+v", res[0])
	}
	if got := backend.SearchFaceted("pgsearchable-"+tag, 4, "", "", ""); len(got) != 0 {
		t.Fatalf("namespace filter broken, got %d", len(got))
	}
	if got := backend.Search("pgsearchable-"+tag, 0); len(got) != 1 {
		t.Fatalf("Search wrapper broken, got %d", len(got))
	}
	sugg := backend.Autocomplete("PgSearchDoc_"+tag, 10)
	if len(sugg) != 1 {
		t.Fatalf("autocomplete broken: %+v", sugg)
	}
	if got := backend.SearchFaceted("", -1, "", "", ""); len(got) < 1 {
		t.Fatal("empty query must list documents")
	}
}

func TestPostgresMediaWorkflow(t *testing.T) {
	_, media, tag := pgTestBackends(t)
	name := "pg-asset-" + tag + ".png"

	a := &MediaAsset{
		Filename: name, OriginalName: name, MimeType: "image/png",
		FileSize: 12, S3Key: "uploads/" + name, SHA256: "abc123",
		LicenseType: "CC_BY_SA_4_0", UploaderID: "pg-uploader",
	}
	if err := media.Add(a); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if a.ID == "" || a.CreatedAt == "" {
		t.Fatalf("Add must fill id/created: %+v", a)
	}
	got, ok := media.Get("media-" + name)
	if !ok || got.Filename != name || got.UploaderID != "pg-uploader" {
		t.Fatalf("Get mismatch: %+v %v", got, ok)
	}
	if _, ok := media.Get(a.ID); !ok {
		t.Fatal("Get by UUID must work")
	}
	if n := len(media.List()); n < 1 {
		t.Fatalf("List must include asset, got %d", n)
	}
	if err := media.Add(&MediaAsset{Filename: "", SHA256: "x"}); err == nil {
		t.Fatal("expected filename validation error")
	}
}
