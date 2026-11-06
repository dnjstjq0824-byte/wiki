package watchlist

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	treepg "github.com/dnjstjqppk-dev/treewiki/pkg/pg"
)

// PostgresWatchlistStore persists per-user page watches in PostgreSQL.
type PostgresWatchlistStore struct {
	pool *pgxpool.Pool
}

var _ WatchlistStorage = (*PostgresWatchlistStore)(nil)

// NewPostgresWatchlistStore creates a PostgreSQL-backed watchlist store.
func NewPostgresWatchlistStore(pool *pgxpool.Pool) *PostgresWatchlistStore {
	return &PostgresWatchlistStore{pool: pool}
}

func parseWatchPageKey(key string) (int, string, bool) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", false
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", false
	}
	return n, parts[1], true
}

// Add watches a page (idempotent; refreshes the email flag).
func (s *PostgresWatchlistStore) Add(userID, pageKey string, notifyEmail bool) error {
	userUUID, ok := treepg.ParseUUID(userID)
	if !ok {
		return errors.New("unknown user")
	}
	ns, title, ok := parseWatchPageKey(pageKey)
	if !ok {
		return errors.New("page not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var pageID pgtype.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM pages WHERE namespace_id = $1 AND title = $2
	`, ns, title).Scan(&pageID)
	if err != nil || !pageID.Valid {
		return errors.New("page not found")
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO watchlists (user_id, page_id, notify_email)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, page_id)
		DO UPDATE SET notify_email = EXCLUDED.notify_email, created_at = NOW()
	`, treepg.UUIDVal(userUUID.String()), pageID, notifyEmail)
	if err != nil {
		if treepg.IsFKViolation(err) {
			return errors.New("unknown user")
		}
		return err
	}
	return nil
}

// Remove unwatches a page (idempotent).
func (s *PostgresWatchlistStore) Remove(userID, pageKey string) error {
	userUUID, ok := treepg.ParseUUID(userID)
	if !ok {
		return errors.New("unknown user")
	}
	ns, title, ok := parseWatchPageKey(pageKey)
	if !ok {
		return errors.New("page not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var pageID pgtype.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM pages WHERE namespace_id = $1 AND title = $2
	`, ns, title).Scan(&pageID)
	if err != nil || !pageID.Valid {
		return errors.New("page not found")
	}
	_, err = s.pool.Exec(ctx, `
		DELETE FROM watchlists WHERE user_id = $1 AND page_id = $2
	`, treepg.UUIDVal(userUUID.String()), pageID)
	return err
}

// List returns a user's watched pages newest-first.
func (s *PostgresWatchlistStore) List(userID string) []*WatchEntry {
	out := []*WatchEntry{}
	userUUID, ok := treepg.ParseUUID(userID)
	if !ok {
		return out
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT p.namespace_id, p.title, w.notify_email, w.created_at
		FROM watchlists w JOIN pages p ON p.id = w.page_id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC
	`, treepg.UUIDVal(userUUID.String()))
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var (
			ns      int
			title   string
			notify  bool
			created time.Time
		)
		if err := rows.Scan(&ns, &title, &notify, &created); err != nil {
			continue
		}
		out = append(out, &WatchEntry{
			UserID:      userUUID.String(),
			PageKey:     strconv.Itoa(ns) + ":" + title,
			NotifyEmail: notify,
			CreatedAt:   created,
		})
	}
	return out
}

// IsWatching reports whether a user watches a page.
func (s *PostgresWatchlistStore) IsWatching(userID, pageKey string) bool {
	userUUID, ok := treepg.ParseUUID(userID)
	if !ok {
		return false
	}
	ns, title, ok := parseWatchPageKey(pageKey)
	if !ok {
		return false
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var one int
	err := s.pool.QueryRow(ctx, `
		SELECT 1 FROM watchlists w JOIN pages p ON p.id = w.page_id
		WHERE w.user_id = $1 AND p.namespace_id = $2 AND p.title = $3
	`, treepg.UUIDVal(userUUID.String()), ns, title).Scan(&one)
	return err == nil
}

// Watchers lists user IDs watching a page.
func (s *PostgresWatchlistStore) Watchers(pageKey string) []string {
	out := []string{}
	ns, title, ok := parseWatchPageKey(pageKey)
	if !ok {
		return out
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT w.user_id FROM watchlists w JOIN pages p ON p.id = w.page_id
		WHERE p.namespace_id = $1 AND p.title = $2
	`, ns, title)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var uid pgtype.UUID
		if err := rows.Scan(&uid); err == nil && uid.Valid {
			out = append(out, treepg.UUIDString(uid))
		}
	}
	return out
}
