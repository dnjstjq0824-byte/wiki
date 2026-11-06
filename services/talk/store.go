package talk

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TalkStatus string

const (
	StatusOpen     TalkStatus = "OPEN"
	StatusRFC      TalkStatus = "RFC"
	StatusClosed   TalkStatus = "CLOSED"
	StatusResolved TalkStatus = "RESOLVED"
)

type TalkThread struct {
	ID        string     `json:"id"`
	PageKey   string     `json:"page_key"`
	Title     string     `json:"title"`
	Status    TalkStatus `json:"status"`
	AuthorID  string     `json:"author_id"`
	Pinned    bool       `json:"pinned"`
	Locked    bool       `json:"locked"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type TalkComment struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	ParentID  string    `json:"parent_id,omitempty"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	Mentions  []string  `json:"mentions"`
	CreatedAt time.Time `json:"created_at"`
}

type RFCVote struct {
	UserID string `json:"user_id"`
	Choice string `json:"choice"` // support, oppose, neutral
}

type RFCPoll struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	Question  string    `json:"question"`
	Votes     []RFCVote `json:"votes"`
	Quorum    int       `json:"quorum"`
	Closed    bool      `json:"closed"`
	CreatedAt time.Time `json:"created_at"`
}

type ArbitrationCase struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	Summary   string    `json:"summary"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// TalkStorage abstracts discussion persistence so the talk service runs on the
// in-memory TalkStore (development/tests) or PostgresTalkStore (production).
// Realtime fan-out (Subscribe/Unsubscribe/Broadcast) is process-local in both.
type TalkStorage interface {
	CreateThread(t *TalkThread) error
	GetThread(id string) (*TalkThread, error)
	ListThreads(pageKey string) []*TalkThread
	UpdateThreadStatus(id string, status TalkStatus) error
	AddComment(c *TalkComment) error
	GetComments(threadID string) []*TalkComment
	CreateRFC(r *RFCPoll) error
	GetRFC(id string) (*RFCPoll, error)
	GetRFCByThread(threadID string) (*RFCPoll, error)
	VoteRFC(id string, vote RFCVote) error
	CloseRFC(id string) error
	Subscribe(pageKey string) chan string
	Unsubscribe(pageKey string, ch chan string)
	Broadcast(pageKey, msg string)
	PinThread(id string, pinned bool) error
	LockThread(id string, locked bool) error
	ArchiveThread(id string) error
	ListArchived() []*TalkThread
	CreateArbitration(a *ArbitrationCase) error
	ListArbitrations() []*ArbitrationCase
}

type TalkStore struct {
	mu              sync.RWMutex
	threads         map[string]*TalkThread
	comments        map[string][]*TalkComment // threadID
	rfcs            map[string]*RFCPoll
	subs            map[string][]chan string // pageKey -> SSE channels
	ArchivedThreads map[string]*TalkThread
	Arbitrations    map[string]*ArbitrationCase
}

var _ TalkStorage = (*TalkStore)(nil)

func NewTalkStore() *TalkStore {
	return &TalkStore{
		threads:         make(map[string]*TalkThread),
		comments:        make(map[string][]*TalkComment),
		rfcs:            make(map[string]*RFCPoll),
		subs:            make(map[string][]chan string),
		ArchivedThreads: make(map[string]*TalkThread),
		Arbitrations:    make(map[string]*ArbitrationCase),
	}
}

func (s *TalkStore) CreateThread(t *TalkThread) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now
	if t.Status == "" {
		t.Status = StatusOpen
	}
	s.threads[t.ID] = t
	return nil
}

func (s *TalkStore) GetThread(id string) (*TalkThread, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.threads[id]
	if !ok {
		return nil, fmt.Errorf("thread not found")
	}
	cp := *t
	return &cp, nil
}

func (s *TalkStore) ListThreads(pageKey string) []*TalkThread {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*TalkThread{}
	for _, t := range s.threads {
		if t.PageKey == pageKey {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out
}

func (s *TalkStore) UpdateThreadStatus(id string, status TalkStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.threads[id]
	if !ok {
		return fmt.Errorf("thread not found")
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *TalkStore) AddComment(c *TalkComment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	c.CreatedAt = time.Now().UTC()
	if c.Mentions == nil {
		c.Mentions = []string{}
	}
	s.comments[c.ThreadID] = append(s.comments[c.ThreadID], c)
	return nil
}

func (s *TalkStore) GetComments(threadID string) []*TalkComment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*TalkComment, len(s.comments[threadID]))
	copy(out, s.comments[threadID])
	return out
}

func (s *TalkStore) CreateRFC(r *RFCPoll) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	r.CreatedAt = time.Now().UTC()
	if r.Votes == nil {
		r.Votes = []RFCVote{}
	}
	s.rfcs[r.ID] = r
	return nil
}

func (s *TalkStore) GetRFC(id string) (*RFCPoll, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rfcs[id]
	if !ok {
		return nil, fmt.Errorf("rfc not found")
	}
	cp := *r
	cp.Votes = append([]RFCVote{}, r.Votes...)
	return &cp, nil
}

func (s *TalkStore) VoteRFC(id string, vote RFCVote) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rfcs[id]
	if !ok {
		return fmt.Errorf("rfc not found")
	}
	if r.Closed {
		return fmt.Errorf("rfc closed")
	}
	// replace existing vote
	for i, v := range r.Votes {
		if v.UserID == vote.UserID {
			r.Votes[i] = vote
			return nil
		}
	}
	r.Votes = append(r.Votes, vote)
	return nil
}

func (s *TalkStore) GetRFCByThread(threadID string) (*RFCPoll, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rfcs {
		if r.ThreadID == threadID {
			cp := *r
			cp.Votes = append([]RFCVote{}, r.Votes...)
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("no rfc attached to thread")
}

func (s *TalkStore) CloseRFC(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rfcs[id]
	if !ok {
		return fmt.Errorf("rfc not found")
	}
	r.Closed = true
	return nil
}

// SSE
func (s *TalkStore) Subscribe(pageKey string) chan string {
	ch := make(chan string, 10)
	s.mu.Lock()
	s.subs[pageKey] = append(s.subs[pageKey], ch)
	s.mu.Unlock()
	return ch
}

func (s *TalkStore) Unsubscribe(pageKey string, ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs := s.subs[pageKey]
	out := []chan string{}
	for _, c := range subs {
		if c != ch {
			out = append(out, c)
		}
	}
	s.subs[pageKey] = out
	close(ch)
}

func (s *TalkStore) Broadcast(pageKey, msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.subs[pageKey] {
		select {
		case ch <- msg:
		default:
		}
	}
}
func (s *TalkStore) PinThread(id string, pinned bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.threads[id]
	if !ok {
		return fmt.Errorf("thread not found")
	}
	t.Pinned = pinned
	t.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *TalkStore) LockThread(id string, locked bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.threads[id]
	if !ok {
		return fmt.Errorf("thread not found")
	}
	t.Locked = locked
	t.UpdatedAt = time.Now().UTC()
	return nil
}
func (s *TalkStore) ArchiveThread(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.threads[id]
	if !ok {
		return fmt.Errorf("thread not found")
	}
	s.ArchivedThreads[id] = t
	delete(s.threads, id)
	t.Status = StatusClosed
	return nil
}

func (s *TalkStore) ListArchived() []*TalkThread {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*TalkThread{}
	for _, t := range s.ArchivedThreads {
		cp := *t
		out = append(out, &cp)
	}
	return out
}

func (s *TalkStore) CreateArbitration(a *ArbitrationCase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	a.CreatedAt = time.Now().UTC()
	if a.Status == "" {
		a.Status = "OPEN"
	}
	s.Arbitrations[a.ID] = a
	return nil
}

func (s *TalkStore) ListArbitrations() []*ArbitrationCase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*ArbitrationCase{}
	for _, a := range s.Arbitrations {
		cp := *a
		out = append(out, &cp)
	}
	return out
}
