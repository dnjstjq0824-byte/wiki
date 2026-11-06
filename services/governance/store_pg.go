package governance

import (
	"errors"
	"log"
	"net"
	"regexp"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	treepg "github.com/dnjstjqppk-dev/treewiki/pkg/pg"
)

// PostgresGovernanceStore persists filters, blocks, oversight entries, and
// takedowns in PostgreSQL. The recent-changes ring and subscriptions stay
// process-local (the durable history is the revision ledger).
type PostgresGovernanceStore struct {
	pool *pgxpool.Pool

	mu     sync.Mutex
	recent []string
	subs   []chan string
}

var _ GovernanceStorage = (*PostgresGovernanceStore)(nil)

// NewPostgresGovernanceStore creates a PostgreSQL-backed governance store.
func NewPostgresGovernanceStore(pool *pgxpool.Pool) *PostgresGovernanceStore {
	return &PostgresGovernanceStore{pool: pool}
}

var validFilterActions = map[string]bool{
	"TAG": true, "WARN": true, "DISALLOW": true, "BLOCK": true,
}

// AddFilter validates and persists an abuse filter.
func (s *PostgresGovernanceStore) AddFilter(f *AbuseFilter) error {
	if f.ID == "" {
		f.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(f.ID); !ok {
		f.ID = treepg.NewUUID()
	}
	c, err := regexp.Compile(f.Pattern)
	if err != nil {
		return err
	}
	f.compiled = c
	if f.Action == "" {
		f.Action = "TAG"
	}
	if !validFilterActions[f.Action] {
		return errors.New("invalid filter action (want TAG, WARN, DISALLOW, or BLOCK)")
	}
	f.IsEnabled = true
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO abuse_filters (id, name, description, pattern_regex, action, is_enabled)
		VALUES ($1, $2, $3, $4, $5, TRUE)
	`, treepg.UUIDVal(f.ID), treepg.TruncateRunes(f.Name, 128), "",
		f.Pattern, f.Action)
	return err
}

// ListFilters lists filters ordered by name with patterns compiled.
func (s *PostgresGovernanceStore) ListFilters() []*AbuseFilter {
	out := []*AbuseFilter{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, pattern_regex, action, is_enabled
		FROM abuse_filters ORDER BY name
	`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id      pgtype.UUID
			name    string
			pattern string
			action  string
			enabled bool
		)
		if err := rows.Scan(&id, &name, &pattern, &action, &enabled); err != nil {
			continue
		}
		f := &AbuseFilter{
			ID:        treepg.UUIDString(id),
			Name:      name,
			Pattern:   pattern,
			Action:    action,
			IsEnabled: enabled,
		}
		if c, err := regexp.Compile(pattern); err == nil {
			f.compiled = c
		}
		out = append(out, f)
	}
	return out
}

