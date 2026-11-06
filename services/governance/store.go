package governance

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

type AbuseFilter struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Pattern   string `json:"pattern"`
	Action    string `json:"action"` // TAG, WARN, DISALLOW, BLOCK
	IsEnabled bool   `json:"is_enabled"`
	compiled  *regexp.Regexp
}

type AbuseFilterLog struct {
	ID             string    `json:"id"`
	FilterID       string    `json:"filter_id"`
	UserID         string    `json:"user_id"`
	IPAddress      string    `json:"ip_address"`
	PageTitle      string    `json:"page_title"`
	MatchedSnippet string    `json:"matched_snippet"`
	ActionTaken    string    `json:"action_taken"`
	CreatedAt      time.Time `json:"created_at"`
}

type Block struct {
	ID           string     `json:"id"`
	TargetUserID string     `json:"target_user_id,omitempty"`
	TargetIP     string     `json:"target_ip,omitempty"`
	BlockedBy    string     `json:"blocked_by"`
	Reason       string     `json:"reason"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsSitewide   bool       `json:"is_sitewide"`
	CreatedAt    time.Time  `json:"created_at"`
}

type OversightLog struct {
	ID        string    `json:"id"`
	ActorID   string    `json:"actor_id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type TakedownRequest struct {
	ID        string    `json:"id"`
	PageTitle string    `json:"page_title"`
	Reason    string    `json:"reason"`
	Requester string    `json:"requester"`
	Status    string    `json:"status"` // PENDING, APPROVED, REJECTED
	CreatedAt time.Time `json:"created_at"`
}

// GovernanceStorage abstracts moderation persistence so the service runs on the
// in-memory GovernanceStore (development/tests) or PostgresGovernanceStore.
// The recent-changes feed and subscriptions stay process-local in both: the
// durable history is the revision ledger, the feed is a live window.
type GovernanceStorage interface {
	AddFilter(f *AbuseFilter) error
	ListFilters() []*AbuseFilter
	Evaluate(text, pageTitle, userID, ip string) (string, *AbuseFilter)
	AddBlock(b *Block) error
	IsBlocked(userID, ip string) bool
	ListBlocks() []*Block
	AddOversight(o *OversightLog) error
	ListOversight() []OversightLog
	RecordRecentChange(desc string)
	RecentChanges() []string
	AddTakedown(t *TakedownRequest) error
	ListTakedowns() []*TakedownRequest
	UpdateTakedownStatus(id, status string) error
	Subscribe() chan string
}

type GovernanceStore struct {
	mu         sync.RWMutex
	filters    map[string]*AbuseFilter
	filterLogs []AbuseFilterLog
	blocks     map[string]*Block
	oversight  []OversightLog
	recent     []string // recent change descriptions
	takedowns  map[string]*TakedownRequest
	subs       []chan string
}

var _ GovernanceStorage = (*GovernanceStore)(nil)

func NewGovernanceStore() *GovernanceStore {
	return &GovernanceStore{
		filters:   make(map[string]*AbuseFilter),
		blocks:    make(map[string]*Block),
		takedowns: make(map[string]*TakedownRequest),
	}
}

func (s *GovernanceStore) AddFilter(f *AbuseFilter) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	c, err := regexp.Compile(f.Pattern)
	if err != nil {
		return err
	}
	f.compiled = c
	if f.Action == "" {
		f.Action = "TAG"
	}
	f.IsEnabled = true
	s.filters[f.ID] = f
	return nil
}

func (s *GovernanceStore) ListFilters() []*AbuseFilter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*AbuseFilter{}
	for _, f := range s.filters {
		cp := *f
		out = append(out, &cp)
	}
	return out
}

func (s *GovernanceStore) Evaluate(text, pageTitle, userID, ip string) (string, *AbuseFilter) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.filters {
		if !f.IsEnabled || f.compiled == nil {
			continue
		}
		if f.compiled.MatchString(text) {
			log := AbuseFilterLog{
				ID:             uuid.New().String(),
				FilterID:       f.ID,
				UserID:         userID,
				IPAddress:      ip,
				PageTitle:      pageTitle,
				MatchedSnippet: text,
				ActionTaken:    f.Action,
				CreatedAt:      time.Now().UTC(),
			}
			s.mu.RUnlock()
			s.mu.Lock()
			s.filterLogs = append(s.filterLogs, log)
			s.mu.Unlock()
			s.mu.RLock()
			return f.Action, f
		}
	}
	return "", nil
}

func (s *GovernanceStore) AddBlock(b *Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	b.CreatedAt = time.Now().UTC()
	s.blocks[b.ID] = b
	return nil
}

func (s *GovernanceStore) IsBlocked(userID, ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()
	for _, b := range s.blocks {
		if b.ExpiresAt != nil && b.ExpiresAt.Before(now) {
			continue
		}
		if b.TargetUserID != "" && b.TargetUserID == userID {
			return true
		}
		if b.TargetIP != "" && b.TargetIP == ip {
			return true
		}
	}
	return false
}

func (s *GovernanceStore) ListBlocks() []*Block {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*Block{}
	for _, b := range s.blocks {
		cp := *b
		out = append(out, &cp)
	}
	return out
}

func (s *GovernanceStore) AddOversight(o *OversightLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	o.CreatedAt = time.Now().UTC()
	s.oversight = append(s.oversight, *o)
	return nil
}

func (s *GovernanceStore) ListOversight() []OversightLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]OversightLog, len(s.oversight))
	copy(out, s.oversight)
	return out
}

func (s *GovernanceStore) RecordRecentChange(desc string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recent = append(s.recent, desc)
	if len(s.recent) > 100 {
		s.recent = s.recent[len(s.recent)-100:]
	}
	for _, ch := range s.subs {
		select {
		case ch <- desc:
		default:
		}
	}
}

func (s *GovernanceStore) RecentChanges() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.recent))
	copy(out, s.recent)
	return out
}

func (s *GovernanceStore) Subscribe() chan string {
	ch := make(chan string, 10)
	s.mu.Lock()
	s.subs = append(s.subs, ch)
	s.mu.Unlock()
	return ch
}
func (s *GovernanceStore) AddTakedown(t *TakedownRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now().UTC()
	if t.Status == "" {
		t.Status = "PENDING"
	}
	s.takedowns[t.ID] = t
	return nil
}

func (s *GovernanceStore) ListTakedowns() []*TakedownRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*TakedownRequest{}
	for _, t := range s.takedowns {
		cp := *t
		out = append(out, &cp)
	}
	return out
}

func (s *GovernanceStore) UpdateTakedownStatus(id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.takedowns[id]
	if !ok {
		return fmt.Errorf("takedown not found")
	}
	t.Status = status
	return nil
}
