package talk

import (
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/db"
)

func pgTestService(t *testing.T) *TalkService {
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
	return NewTalkService(NewPostgresTalkStore(pgDB.Pool))
}

func TestPostgresTalkWorkflow(t *testing.T) {
	svc := pgTestService(t)
	pageKey := "0:PgTalk_" + uuid.NewString()[:8]

	thread, err := svc.CreateThread(pageKey, "pg thread", "pg-user")
	if err != nil {
		t.Fatalf("CreateThread: %v", err)
	}
	if thread.AuthorID != "pg-user" {
		t.Fatalf("author mismatch: %q", thread.AuthorID)
	}
	if n := len(svc.ListThreads(pageKey)); n != 1 {
		t.Fatalf("expected 1 thread, got %d", n)
	}

	c, err := svc.AddComment(thread.ID, "", "pg-replier", "hello @pguser")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}
	if len(c.Mentions) != 1 || c.Mentions[0] != "pguser" {
		t.Fatalf("mentions not persisted: %+v", c.Mentions)
	}
	if n := len(svc.GetComments(thread.ID)); n != 1 {
		t.Fatalf("expected 1 comment, got %d", n)
	}

	rfc, err := svc.CreateRFC(thread.ID, "pg question?", 2)
	if err != nil {
		t.Fatalf("CreateRFC: %v", err)
	}
	if err := svc.VoteRFC(rfc.ID, "pg-voter", "support"); err != nil {
		t.Fatalf("VoteRFC: %v", err)
	}
	if err := svc.VoteRFC(rfc.ID, "pg-voter", "oppose"); err != nil {
		t.Fatalf("VoteRFC replace: %v", err)
	}
	fetched, err := svc.GetRFCByThread(thread.ID)
	if err != nil {
		t.Fatalf("GetRFCByThread: %v", err)
	}
	if len(fetched.Votes) != 1 || fetched.Votes[0].Choice != "OPPOSE" {
		t.Fatalf("vote replace broken: %+v", fetched.Votes)
	}
	if err := svc.CloseRFC(rfc.ID); err != nil {
		t.Fatalf("CloseRFC: %v", err)
	}
	if err := svc.VoteRFC(rfc.ID, "late-voter", "SUPPORT"); err == nil {
		t.Fatal("expected closed-poll rejection")
	}

	if err := svc.PinThread(thread.ID, true); err != nil {
		t.Fatalf("PinThread: %v", err)
	}
	if err := svc.LockThread(thread.ID, true); err != nil {
		t.Fatalf("LockThread: %v", err)
	}
	if err := svc.ArchiveThread(thread.ID); err != nil {
		t.Fatalf("ArchiveThread: %v", err)
	}
	if n := len(svc.ListThreads(pageKey)); n != 0 {
		t.Fatalf("archived thread must leave live listing, got %d", n)
	}

	if _, err := svc.CreateArbitration(thread.ID, "pg dispute"); err != nil {
		t.Fatalf("CreateArbitration: %v", err)
	}
	if n := len(svc.ListArbitrations()); n < 1 {
		t.Fatalf("expected arbitrations, got %d", n)
	}
}