// Evaluate matches text against enabled filters and logs the hit (best
// effort: a logging failure never blocks the moderation decision).
func (s *PostgresGovernanceStore) Evaluate(text, pageTitle, userID, ip string) (string, *AbuseFilter) {
	for _, f := range s.ListFilters() {
		if !f.IsEnabled || f.compiled == nil || !f.compiled.MatchString(text) {
			continue
		}
		ctx, cancel := treepg.Ctx()
		_, err := s.pool.Exec(ctx, `
			INSERT INTO abuse_filter_logs (filter_id, user_id, actor_name, ip_address,
			                               page_title, matched_snippet, action_taken)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, treepg.UUIDVal(f.ID), treepg.UUIDVal(userID), treepg.TextVal(userID),
			treepg.InetVal(ip), treepg.TruncateRunes(pageTitle, 255),
			treepg.TruncateRunes(text, 2000), f.Action)
		cancel()
		if err != nil {
			log.Printf("[governance] filter log insert failed: %v", err)
		}
		return f.Action, f
	}
	return "", nil
}

// AddBlock persists an account/IP block.
func (s *PostgresGovernanceStore) AddBlock(b *Block) error {
	if b.ID == "" {
		b.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(b.ID); !ok {
		b.ID = treepg.NewUUID()
	}
	now := time.Now().UTC()
	b.CreatedAt = now
	targetUserID := treepg.UUIDVal(b.TargetUserID)
	targetUserName := ""
	if !targetUserID.Valid {
		targetUserName = b.TargetUserID
	}
	targetIP := treepg.InetVal(b.TargetIP)
	targetIPRaw := ""
	if targetIP == nil {
		targetIPRaw = b.TargetIP
	}
	blockedByID := treepg.UUIDVal(b.BlockedBy)
	blockedByName := ""
	if !blockedByID.Valid {
		blockedByName = b.BlockedBy
	}
	var expires pgtype.Timestamptz
	if b.ExpiresAt != nil {
		expires = pgtype.Timestamptz{Time: *b.ExpiresAt, Valid: true}
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO blocks (id, target_user_id, target_user_name, target_ip, target_ip_raw,
		                    blocked_by_id, blocked_by_name, reason, expires_at,
		                    is_sitewide, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, treepg.UUIDVal(b.ID), targetUserID, treepg.TextVal(targetUserName),
		targetIP, treepg.TextVal(targetIPRaw), blockedByID, treepg.TextVal(blockedByName),
		b.Reason, expires, b.IsSitewide, now)
	if err != nil {
		if treepg.IsFKViolation(err) {
			// Unknown user UUIDs degrade to display names.
			_, err = s.pool.Exec(ctx, `
				INSERT INTO blocks (id, target_user_name, target_ip, target_ip_raw,
				                    blocked_by_name, reason, expires_at, is_sitewide, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, treepg.UUIDVal(b.ID), treepg.TextVal(b.TargetUserID),
				targetIP, treepg.TextVal(targetIPRaw), treepg.TextVal(b.BlockedBy),
				b.Reason, expires, b.IsSitewide, now)
		}
		return err
	}
	return nil
}

// IsBlocked reports whether a user or IP has a live block.
func (s *PostgresGovernanceStore) IsBlocked(userID, ip string) bool {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var one int
	err := s.pool.QueryRow(ctx, `
		SELECT 1 FROM blocks
		WHERE (expires_at IS NULL OR expires_at > NOW())
		  AND ((target_user_id = $1 AND $1 IS NOT NULL)
		    OR target_user_name = $2
		    OR (target_ip = $3 AND $3 IS NOT NULL)
		    OR target_ip_raw = $4)
		LIMIT 1
	`, treepg.UUIDVal(userID), userID, treepg.InetVal(ip), ip).Scan(&one)
	return err == nil
}

func scanBlock(row pgx.Row) (*Block, error) {
	var (
		id            pgtype.UUID
		targetUID     pgtype.UUID
		targetUName   pgtype.Text
		targetIP      *net.IPNet
		targetIPRaw   pgtype.Text
		blockedByID   pgtype.UUID
		blockedByName pgtype.Text
		reason        string
		expires       pgtype.Timestamptz
		sitewide      bool
		created       time.Time
	)
	if err := row.Scan(&id, &targetUID, &targetUName, &targetIP, &targetIPRaw,
		&blockedByID, &blockedByName, &reason, &expires, &sitewide, &created); err != nil {
		return nil, err
	}
	targetUser := treepg.TextOrEmpty(targetUName)
	if targetUser == "" {
		targetUser = treepg.UUIDString(targetUID)
	}
	targetAddr := treepg.InetString(targetIP)
	if targetAddr == "" {
		targetAddr = treepg.TextOrEmpty(targetIPRaw)
	}
	blockedBy := treepg.TextOrEmpty(blockedByName)
	if blockedBy == "" {
		blockedBy = treepg.UUIDString(blockedByID)
	}
	b := &Block{
		ID:           treepg.UUIDString(id),
		TargetUserID: targetUser,
		TargetIP:     targetAddr,
		BlockedBy:    blockedBy,
		Reason:       reason,
		IsSitewide:   sitewide,
		CreatedAt:    created,
	}
	if expires.Valid {
		t := expires.Time
		b.ExpiresAt = &t
	}
	return b, nil
}

const blockSelectColumns = `
	SELECT id, target_user_id, target_user_name, target_ip, target_ip_raw,
	       blocked_by_id, blocked_by_name, reason, expires_at, is_sitewide, created_at
	FROM blocks
	`

// ListBlocks lists blocks newest-first.
func (s *PostgresGovernanceStore) ListBlocks() []*Block {
	out := []*Block{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, blockSelectColumns+` ORDER BY created_at DESC`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		if b, err := scanBlock(rows); err == nil {
			out = append(out, b)
		}
	}
	return out
}

