package watchlist

type Service struct {
	store *WatchlistStore
}

func NewService(store *WatchlistStore) *Service {
	return &Service{store: store}
}

func (s *Service) Add(userID, pageKey string, notifyEmail bool) {
	s.store.Add(userID, pageKey, notifyEmail)
}

func (s *Service) Remove(userID, pageKey string) {
	s.store.Remove(userID, pageKey)
}

func (s *Service) List(userID string) []*WatchEntry {
	return s.store.List(userID)
}

func (s *Service) IsWatching(userID, pageKey string) bool {
	return s.store.IsWatching(userID, pageKey)
}