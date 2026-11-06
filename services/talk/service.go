package talk

import (
	"fmt"
	"regexp"
	"strings"
)

var mentionRe = regexp.MustCompile(`@([a-zA-Z0-9_]+)`)

type TalkService struct {
	store TalkStorage
}

func NewTalkService(store TalkStorage) *TalkService {
	return &TalkService{store: store}
}

// normalizeVoteChoice maps UI and legacy vote labels to the canonical
// SUPPORT / OPPOSE / NEUTRAL vocabulary the frontend renders.
func normalizeVoteChoice(raw string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "SUPPORT", "PRO":
		return "SUPPORT", nil
	case "OPPOSE", "CON":
		return "OPPOSE", nil
	case "NEUTRAL", "ABSTAIN":
		return "NEUTRAL", nil
	default:
		return "", fmt.Errorf("invalid vote choice %q (want SUPPORT, OPPOSE, or NEUTRAL)", raw)
	}
}

func (s *TalkService) CreateThread(pageKey, title, authorID string) (*TalkThread, error) {
	t := &TalkThread{PageKey: pageKey, Title: title, AuthorID: authorID}
	if err := s.store.CreateThread(t); err != nil {
		return nil, err
	}
	s.store.Broadcast(pageKey, "thread_created:"+t.ID)
	return t, nil
}

func (s *TalkService) ListThreads(pageKey string) []*TalkThread {
	return s.store.ListThreads(pageKey)
}

func (s *TalkService) SetStatus(threadID string, status TalkStatus) error {
	return s.store.UpdateThreadStatus(threadID, status)
}

func (s *TalkService) AddComment(threadID, parentID, authorID, content string) (*TalkComment, error) {
	mentions := extractMentions(content)
	c := &TalkComment{
		ThreadID: threadID,
		ParentID: parentID,
		AuthorID: authorID,
		Content:  content,
		Mentions: mentions,
	}
	if err := s.store.AddComment(c); err != nil {
		return nil, err
	}
	// broadcast
	thread, _ := s.store.GetThread(threadID)
	if thread != nil {
		s.store.Broadcast(thread.PageKey, "comment_added:"+c.ID)
	}
	return c, nil
}

func (s *TalkService) GetComments(threadID string) []*TalkComment {
	return s.store.GetComments(threadID)
}

func (s *TalkService) CreateRFC(threadID, question string, quorum int) (*RFCPoll, error) {
	r := &RFCPoll{ThreadID: threadID, Question: question, Quorum: quorum}
	if err := s.store.CreateRFC(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *TalkService) VoteRFC(rfcID, userID, choice string) error {
	canonical, err := normalizeVoteChoice(choice)
	if err != nil {
		return err
	}
	return s.store.VoteRFC(rfcID, RFCVote{UserID: userID, Choice: canonical})
}

func (s *TalkService) GetRFC(id string) (*RFCPoll, error) {
	return s.store.GetRFC(id)
}

func (s *TalkService) GetRFCByThread(threadID string) (*RFCPoll, error) {
	return s.store.GetRFCByThread(threadID)
}

func (s *TalkService) CloseRFC(id string) error {
	return s.store.CloseRFC(id)
}

func (s *TalkService) PinThread(id string, pinned bool) error {
	return s.store.PinThread(id, pinned)
}

func (s *TalkService) LockThread(id string, locked bool) error {
	return s.store.LockThread(id, locked)
}
func (s *TalkService) Subscribe(pageKey string) (chan string, func()) {
	ch := s.store.Subscribe(pageKey)
	cancel := func() { s.store.Unsubscribe(pageKey, ch) }
	return ch, cancel
}

func (s *TalkService) ArchiveThread(id string) error {
	return s.store.ArchiveThread(id)
}

func (s *TalkService) ListArchived() []*TalkThread {
	return s.store.ListArchived()
}

func (s *TalkService) CreateArbitration(threadID, summary string) (*ArbitrationCase, error) {
	a := &ArbitrationCase{ThreadID: threadID, Summary: summary}
	if err := s.store.CreateArbitration(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *TalkService) ListArbitrations() []*ArbitrationCase {
	return s.store.ListArbitrations()
}
func (s *TalkService) Broadcast(pageKey, msg string) {
	s.store.Broadcast(pageKey, msg)
}
func extractMentions(content string) []string {
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	seen := map[string]bool{}
	out := []string{}
	for _, m := range matches {
		if len(m) > 1 {
			name := m[1]
			if !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	return out
}
