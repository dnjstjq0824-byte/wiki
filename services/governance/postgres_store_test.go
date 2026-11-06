package governance

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/db"
)

func pgTestService(t *testing.T) *GovernanceService {
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
	return NewGovernanceService(NewPostgresGovernanceStore(pgDB.Pool))
}

func TestPostgresGovernanceWorkflow(t *testing.T) {
	svc := pgTestService(t)
	tag := uuid.NewString()[:8]

	if err := svc.AddFilter(&AbuseFilter{Name: "pgfilter-" + tag, Pattern: "pgspam-" + tag, Action: "WARN"}); err != nil {
		t.Fatalf("AddFilter: %v", err)
	}
	if err := svc.AddFilter(&AbuseFilter{Name: "bad", Pattern: "([", Action: "TAG"}); err == nil {
		t.Fatal("expected regex compile error")
	}
	if err := svc.AddFilter(&AbuseFilter{Name: "bad2", Pattern: "x", Action: "NUKE"}); err == nil {
		t.Fatal("expected action validation error")
	}
	action, f := svc.Evaluate("contains pgspam-"+tag+" here", "SomePage", "pg-user", "192.0.2.9")
	if action != "WARN" || f == nil {
		t.Fatalf("Evaluate mismatch: %q %+v", action, f)
	}
	if action, _ := svc.Evaluate("clean text", "SomePage", "u", ""); action != "" {
		t.Fatalf("clean text must not match: %q", action)
	}

	expiry := time.Now().UTC().Add(time.Hour)
	if err := svc.AddBlock(&Block{TargetUserID: "pg-blocked-" + tag, BlockedBy: "pg-admin", Reason: "test", ExpiresAt: &expiry}); err != nil {
		t.Fatalf("AddBlock: %v", err)
	}
	if !svc.IsBlocked("pg-blocked-"+tag, "") {
		t.Fatal("IsBlocked must hit user block")
	}
	if err := svc.AddBlock(&Block{TargetIP: "192.0.2.44", BlockedBy: "pg-admin", Reason: "test"}); err != nil {
		t.Fatalf("AddBlock ip: %v", err)
	}
	if !svc.IsBlocked("", "192.0.2.44") {
		t.Fatal("IsBlocked must hit ip block")
	}
	past := time.Now().UTC().Add(-time.Hour)
	if err := svc.AddBlock(&Block{TargetUserID: "pg-expired-" + tag, BlockedBy: "a", Reason: "x", ExpiresAt: &past}); err != nil {
		t.Fatalf("AddBlock expired: %v", err)
	}
	if svc.IsBlocked("pg-expired-"+tag, "") {
		t.Fatal("expired block must not match")
	}

	if err := svc.AddOversight(&OversightLog{ActorID: "pg-oversight", Action: "suppress", Target: "rev", Details: "d"}); err != nil {
		t.Fatalf("AddOversight: %v", err)
	}
	if n := len(svc.ListOversight()); n < 1 {
		t.Fatalf("expected oversight rows, got %d", n)
	}

	td := &TakedownRequest{PageTitle: "PgPage", Reason: "pg", Requester: "pg-req"}
	if err := svc.AddTakedown(td); err != nil {
		t.Fatalf("AddTakedown: %v", err)
	}
	if td.Status != "PENDING" {
		t.Fatalf("default status wrong: %q", td.Status)
	}
	if err := svc.UpdateTakedownStatus(td.ID, "APPROVED"); err != nil {
		t.Fatalf("UpdateTakedownStatus: %v", err)
	}
	if err := svc.UpdateTakedownStatus(uuid.NewString(), "X"); err == nil {
		t.Fatal("expected not-found for bogus takedown")
	}

	svc.RecordRecentChange("pg-change-" + tag)
	if n := len(svc.RecentChanges()); n < 1 {
		t.Fatal("recent ring must hold the record")
	}
}
