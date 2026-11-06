package talk

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	treepg "github.com/dnjstjqppk-dev/treewiki/pkg/pg"
)

// PostgresTalkStore persists threads, replies, polls, and arbitration cases in
// PostgreSQL. Threads are keyed by page key ("ns:Title") with an optional page
// link; realtime fan-out stays process-local like the memory store.
type PostgresTalkStore struct {
	pool *pgxpool.Pool

	mu   sync.Mutex
	subs map[string][]chan string
}

var _ TalkStorage = (*PostgresTalkStore)(nil)

// NewPostgresTalkStore creates a PostgreSQL-backed talk store.
func NewPostgresTalkStore(pool *pgxpool.Pool) *PostgresTalkStore {
	return &PostgresTalkStore{pool: pool, subs: make(map[string][]chan string)}
}

var validThreadStatuses = map[TalkStatus]bool{
	StatusOpen:         true,
	"UNDER_DISCUSSION": true,
	StatusRFC:          true,
	StatusClosed:       true,
	StatusResolved:     true,
}

func checkThreadStatus(status TalkStatus) error {
	if !validThreadStatuses[status] {
		return errors.New("invalid thread status")
	}
	return nil
}

// splitAuthor maps a display author to UUID / IP / name columns.
func splitAuthor(raw string) (id pgtype.UUID, ip *net.IPNet, name string) {
	id = treepg.UUIDVal(raw)
	if id.Valid {
		return id, nil, ""
	}
	if parsed := treepg.InetVal(raw); parsed != nil {
		return pgtype.UUID{}, parsed, ""
	}
	return pgtype.UUID{}, nil, raw
}

const threadSelectColumns = `
	SELECT t.id, t.page_key, t.title, t.status, t.author_name, t.author_ip,
	       u.username, t.is_pinned, t.is_locked, t.created_at, t.updated_at
	FROM talk_threads t
	LEFT JOIN users u ON u.id = t.author_id
	`

func scanThread(row pgx.Row) (*TalkThread, error) {
	var (
		id       pgtype.UUID
		pageKey  pgtype.Text
		title    string
		status   string
		authName pgtype.Text
		authIP   *net.IPNet
		username pgtype.Text
		pinned   bool
		locked   bool
		created  time.Time
		updated  time.Time
	)
	if err := row.Scan(&id, &pageKey, &title, &status, &authName, &authIP,
		&username, &pinned, &locked, &created, &updated); err != nil {
		return nil, err
	}
	return &TalkThread{
		ID:        treepg.UUIDString(id),
		PageKey:   treepg.TextOrEmpty(pageKey),
		Title:     title,
		Status:    TalkStatus(status),
		AuthorID:  treepg.DisplayAuthor(treepg.TextOrEmpty(username), treepg.TextOrEmpty(authName), treepg.InetString(authIP), ""),
		Pinned:    pinned,
		Locked:    locked,
		CreatedAt: created,
		UpdatedAt: updated,
	}, nil
}