// AddOversight appends a legal-suppression audit entry.
func (s *PostgresGovernanceStore) AddOversight(o *OversightLog) error {
	if o.ID == "" {
		o.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(o.ID); !ok {
		o.ID = treepg.NewUUID()
	}
	now := time.Now().UTC()
	o.CreatedAt = now
	actorID := treepg.UUIDVal(o.ActorID)
	actorName := ""
	if !actorID.Valid {
		actorName = o.ActorID
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO oversight_logs (id, actor_id, actor_name, action, target, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, treepg.UUIDVal(o.ID), actorID, treepg.TextVal(actorName),
		o.Action, o.Target, o.Details, now)
	if err != nil {
		if treepg.IsFKViolation(err) && actorID.Valid {
			_, err = s.pool.Exec(ctx, `
				INSERT INTO oversight_logs (id, actor_name, action, target, details, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, treepg.UUIDVal(o.ID), o.ActorID, o.Action, o.Target, o.Details, now)
		}
		return err
	}
	return nil
}

// ListOversight lists audit entries newest-first.
func (s *PostgresGovernanceStore) ListOversight() []OversightLog {
	out := []OversightLog{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT id, actor_id, actor_name, action, target, details, created_at
		FROM oversight_logs ORDER BY created_at DESC
	`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id        pgtype.UUID
			actorID   pgtype.UUID
			actorName pgtype.Text
			action    string
			target    string
			details   string
			created   time.Time
		)
		if err := rows.Scan(&id, &actorID, &actorName, &action, &target, &details, &created); err != nil {
			continue
		}
		actor := treepg.TextOrEmpty(actorName)
		if actor == "" {
			actor = treepg.UUIDString(actorID)
		}
		out = append(out, OversightLog{
			ID:        treepg.UUIDString(id),
			ActorID:   actor,
			Action:    action,
			Target:    target,
			Details:   details,
			CreatedAt: created,
		})
	}
	return out
}

// AddTakedown records a takedown request (PENDING by default).
func (s *PostgresGovernanceStore) AddTakedown(t *TakedownRequest) error {
	if t.ID == "" {
		t.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(t.ID); !ok {
		t.ID = treepg.NewUUID()
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	if t.Status == "" {
		t.Status = "PENDING"
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO takedown_requests (id, page_title, reason, requester_name, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, treepg.UUIDVal(t.ID), treepg.TruncateRunes(t.PageTitle, 255),
		t.Reason, t.Requester, t.Status, now)
	return err
}

// ListTakedowns lists requests newest-first.
func (s *PostgresGovernanceStore) ListTakedowns() []*TakedownRequest {
	out := []*TakedownRequest{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT id, page_title, reason, requester_name, status, created_at
		FROM takedown_requests ORDER BY created_at DESC
	`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id        pgtype.UUID
			pageTitle string
			reason    string
			requester string
			status    string
			created   time.Time
		)
		if err := rows.Scan(&id, &pageTitle, &reason, &requester, &status, &created); err != nil {
			continue
		}
		out = append(out, &TakedownRequest{
			ID:        treepg.UUIDString(id),
			PageTitle: pageTitle,
			Reason:    reason,
			Requester: requester,
			Status:    status,
			CreatedAt: created,
		})
	}
	return out
}

// UpdateTakedownStatus transitions a request status.
func (s *PostgresGovernanceStore) UpdateTakedownStatus(id, status string) error {
	uid, ok := treepg.ParseUUID(id)
	if !ok {
		return errors.New("takedown not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `UPDATE takedown_requests SET status = $1 WHERE id = $2`,
		status, treepg.UUIDVal(uid.String()))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("takedown not found")
	}
	return nil
}

// RecordRecentChange appends to the live feed ring (cap 100) and fans out.
func (s *PostgresGovernanceStore) RecordRecentChange(desc string) {
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

// RecentChanges returns the live feed window.
func (s *PostgresGovernanceStore) RecentChanges() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.recent))
	copy(out, s.recent)
	return out
}

// Subscribe registers a feed channel.
func (s *PostgresGovernanceStore) Subscribe() chan string {
	ch := make(chan string, 10)
	s.mu.Lock()
	s.subs = append(s.subs, ch)
	s.mu.Unlock()
	return ch
}
