package watchlist

import (
	"testing"
)

func TestWatchlistOperations(t *testing.T) {
	store := NewWatchlistStore()
	svc := NewService(store)

	userID := "user-alice"
	pageKey := "0:Quantum_Computing"

	if svc.IsWatching(userID, pageKey) {
		t.Errorf("user should not be watching initially")
	}

	// 1. Add to Watchlist
	svc.Add(userID, pageKey, true)
	if !svc.IsWatching(userID, pageKey) {
		t.Errorf("expected user to be watching page")
	}

	entries := svc.List(userID)
	if len(entries) != 1 || entries[0].PageKey != pageKey {
		t.Errorf("expected 1 entry, got %+v", entries)
	}

	// Check watchers
	watchers := store.Watchers(pageKey)
	if len(watchers) != 1 || watchers[0] != userID {
		t.Errorf("expected watcher user-alice, got %+v", watchers)
	}

	// 2. Remove from Watchlist
	svc.Remove(userID, pageKey)
	if svc.IsWatching(userID, pageKey) {
		t.Errorf("user should not be watching after removal")
	}
}

