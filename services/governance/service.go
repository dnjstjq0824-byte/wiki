package governance

type GovernanceService struct {
	store *GovernanceStore
}

func NewGovernanceService(store *GovernanceStore) *GovernanceService {
	return &GovernanceService{store: store}
}

func (s *GovernanceService) AddFilter(f *AbuseFilter) error {
	return s.store.AddFilter(f)
}

func (s *GovernanceService) ListFilters() []*AbuseFilter {
	return s.store.ListFilters()
}

func (s *GovernanceService) Evaluate(text, pageTitle, userID, ip string) (string, *AbuseFilter) {
	return s.store.Evaluate(text, pageTitle, userID, ip)
}

func (s *GovernanceService) AddBlock(b *Block) error {
	return s.store.AddBlock(b)
}

func (s *GovernanceService) IsBlocked(userID, ip string) bool {
	return s.store.IsBlocked(userID, ip)
}

func (s *GovernanceService) ListBlocks() []*Block {
	return s.store.ListBlocks()
}

func (s *GovernanceService) AddOversight(o *OversightLog) {
	s.store.AddOversight(o)
}

func (s *GovernanceService) ListOversight() []OversightLog {
	return s.store.ListOversight()
}

func (s *GovernanceService) RecordRecentChange(desc string) {
	s.store.RecordRecentChange(desc)
}

func (s *GovernanceService) AddTakedown(t *TakedownRequest) error {
	return s.store.AddTakedown(t)
}

func (s *GovernanceService) ListTakedowns() []*TakedownRequest {
	return s.store.ListTakedowns()
}

func (s *GovernanceService) UpdateTakedownStatus(id, status string) error {
	return s.store.UpdateTakedownStatus(id, status)
}
func (s *GovernanceService) RecentChanges() []string {
	return s.store.RecentChanges()
}
