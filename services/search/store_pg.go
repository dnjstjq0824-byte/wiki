package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	treepg "github.com/dnjstjqppk-dev/treewiki/pkg/pg"
)

// PostgresSearchBackend runs full-text search as live SQL over the pages
// table (tsvector + trigram), so no separate index process is needed. The
// wiki write path refreshes search_vector on every revision; Index re-applies
// it for out-of-band documents.
type PostgresSearchBackend struct {
	pool *pgxpool.Pool
}

var _ SearchBackend = (*PostgresSearchBackend)(nil)

// NewPostgresSearchBackend creates a SQL-backed search backend.
func NewPostgresSearchBackend(pool *pgxpool.Pool) *PostgresSearchBackend {
	return &PostgresSearchBackend{pool: pool}
}

// Index refreshes the full-text vector of one page document.
func (s *PostgresSearchBackend) Index(doc *SearchDocument) {
	id, ok := treepg.ParseUUID(doc.ID)
	if !ok {
		return
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, _ = s.pool.Exec(ctx, `
		UPDATE pages
		SET search_vector = to_tsvector('simple', title || ' ' || $1)
		WHERE id = $2
	`, doc.Content, treepg.UUIDVal(id.String()))
}

// Remove is a no-op: deleted pages are excluded by the live queries.
func (s *PostgresSearchBackend) Remove(_ string) {}

// Search runs a plain ranked query (namespace < 0 means all).
func (s *PostgresSearchBackend) Search(query string, namespace int) []SearchResult {
	return s.SearchFaceted(query, namespace, "", "", "")
}

// Autocomplete suggests titles by prefix.
func (s *PostgresSearchBackend) Autocomplete(prefix string, limit int) []string {
	out := []string{}
	p := strings.TrimSpace(prefix)
	if p == "" {
		return out
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT title FROM pages
		WHERE is_deleted = FALSE AND title ILIKE $1 ESCAPE '\'
		ORDER BY title LIMIT $2
	`, escapeLike(p)+"%", limit)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err == nil {
			out = append(out, title)
		}
	}
	return out
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func parseTimeBound(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// SearchFaceted runs a ranked full-text query with namespace, category, and
// date-range filters (plus title boost, mirroring the memory backend).
func (s *PostgresSearchBackend) SearchFaceted(query string, namespace int, category string, from, to string) []SearchResult {
	out := []SearchResult{}
	q := strings.TrimSpace(query)
	ctx, cancel := treepg.Ctx()
	defer cancel()

	where := []string{"p.is_deleted = FALSE"}
	args := []any{}
	// Build predicates with positional parameters.
	addArg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}
	if namespace >= 0 {
		where = append(where, "p.namespace_id = "+addArg(namespace))
	}
	if strings.TrimSpace(category) != "" {
		where = append(where, `EXISTS (
			SELECT 1 FROM page_categories pc JOIN categories c ON c.id = pc.category_id
			WHERE pc.page_id = p.id AND c.name = `+addArg(strings.TrimSpace(category))+`)`)
	}
	if t, ok := parseTimeBound(from); ok {
		where = append(where, "p.created_at >= "+addArg(t))
	}
	if t, ok := parseTimeBound(to); ok {
		where = append(where, "p.created_at <= "+addArg(t))
	}

	selectRank := "0 AS rnk"
	orderBy := "p.updated_at DESC"
	if q != "" {
		tsQuery := addArg(q)
		where = append(where, "(p.search_vector @@ plainto_tsquery('simple', "+tsQuery+")"+
			" OR p.title ILIKE '%' || "+addArg(escapeLike(q))+" || '%' ESCAPE '\\')")
		selectRank = "ts_rank(p.search_vector, plainto_tsquery('simple', " + tsQuery + ")) AS rnk"
		orderBy = "title_hit DESC, rnk DESC, p.updated_at DESC"
	}

	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.namespace_id, p.title, rt.raw_wikitext,
		       (p.title ILIKE '%' || $`+fmt.Sprintf("%d", len(args)+1)+` || '%' ESCAPE '\') AS title_hit,
		       `+selectRank+`
		FROM pages p
		LEFT JOIN revisions r ON r.id = p.current_revision_id
		LEFT JOIN revision_text rt ON rt.revision_id = r.id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY `+orderBy+`
		LIMIT 50
	`, append(args, escapeLike(q))...)
	if err != nil {
		return out
	}
	defer rows.Close()
	lowerQ := strings.ToLower(q)
	for rows.Next() {
		var (
			id      pgtype.UUID
			ns      int
			title   string
			content pgtype.Text
			hit     bool
			rank    float64
		)
		if err := rows.Scan(&id, &ns, &title, &content, &hit, &rank); err != nil {
			continue
		}
		body := treepg.TextOrEmpty(content)
		score := rank
		if q == "" {
			score = 1
		} else if hit || strings.Contains(strings.ToLower(title), lowerQ) {
			score += 5
		}
		doc := SearchDocument{
			ID:        treepg.UUIDString(id),
			Namespace: ns,
			Title:     title,
			Content:   body,
		}
		out = append(out, SearchResult{Document: doc, Score: score, Snippet: makeSnippet(body, lowerQ)})
	}
	return out
}

// PostgresMediaStore persists media metadata in PostgreSQL. Assets resolve by
// filename ("media-<name>" IDs included) or UUID.
type PostgresMediaStore struct {
	pool *pgxpool.Pool
}

var _ MediaStorage = (*PostgresMediaStore)(nil)

// NewPostgresMediaStore creates a PostgreSQL-backed media store.
func NewPostgresMediaStore(pool *pgxpool.Pool) *PostgresMediaStore {
	return &PostgresMediaStore{pool: pool}
}

// Add upserts an asset by filename.
func (s *PostgresMediaStore) Add(a *MediaAsset) error {
	if strings.TrimSpace(a.Filename) == "" {
		return fmt.Errorf("filename required")
	}
	if len([]rune(a.Filename)) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}
	if strings.TrimSpace(a.SHA256) == "" {
		return fmt.Errorf("checksum required")
	}
	mime := a.MimeType
	if mime == "" {
		mime = "application/octet-stream"
	}
	key := a.S3Key
	if key == "" {
		key = a.Filename
	}
	uploaderID := treepg.UUIDVal(a.UploaderID)
	uploaderName := ""
	if !uploaderID.Valid {
		uploaderName = a.UploaderID
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var (
		id      pgtype.UUID
		created time.Time
	)
	err := s.pool.QueryRow(ctx, `
		INSERT INTO media_assets (filename, original_name, mime_type, file_size_bytes,
		                          s3_key, sha256_hash, license_type, uploader_id, uploader_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (filename) DO UPDATE SET
			original_name = EXCLUDED.original_name, mime_type = EXCLUDED.mime_type,
			file_size_bytes = EXCLUDED.file_size_bytes, s3_key = EXCLUDED.s3_key,
			sha256_hash = EXCLUDED.sha256_hash, license_type = EXCLUDED.license_type,
			uploader_id = EXCLUDED.uploader_id, uploader_name = EXCLUDED.uploader_name
		RETURNING id, created_at
	`, a.Filename, a.OriginalName, mime, a.FileSize, key, a.SHA256,
		a.LicenseType, uploaderID, treepg.TextVal(uploaderName)).Scan(&id, &created)
	if err != nil {
		if treepg.IsFKViolation(err) && uploaderID.Valid {
			err = s.pool.QueryRow(ctx, `
				INSERT INTO media_assets (filename, original_name, mime_type, file_size_bytes,
				                          s3_key, sha256_hash, license_type, uploader_name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (filename) DO UPDATE SET
					original_name = EXCLUDED.original_name, mime_type = EXCLUDED.mime_type,
					file_size_bytes = EXCLUDED.file_size_bytes, s3_key = EXCLUDED.s3_key,
					sha256_hash = EXCLUDED.sha256_hash, license_type = EXCLUDED.license_type,
					uploader_id = NULL, uploader_name = EXCLUDED.uploader_name
				RETURNING id, created_at
			`, a.Filename, a.OriginalName, mime, a.FileSize, key, a.SHA256,
				a.LicenseType, a.UploaderID).Scan(&id, &created)
		}
		if err != nil {
			return err
		}
	}
	a.ID = treepg.UUIDString(id)
	a.CreatedAt = created.Format(time.RFC3339)
	return nil
}

const mediaSelectColumns = `
	SELECT m.id, m.filename, m.original_name, m.mime_type, m.file_size_bytes,
	       m.s3_key, m.sha256_hash, m.license_type, m.uploader_id, m.uploader_name,
	       u.username, m.created_at
	FROM media_assets m
	LEFT JOIN users u ON u.id = m.uploader_id
	`

func scanMedia(row interface{ Scan(...any) error }) (*MediaAsset, bool) {
	var (
		id       pgtype.UUID
		filename string
		original string
		mime     string
		size     int64
		key      string
		sha      string
		license  string
		uplID    pgtype.UUID
		uplName  pgtype.Text
		username pgtype.Text
		created  time.Time
	)
	if err := row.Scan(&id, &filename, &original, &mime, &size, &key, &sha,
		&license, &uplID, &uplName, &username, &created); err != nil {
		return nil, false
	}
	uploader := treepg.TextOrEmpty(username)
	if uploader == "" {
		uploader = treepg.TextOrEmpty(uplName)
	}
	if uploader == "" {
		uploader = treepg.UUIDString(uplID)
	}
	return &MediaAsset{
		ID:           treepg.UUIDString(id),
		Filename:     filename,
		OriginalName: original,
		MimeType:     mime,
		FileSize:     size,
		S3Key:        key,
		SHA256:       sha,
		LicenseType:  license,
		UploaderID:   uploader,
		CreatedAt:    created.Format(time.RFC3339),
	}, true
}

// Get resolves an asset by filename, "media-<filename>" ID, or UUID.
func (s *PostgresMediaStore) Get(id string) (*MediaAsset, bool) {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	names := []string{id}
	if stripped, ok := strings.CutPrefix(id, "media-"); ok && stripped != "" {
		names = append(names, stripped)
	}
	if a, ok := scanMedia(s.pool.QueryRow(ctx, mediaSelectColumns+` WHERE m.filename = ANY($1)`, names)); ok {
		return a, true
	}
	if uid, ok := treepg.ParseUUID(id); ok {
		if a, ok := scanMedia(s.pool.QueryRow(ctx, mediaSelectColumns+` WHERE m.id = $1`, treepg.UUIDVal(uid.String()))); ok {
			return a, true
		}
	}
	return nil, false
}

// List returns assets newest-first.
func (s *PostgresMediaStore) List() []*MediaAsset {
	out := []*MediaAsset{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, mediaSelectColumns+` ORDER BY m.created_at DESC`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id       pgtype.UUID
			filename string
			original string
			mime     string
			size     int64
			key      string
			sha      string
			license  string
			uplID    pgtype.UUID
			uplName  pgtype.Text
			username pgtype.Text
			created  time.Time
		)
		if err := rows.Scan(&id, &filename, &original, &mime, &size, &key, &sha,
			&license, &uplID, &uplName, &username, &created); err != nil {
			continue
		}
		uploader := treepg.TextOrEmpty(username)
		if uploader == "" {
			uploader = treepg.TextOrEmpty(uplName)
		}
		if uploader == "" {
			uploader = treepg.UUIDString(uplID)
		}
		out = append(out, &MediaAsset{
			ID:           treepg.UUIDString(id),
			Filename:     filename,
			OriginalName: original,
			MimeType:     mime,
			FileSize:     size,
			S3Key:        key,
			SHA256:       sha,
			LicenseType:  license,
			UploaderID:   uploader,
			CreatedAt:    created.Format(time.RFC3339),
		})
	}
	return out
}
