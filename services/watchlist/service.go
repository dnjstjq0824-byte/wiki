package watchlist

type Service struct {
	store WatchlistStorage
}

func NewService(store WatchlistStorage) *Service {
	return &Service{store: store}
}

func (s *Service) Add(userID, pageKey string, notifyEmail bool) error {
	return s.store.Add(userID, pageKey, notifyEmail)
}

func (s *Service) Remove(userID, pageKey string) error {
	return s.store.Remove(userID, pageKey)
}

func (s *Service) List(userID string) []*WatchEntry {
	return s.store.List(userID)
}

func (s *Service) IsWatching(userID, pageKey string) bool {
	return s.store.IsWatching(userID, pageKey)
}

func (s *Service) Watchers(pageKey string) []string {
	return s.store.Watchers(pageKey)
}