// CreateThread opens a discussion thread on a page key.
func (s *PostgresTalkStore) CreateThread(t *TalkThread) error {
	if t.ID == "" {
		t.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(t.ID); !ok {
		t.ID = treepg.NewUUID()
	}
	if t.Status == "" {
		t.Status = StatusOpen
	}
	if err := checkThreadStatus(t.Status); err != nil {
		return err
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now
	authorID, authorIP, authorName := splitAuthor(t.AuthorID)
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO talk_threads (id, page_key, title, author_id, author_ip, author_name,
		                          status, is_locked, is_pinned, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, treepg.UUIDVal(t.ID), treepg.TextVal(t.PageKey), treepg.TruncateRunes(t.Title, 255),
		authorID, authorIP, treepg.TextVal(authorName), string(t.Status),
		t.Locked, t.Pinned, now, now)
	if err != nil {
		if treepg.IsFKViolation(err) && authorID.Valid {
			// Unknown author UUID; keep it as a display name.
			_, err = s.pool.Exec(ctx, `
				INSERT INTO talk_threads (id, page_key, title, author_name,
				                          status, is_locked, is_pinned, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, treepg.UUIDVal(t.ID), treepg.TextVal(t.PageKey), treepg.TruncateRunes(t.Title, 255),
				t.AuthorID, string(t.Status), t.Locked, t.Pinned, now, now)
		}
		if err != nil {
			return err
		}
	}
	// Opportunistic page linkage (best effort; threads stand alone by key).
	s.linkThreadPage(ctx, t.ID, t.PageKey)
	return nil
}

// linkThreadPage attaches a thread to its page row when one exists.
func (s *PostgresTalkStore) linkThreadPage(ctx context.Context, threadID, pageKey string) {
	parts := strings.SplitN(pageKey, ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return
	}
	ns, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	var pageID pgtype.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT id FROM pages WHERE namespace_id = $1 AND title = $2
	`, ns, parts[1]).Scan(&pageID)
	if err != nil || !pageID.Valid {
		return
	}
	_, _ = s.pool.Exec(ctx, `UPDATE talk_threads SET page_id = $1 WHERE id = $2`,
		pageID, treepg.UUIDVal(threadID))
}

// GetThread loads one live (non-archived) thread.
func (s *PostgresTalkStore) GetThread(id string) (*TalkThread, error) {
	threadID, ok := treepg.ParseUUID(id)
	if !ok {
		return nil, errors.New("thread not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	t, err := scanThread(s.pool.QueryRow(ctx, threadSelectColumns+`
		WHERE t.id = $1 AND t.is_archived = FALSE
	`, treepg.UUIDVal(threadID.String())))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("thread not found")
		}
		return nil, err
	}
	return t, nil
}

// ListThreads lists live threads for a page key, pinned first.
func (s *PostgresTalkStore) ListThreads(pageKey string) []*TalkThread {
	out := []*TalkThread{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, threadSelectColumns+`
		WHERE t.page_key = $1 AND t.is_archived = FALSE
		ORDER BY t.is_pinned DESC, t.created_at ASC
	`, pageKey)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		if t, err := scanThread(rows); err == nil {
			out = append(out, t)
		}
	}
	return out
}

// UpdateThreadStatus transitions a thread status.
func (s *PostgresTalkStore) UpdateThreadStatus(id string, status TalkStatus) error {
	if err := checkThreadStatus(status); err != nil {
		return err
	}
	threadID, ok := treepg.ParseUUID(id)
	if !ok {
		return errors.New("thread not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `
		UPDATE talk_threads SET status = $1, updated_at = NOW()
		WHERE id = $2 AND is_archived = FALSE
	`, string(status), treepg.UUIDVal(threadID.String()))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("thread not found")
	}
	return nil
}

const commentSelectColumns = `
	SELECT c.id, c.parent_reply_id, c.author_name, c.author_ip, u.username,
	       c.content, c.mentions, c.created_at
	FROM talk_replies c
	LEFT JOIN users u ON u.id = c.author_id
	`

func scanComment(threadID string, row pgx.Row) (*TalkComment, error) {
	var (
		id       pgtype.UUID
		parent   pgtype.UUID
		authName pgtype.Text
		authIP   *net.IPNet
		username pgtype.Text
		content  string
		mentions []string
		created  time.Time
	)
	if err := row.Scan(&id, &parent, &authName, &authIP, &username,
		&content, &mentions, &created); err != nil {
		return nil, err
	}
	if mentions == nil {
		mentions = []string{}
	}
	return &TalkComment{
		ID:        treepg.UUIDString(id),
		ThreadID:  threadID,
		ParentID:  treepg.UUIDString(parent),
		AuthorID:  treepg.DisplayAuthor(treepg.TextOrEmpty(username), treepg.TextOrEmpty(authName), treepg.InetString(authIP), ""),
		Content:   content,
		Mentions:  mentions,
		CreatedAt: created,
	}, nil
}

// AddComment appends a reply and bumps the thread counters.
func (s *PostgresTalkStore) AddComment(c *TalkComment) error {
	threadID, ok := treepg.ParseUUID(c.ThreadID)
	if !ok {
		return errors.New("thread not found")
	}
	if c.ID == "" {
		c.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(c.ID); !ok {
		c.ID = treepg.NewUUID()
	}
	var parent pgtype.UUID
	if c.ParentID != "" {
		p, ok := treepg.ParseUUID(c.ParentID)
		if !ok {
			return errors.New("invalid parent comment")
		}
		parent = pgtype.UUID{Bytes: p, Valid: true}
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	if c.Mentions == nil {
		c.Mentions = []string{}
	}
	authorID, authorIP, authorName := splitAuthor(c.AuthorID)
	ctx, cancel := treepg.Ctx()
	defer cancel()
	insert := func(aid pgtype.UUID, aname string) error {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO talk_replies (id, thread_id, parent_reply_id, author_id,
			                          author_ip, author_name, content, mentions, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, treepg.UUIDVal(c.ID), treepg.UUIDVal(threadID.String()), parent,
			aid, authorIP, treepg.TextVal(aname), c.Content, c.Mentions, now)
		return err
	}
	if err := insert(authorID, authorName); err != nil {
		if treepg.IsFKViolation(err) && authorID.Valid {
			if retryErr := insert(pgtype.UUID{}, c.AuthorID); retryErr != nil {
				return retryErr
			}
		} else if treepg.IsFKViolation(err) {
			return errors.New("thread not found")
		} else {
			return err
		}
	}
	_, _ = s.pool.Exec(ctx, `
		UPDATE talk_threads SET reply_count = reply_count + 1, updated_at = NOW()
		WHERE id = $1
	`, treepg.UUIDVal(threadID.String()))
	return nil
}

// GetComments lists replies oldest-first.
func (s *PostgresTalkStore) GetComments(threadID string) []*TalkComment {
	out := []*TalkComment{}
	threadUUID, ok := treepg.ParseUUID(threadID)
	if !ok {
		return out
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, commentSelectColumns+`
		WHERE c.thread_id = $1 ORDER BY c.created_at ASC, c.id ASC
	`, treepg.UUIDVal(threadUUID.String()))
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		if c, err := scanComment(threadID, rows); err == nil {
			out = append(out, c)
		}
	}
	return out
}

// CreateRFC opens a consensus poll on a thread (30-day default expiry).
func (s *PostgresTalkStore) CreateRFC(r *RFCPoll) error {
	threadID, ok := treepg.ParseUUID(r.ThreadID)
	if !ok {
		return errors.New("thread not found")
	}
	if r.ID == "" {
		r.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(r.ID); !ok {
		r.ID = treepg.NewUUID()
	}
	now := time.Now().UTC()
	r.CreatedAt = now
	if r.Votes == nil {
		r.Votes = []RFCVote{}
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO talk_polls (id, thread_id, question, quorum, expires_at, is_closed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, treepg.UUIDVal(r.ID), treepg.UUIDVal(threadID.String()),
		treepg.TruncateRunes(r.Question, 255), r.Quorum,
		now.Add(30*24*time.Hour), r.Closed, now)
	if err != nil {
		if treepg.IsFKViolation(err) {
			return errors.New("thread not found")
		}
		return err
	}
	return nil
}

func (s *PostgresTalkStore) pollVotes(ctx context.Context, pollID pgtype.UUID) []RFCVote {
	out := []RFCVote{}
	rows, err := s.pool.Query(ctx, `
		SELECT v.user_id, v.voter_name, v.choice
		FROM talk_poll_votes v
		WHERE v.poll_id = $1
		ORDER BY v.created_at ASC
	`, pollID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var (
			uid  pgtype.UUID
			name pgtype.Text
			cho  string
		)
		if err := rows.Scan(&uid, &name, &cho); err != nil {
			continue
		}
		voter := treepg.TextOrEmpty(name)
		if voter == "" {
			voter = treepg.UUIDString(uid)
		}
		out = append(out, RFCVote{UserID: voter, Choice: cho})
	}
	return out
}

// GetRFC loads a poll with its votes.
func (s *PostgresTalkStore) GetRFC(id string) (*RFCPoll, error) {
	pollID, ok := treepg.ParseUUID(id)
	if !ok {
		return nil, errors.New("rfc not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var (
		threadID pgtype.UUID
		question string
		quorum   int
		closed   bool
		created  time.Time
	)
	err := s.pool.QueryRow(ctx, `
		SELECT thread_id, question, quorum, is_closed, created_at
		FROM talk_polls WHERE id = $1
	`, treepg.UUIDVal(pollID.String())).Scan(&threadID, &question, &quorum, &closed, &created)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("rfc not found")
		}
		return nil, err
	}
	return &RFCPoll{
		ID:        treepg.UUIDString(treepg.UUIDVal(pollID.String())),
		ThreadID:  treepg.UUIDString(threadID),
		Question:  question,
		Votes:     s.pollVotes(ctx, treepg.UUIDVal(pollID.String())),
		Quorum:    quorum,
		Closed:    closed,
		CreatedAt: created,
	}, nil
}

// GetRFCByThread loads the latest poll attached to a thread.
func (s *PostgresTalkStore) GetRFCByThread(threadID string) (*RFCPoll, error) {
	threadUUID, ok := treepg.ParseUUID(threadID)
	if !ok {
		return nil, errors.New("no rfc attached to thread")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var id pgtype.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM talk_polls WHERE thread_id = $1
		ORDER BY created_at DESC LIMIT 1
	`, treepg.UUIDVal(threadUUID.String())).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("no rfc attached to thread")
		}
		return nil, err
	}
	return s.GetRFC(treepg.UUIDString(id))
}

// VoteRFC records (or replaces) one ballot. Registered users vote by UUID;
// display-name voters are matched by name.
func (s *PostgresTalkStore) VoteRFC(id string, vote RFCVote) error {
	pollID, ok := treepg.ParseUUID(id)
	if !ok {
		return errors.New("rfc not found")
	}
	choice, err := normalizeVoteChoice(vote.Choice)
	if err != nil {
		return err
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var closed bool
	err = s.pool.QueryRow(ctx, `SELECT is_closed FROM talk_polls WHERE id = $1`,
		treepg.UUIDVal(pollID.String())).Scan(&closed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("rfc not found")
		}
		return err
	}
	if closed {
		return errors.New("rfc closed")
	}
	if voterID, ok := treepg.ParseUUID(vote.UserID); ok {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO talk_poll_votes (poll_id, user_id, choice)
			VALUES ($1, $2, $3)
			ON CONFLICT (poll_id, user_id) DO UPDATE SET choice = EXCLUDED.choice
		`, treepg.UUIDVal(pollID.String()), treepg.UUIDVal(voterID.String()), choice)
		if err != nil {
			if treepg.IsFKViolation(err) {
				// Not a registered user UUID; fall through to name vote.
				return s.voteByName(ctx, treepg.UUIDVal(pollID.String()), vote.UserID, choice)
			}
			return err
		}
		return nil
	}
	return s.voteByName(ctx, treepg.UUIDVal(pollID.String()), vote.UserID, choice)
}

func (s *PostgresTalkStore) voteByName(ctx context.Context, pollID pgtype.UUID, name, choice string) error {
	var existing pgtype.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM talk_poll_votes WHERE poll_id = $1 AND voter_name = $2
	`, pollID, name).Scan(&existing)
	if err == nil && existing.Valid {
		_, err = s.pool.Exec(ctx, `
			UPDATE talk_poll_votes SET choice = $1 WHERE id = $2
		`, choice, existing)
		return err
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO talk_poll_votes (poll_id, voter_name, choice)
		VALUES ($1, $2, $3)
		ON CONFLICT (poll_id, voter_name) WHERE voter_name IS NOT NULL
		DO UPDATE SET choice = EXCLUDED.choice
	`, pollID, treepg.TextVal(name), choice)
	return err
}

// CloseRFC closes a poll to further ballots.
func (s *PostgresTalkStore) CloseRFC(id string) error {
	pollID, ok := treepg.ParseUUID(id)
	if !ok {
		return errors.New("rfc not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `UPDATE talk_polls SET is_closed = TRUE WHERE id = $1`,
		treepg.UUIDVal(pollID.String()))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("rfc not found")
	}
	return nil
}

// Subscribe registers a process-local SSE channel for a page key.
func (s *PostgresTalkStore) Subscribe(pageKey string) chan string {
	ch := make(chan string, 10)
	s.mu.Lock()
	s.subs[pageKey] = append(s.subs[pageKey], ch)
	s.mu.Unlock()
	return ch
}

// Unsubscribe removes an SSE channel.
func (s *PostgresTalkStore) Unsubscribe(pageKey string, ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := []chan string{}
	for _, c := range s.subs[pageKey] {
		if c != ch {
			kept = append(kept, c)
		}
	}
	s.subs[pageKey] = kept
	close(ch)
}

// Broadcast fans a message out to page subscribers (non-blocking).
func (s *PostgresTalkStore) Broadcast(pageKey, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.subs[pageKey] {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (s *PostgresTalkStore) setThreadFlag(id, column string, value bool) error {
	threadID, ok := treepg.ParseUUID(id)
	if !ok {
		return errors.New("thread not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `
		UPDATE talk_threads SET `+column+` = $1, updated_at = NOW()
		WHERE id = $2 AND is_archived = FALSE
	`, value, treepg.UUIDVal(threadID.String()))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("thread not found")
	}
	return nil
}

// PinThread pins or unpins a thread.
func (s *PostgresTalkStore) PinThread(id string, pinned bool) error {
	return s.setThreadFlag(id, "is_pinned", pinned)
}

// LockThread locks or unlocks a thread.
func (s *PostgresTalkStore) LockThread(id string, locked bool) error {
	return s.setThreadFlag(id, "is_locked", locked)
}

// ArchiveThread retires a thread from the live listing.
func (s *PostgresTalkStore) ArchiveThread(id string) error {
	threadID, ok := treepg.ParseUUID(id)
	if !ok {
		return errors.New("thread not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `
		UPDATE talk_threads
		SET is_archived = TRUE, status = 'CLOSED', updated_at = NOW()
		WHERE id = $1 AND is_archived = FALSE
	`, treepg.UUIDVal(threadID.String()))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("thread not found")
	}
	return nil
}

// ListArchived lists retired threads newest-first.
func (s *PostgresTalkStore) ListArchived() []*TalkThread {
	out := []*TalkThread{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, threadSelectColumns+`
		WHERE t.is_archived = TRUE ORDER BY t.updated_at DESC
	`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		if t, err := scanThread(rows); err == nil {
			out = append(out, t)
		}
	}
	return out
}

// CreateArbitration opens a formal dispute case on a thread.
func (s *PostgresTalkStore) CreateArbitration(a *ArbitrationCase) error {
	threadID, ok := treepg.ParseUUID(a.ThreadID)
	if !ok {
		return errors.New("thread not found")
	}
	if a.ID == "" {
		a.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(a.ID); !ok {
		a.ID = treepg.NewUUID()
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	if a.Status == "" {
		a.Status = "OPEN"
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO arbitration_cases (id, thread_id, summary, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, treepg.UUIDVal(a.ID), treepg.UUIDVal(threadID.String()), a.Summary, a.Status, now)
	if err != nil {
		if treepg.IsFKViolation(err) {
			return errors.New("thread not found")
		}
		return err
	}
	return nil
}

// ListArbitrations lists dispute cases newest-first.
func (s *PostgresTalkStore) ListArbitrations() []*ArbitrationCase {
	out := []*ArbitrationCase{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT id, thread_id, summary, status, created_at
		FROM arbitration_cases ORDER BY created_at DESC
	`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id       pgtype.UUID
			threadID pgtype.UUID
			summary  string
			status   string
			created  time.Time
		)
		if err := rows.Scan(&id, &threadID, &summary, &status, &created); err != nil {
			continue
		}
		out = append(out, &ArbitrationCase{
			ID:        treepg.UUIDString(id),
			ThreadID:  treepg.UUIDString(threadID),
			Summary:   summary,
			Status:    status,
			CreatedAt: created,
		})
	}
	return out
}
