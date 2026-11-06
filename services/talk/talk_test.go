package talk

import (
	"testing"
)

func TestTalkWorkflow(t *testing.T) {
	store := NewTalkStore()
	svc := NewTalkService(store)

	pageKey := "0:Quantum_Computing"

	// 1. Create Thread
	thread, err := svc.CreateThread(pageKey, "Discussion regarding Qubit Error Rates", "user-123")
	if err != nil {
		t.Fatalf("CreateThread failed: %v", err)
	}

	threads := svc.ListThreads(pageKey)
	if len(threads) != 1 || threads[0].ID != thread.ID {
		t.Errorf("expected 1 thread, got %d", len(threads))
	}

	// 2. Add Comment with Mention
	comment, err := svc.AddComment(thread.ID, "", "user-456", "Hello @CuratorAlpha, should we add topological qubits?")
	if err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}
	if len(comment.Mentions) != 1 || comment.Mentions[0] != "CuratorAlpha" {
		t.Errorf("mention extraction failed, got: %+v", comment.Mentions)
	}

	comments := svc.GetComments(thread.ID)
	if len(comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(comments))
	}

	// 3. RFC and Voting
	rfc, err := svc.CreateRFC(thread.ID, "Should topological quantum computers be prioritized?", 3)
	if err != nil {
		t.Fatalf("CreateRFC failed: %v", err)
	}

	err = svc.VoteRFC(rfc.ID, "user-1", "support")
	if err != nil {
		t.Fatalf("VoteRFC failed: %v", err)
	}
	err = svc.VoteRFC(rfc.ID, "user-2", "support")
	if err != nil {
		t.Fatalf("VoteRFC 2 failed: %v", err)
	}

	fetchedRFC, err := svc.GetRFC(rfc.ID)
	if err != nil {
		t.Fatalf("GetRFC failed: %v", err)
	}
	if len(fetchedRFC.Votes) != 2 {
		t.Errorf("expected 2 votes, got %d", len(fetchedRFC.Votes))
	}

	if err := svc.CloseRFC(rfc.ID); err != nil {
		t.Fatalf("CloseRFC failed: %v", err)
	}

	// 4. Pin and Lock
	if err := svc.PinThread(thread.ID, true); err != nil {
		t.Fatalf("PinThread failed: %v", err)
	}
	if err := svc.LockThread(thread.ID, true); err != nil {
		t.Fatalf("LockThread failed: %v", err)
	}

	// 5. Arbitration
	arb, err := svc.CreateArbitration(thread.ID, "Dispute between contributors on section scope")
	if err != nil {
		t.Fatalf("CreateArbitration failed: %v", err)
	}
	arbs := svc.ListArbitrations()
	if len(arbs) != 1 || arbs[0].ID != arb.ID {
		t.Errorf("expected 1 arbitration, got %d", len(arbs))
	}
}
