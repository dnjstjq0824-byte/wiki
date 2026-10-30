package watchlist

import (
	"sync"
	"time"
)

type WatchEntry struct {
	UserID      string    `json:"user_id"`
	PageKey     string    `json:"page_key"`
	NotifyEmail bool      `json:"notify_email"`
	CreatedAt   time.Time `json:"created_at"`
}

type WatchlistStore struct {
	mu      sync.RWMutex
	entries map[string]*WatchEntry
}

func NewWatchlistStore() *WatchlistStore {
	return &WatchlistStore{entries: make(map[string]*WatchEntry)}
}

func wlKey(userID, pageKey string) string {
	return userID + ":" + pageKey
}

func (s *WatchlistStore) Add(userID, pageKey string, notifyEmail bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[wlKey(userID, pageKey)] = &WatchEntry{
		UserID:      userID,
		PageKey:     pageKey,
		NotifyEmail: notifyEmail,
		CreatedAt:   time.Now().UTC(),
	}
}

func (s *WatchlistStore) Remove(userID, pageKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, wlKey(userID, pageKey))
}

func (s *WatchlistStore) List(userID string) []*WatchEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*WatchEntry{}
	for _, e := range s.entries {
		if e.UserID == userID {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out
}

func (s *WatchlistStore) IsWatching(userID, pageKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.entries[wlKey(userID, pageKey)]
	return ok
}

func (s *WatchlistStore) Watchers(pageKey string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []string{}
	for _, e := range s.entries {
		if e.PageKey == pageKey {
			out = append(out, e.UserID)
		}
	}
	return out
}