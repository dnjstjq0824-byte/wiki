package governance

import (
	"testing"
	"time"
)

func TestGovernanceFiltersAndBlocks(t *testing.T) {
	store := NewGovernanceStore()
	svc := NewGovernanceService(store)

	// 1. Abuse Filter
	err := svc.AddFilter(&AbuseFilter{
		Name:    "Vandalism Spam",
		Pattern: `(?i)spam_link\.com`,
		Action:  "DISALLOW",
	})
	if err != nil {
		t.Fatalf("AddFilter failed: %v", err)
	}

	filters := svc.ListFilters()
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}

	action, matched := svc.Evaluate("Check this http://spam_link.com now!", "Main_Page", "user1", "1.2.3.4")
	if action != "DISALLOW" || matched == nil {
		t.Errorf("expected DISALLOW, got %s", action)
	}

	action2, matched2 := svc.Evaluate("This is clean educational content", "Main_Page", "user1", "1.2.3.4")
	if action2 != "" || matched2 != nil {
		t.Errorf("clean text should not trigger filter, got %s", action2)
	}

	// 2. Block enforcement
	if svc.IsBlocked("bad_user", "10.0.0.1") {
		t.Errorf("user should not be blocked initially")
	}

	future := time.Now().Add(1 * time.Hour)
	err = svc.AddBlock(&Block{
		TargetUserID: "bad_user",
		TargetIP:     "10.0.0.1",
		BlockedBy:    "sysop",
		Reason:       "Repeated vandalism",
		ExpiresAt:    &future,
		IsSitewide:   true,
	})
	if err != nil {
		t.Fatalf("AddBlock failed: %v", err)
	}

	if !svc.IsBlocked("bad_user", "10.0.0.1") {
		t.Errorf("expected user to be blocked")
	}
	if !svc.IsBlocked("other_user", "10.0.0.1") {
		t.Errorf("expected IP to be blocked")
	}

	// 3. Oversight and Takedown
	svc.AddOversight(&OversightLog{
		ActorID: "oversight_officer",
		Action:  "SUPPRESS_REVISION",
		Target:  "rev-12345",
		Details: "PII telephone number scrubbed",
	})
	logs := svc.ListOversight()
	if len(logs) != 1 {
		t.Errorf("expected 1 oversight log, got %d", len(logs))
	}

	err = svc.AddTakedown(&TakedownRequest{
		PageTitle: "Infringing_Article",
		Reason:    "Copyright Infringement",
		Requester: "Legal Counsel",
	})
	if err != nil {
		t.Fatalf("AddTakedown failed: %v", err)
	}

	takedowns := svc.ListTakedowns()
	if len(takedowns) != 1 {
		t.Fatalf("expected 1 takedown, got %d", len(takedowns))
	}
	if err := svc.UpdateTakedownStatus(takedowns[0].ID, "APPROVED"); err != nil {
		t.Fatalf("UpdateTakedownStatus failed: %v", err)
	}

	// 4. Recent changes broadcast
	svc.RecordRecentChange("Page Edited: Quantum_Computing by CuratorAlpha")
	rcs := svc.RecentChanges()
	if len(rcs) != 1 {
		t.Errorf("expected 1 recent change, got %d", len(rcs))
	}
}

