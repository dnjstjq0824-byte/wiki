package wiki

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	treepg "github.com/dnjstjqppk-dev/treewiki/pkg/pg"
)

// PostgresPageStore persists pages, revisions, links, and categories in
// PostgreSQL. It implements PageStorage with the same observable behavior as
// the in-memory PageStore:
//
//   - Page content/summary/author live on the current revision row (joined on
//     read), matching the revisions + revision_text schema.
//   - Authors that are not user UUIDs (usernames, masked IPs) are stored in
//     author_name/author_ip and resolved back to the same display string.
//   - Backlinks and categories are keyed by page UUID, so Move needs no
//     index rewrites.
type PostgresPageStore struct {
	pool *pgxpool.Pool
}

var _ PageStorage = (*PostgresPageStore)(nil)

// NewPostgresPageStore creates a PostgreSQL-backed page store.
func NewPostgresPageStore(pool *pgxpool.Pool) *PostgresPageStore {
	return &PostgresPageStore{pool: pool}
}

// parsePageKey splits a "namespace:Title" storage key.
func parsePageKey(key string) (Namespace, string, bool) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", false
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", false
	}
	return Namespace(n), parts[1], true
}

const pageSelectColumns = `
	SELECT p.id, p.namespace_id, p.title, p.is_redirect, p.redirect_target_title,
	       p.current_revision_id, p.protection_level, p.is_deleted, p.deleted_at,
	       p.created_at, p.updated_at,
	       r.summary, r.author_name, r.author_ip, u.username, rt.raw_wikitext
	FROM pages p
	LEFT JOIN revisions r ON r.id = p.current_revision_id
	LEFT JOIN revision_text rt ON rt.revision_id = r.id
	LEFT JOIN users u ON u.id = r.author_id
	`

func scanPage(row pgx.Row) (*Page, error) {
	var (
		id        pgtype.UUID
		ns        int
		title     string
		isRedir   bool
		redirTgt  pgtype.Text
		curRev    pgtype.UUID
		prot      string
		isDeleted bool
		deletedAt pgtype.Timestamptz
		createdAt time.Time
		updatedAt time.Time
		summary   pgtype.Text
		authorNm  pgtype.Text
		authorIP  *net.IPNet
		username  pgtype.Text
		content   pgtype.Text
	)
	if err := row.Scan(&id, &ns, &title, &isRedir, &redirTgt, &curRev, &prot,
		&isDeleted, &deletedAt, &createdAt, &updatedAt,
		&summary, &authorNm, &authorIP, &username, &content); err != nil {
		return nil, err
	}
	p := &Page{
		ID:              treepg.UUIDString(id),
		Namespace:       Namespace(ns),
		Title:           title,
		Content:         treepg.TextOrEmpty(content),
		Summary:         treepg.TextOrEmpty(summary),
		AuthorID:        treepg.DisplayAuthor(treepg.TextOrEmpty(username), treepg.TextOrEmpty(authorNm), treepg.InetString(authorIP), ""),
		ProtectionLevel: prot,
		IsRedirect:      isRedir,
		RedirectTarget:  treepg.TextOrEmpty(redirTgt),
		IsDeleted:       isDeleted,
		CurrentRevID:    treepg.UUIDString(curRev),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		p.DeletedAt = &t
	}
	return p, nil
}

// Create inserts a page row. Content arrives with the first revision (the
// service always writes one right after Create).
func (s *PostgresPageStore) Create(page *Page) error {
	if len([]rune(page.Title)) > 255 {
		return errors.New("title too long (max 255 characters)")
	}
	id := page.ID
	if _, ok := treepg.ParseUUID(id); !ok {
		// PostgreSQL IDs are UUIDs; generate one when the caller did not
		// supply a usable ID.
		id = treepg.NewUUID()
	}
	now := time.Now().UTC()
	prot := page.ProtectionLevel
	if prot == "" {
		if info, ok := GetNamespaceInfo(page.Namespace); ok {
			prot = info.DefaultProtection
		} else {
			prot = "all"
		}
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO pages (id, namespace_id, title, slug, is_redirect, redirect_target_title,
		                   protection_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, treepg.UUIDVal(id), int(page.Namespace), page.Title, Slugify(page.Title),
		page.IsRedirect, treepg.TextVal(page.RedirectTarget), prot, now, now)
	if err != nil {
		if treepg.IsUniqueViolation(err) {
			return ErrPageExists
		}
		if treepg.IsFKViolation(err) {
			return errors.New("unknown namespace")
		}
		if treepg.IsCheckViolation(err) {
			return errors.New("invalid protection level")
		}
		return err
	}
	page.ID = id
	page.CreatedAt = now
	page.UpdatedAt = now
	page.ProtectionLevel = prot
	return nil
}

// Get loads a page with its current revision content (deleted pages included;
// the service maps them to ErrPageDeleted).
func (s *PostgresPageStore) Get(ns Namespace, title string) (*Page, error) {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	p, err := scanPage(s.pool.QueryRow(ctx, pageSelectColumns+`
		WHERE p.namespace_id = $1 AND p.title = $2
	`, int(ns), title))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	return p, nil
}

// Update persists page-level flags. Content/summary/author are revision
// scoped and intentionally ignored here.
func (s *PostgresPageStore) Update(page *Page) error {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `
		UPDATE pages
		SET is_redirect = $1, redirect_target_title = $2, protection_level = $3,
		    current_revision_id = $4, updated_at = NOW()
		WHERE namespace_id = $5 AND title = $6
	`, page.IsRedirect, treepg.TextVal(page.RedirectTarget), page.ProtectionLevel,
		treepg.UUIDVal(page.CurrentRevID), int(page.Namespace), page.Title)
	if err != nil {
		if treepg.IsCheckViolation(err) {
			return errors.New("invalid protection level")
		}
		if treepg.IsFKViolation(err) {
			return errors.New("unknown current revision")
		}
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrPageNotFound
	}
	return nil
}

// Delete soft-deletes a page.
func (s *PostgresPageStore) Delete(ns Namespace, title string) error {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `
		UPDATE pages SET is_deleted = TRUE, deleted_at = NOW(), updated_at = NOW()
		WHERE namespace_id = $1 AND title = $2
	`, int(ns), title)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrPageNotFound
	}
	return nil
}

// Undelete restores a soft-deleted page.
func (s *PostgresPageStore) Undelete(ns Namespace, title string) error {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `
		UPDATE pages SET is_deleted = FALSE, deleted_at = NULL, updated_at = NOW()
		WHERE namespace_id = $1 AND title = $2
	`, int(ns), title)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrPageNotFound
	}
	return nil
}

// Move renames a page. Links and categories reference the page UUID, so no
// index rewrites are needed (unlike the memory store key rewrite).
func (s *PostgresPageStore) Move(oldNS Namespace, oldTitle string, newNS Namespace, newTitle string) error {
	if len([]rune(newTitle)) > 255 {
		return errors.New("title too long (max 255 characters)")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var id pgtype.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM pages WHERE namespace_id = $1 AND title = $2
	`, int(oldNS), oldTitle).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPageNotFound
		}
		return err
	}
	var exists bool
	err = s.pool.QueryRow(ctx, `
		SELECT TRUE FROM pages WHERE namespace_id = $1 AND title = $2
	`, int(newNS), newTitle).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if exists {
		return ErrPageExists
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE pages
		SET namespace_id = $1, title = $2, slug = $3, updated_at = NOW()
		WHERE id = $4
	`, int(newNS), newTitle, Slugify(newTitle), id)
	if err != nil {
		if treepg.IsUniqueViolation(err) {
			return ErrPageExists
		}
		if treepg.IsFKViolation(err) {
			return errors.New("unknown namespace")
		}
		return err
	}
	return nil
}

// AddRevision appends an immutable revision with its text, points the page at
// it, and refreshes the full-text vector atomically.
func (s *PostgresPageStore) AddRevision(rev *Revision) error {
	pageID, ok := treepg.ParseUUID(rev.PageID)
	if !ok {
		return errors.New("unknown page")
	}
	if rev.ID == "" {
		rev.ID = treepg.NewUUID()
	}
	if _, ok := treepg.ParseUUID(rev.ID); !ok {
		rev.ID = treepg.NewUUID()
	}
	if rev.CreatedAt.IsZero() {
		rev.CreatedAt = time.Now().UTC()
	}
	var parent pgtype.UUID
	if rev.ParentRevisionID != "" {
		p, ok := treepg.ParseUUID(rev.ParentRevisionID)
		if !ok {
			return errors.New("invalid parent revision")
		}
		parent = pgtype.UUID{Bytes: p, Valid: true}
	}

	authorID := treepg.UUIDVal(rev.AuthorID)
	var authorIP *net.IPNet
	authorName := ""
	if !authorID.Valid {
		if ip := treepg.InetVal(rev.AuthorID); ip != nil {
			authorIP = ip
		} else {
			authorName = rev.AuthorID
		}
	}
	// Edit summaries are capped by the schema (MediaWiki truncates too).
	summary := treepg.TruncateRunes(rev.Summary, 500)
	sum := sha256.Sum256([]byte(rev.Content))
	sha := hex.EncodeToString(sum[:])
	byteLen := len(rev.Content)
	byteDelta := byteLen

	ctx, cancel := treepg.Ctx()
	defer cancel()
	if parent.Valid {
		var parentText pgtype.Text
		err := s.pool.QueryRow(ctx, `
			SELECT raw_wikitext FROM revision_text WHERE revision_id = $1
		`, parent).Scan(&parentText)
		if err == nil && parentText.Valid {
			byteDelta = byteLen - len(parentText.String)
		}
	}

	insert := func(tx pgx.Tx, aid pgtype.UUID, aname string) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO revisions (id, page_id, parent_revision_id, author_id, author_ip,
			                       author_name, summary, byte_length, byte_delta,
			                       sha256_hash, is_minor, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, treepg.UUIDVal(rev.ID), treepg.UUIDVal(pageID.String()), parent, aid, authorIP,
			treepg.TextVal(aname), summary, byteLen, byteDelta, sha, rev.IsMinor, rev.CreatedAt)
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := insert(tx, authorID, authorName); err != nil {
		if treepg.IsFKViolation(err) && authorID.Valid {
			// The author UUID is not a registered user; keep it as a
			// display name instead of failing the edit.
			if retryErr := insert(tx, pgtype.UUID{}, rev.AuthorID); retryErr != nil {
				return retryErr
			}
		} else if treepg.IsFKViolation(err) {
			return errors.New("page not found")
		} else {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO revision_text (revision_id, raw_wikitext) VALUES ($1, $2)
	`, treepg.UUIDVal(rev.ID), rev.Content); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pages
		SET current_revision_id = $1, updated_at = NOW(),
		    search_vector = to_tsvector('simple', title || ' ' || $2)
		WHERE id = $3
	`, treepg.UUIDVal(rev.ID), rev.Content, treepg.UUIDVal(pageID.String())); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

const revisionSelectColumns = `
	SELECT r.id, r.summary, r.author_name, r.author_ip, u.username,
	       rt.raw_wikitext, r.is_minor, r.parent_revision_id, r.created_at
	FROM revisions r
	LEFT JOIN revision_text rt ON rt.revision_id = r.id
	LEFT JOIN users u ON u.id = r.author_id
	`

func scanRevision(pageID string, row pgx.Row) (*Revision, error) {
	var (
		id       pgtype.UUID
		summary  string
		authName pgtype.Text
		authIP   *net.IPNet
		username pgtype.Text
		content  pgtype.Text
		isMinor  bool
		parent   pgtype.UUID
		created  time.Time
	)
	if err := row.Scan(&id, &summary, &authName, &authIP, &username,
		&content, &isMinor, &parent, &created); err != nil {
		return nil, err
	}
	return &Revision{
		ID:               treepg.UUIDString(id),
		PageID:           pageID,
		Content:          treepg.TextOrEmpty(content),
		Summary:          summary,
		AuthorID:         treepg.DisplayAuthor(treepg.TextOrEmpty(username), treepg.TextOrEmpty(authName), treepg.InetString(authIP), ""),
		IsMinor:          isMinor,
		ParentRevisionID: treepg.UUIDString(parent),
		CreatedAt:        created,
	}, nil
}

// GetRevisions returns the page history oldest-first (rollback relies on it).
func (s *PostgresPageStore) GetRevisions(pageID string) []*Revision {
	out := []*Revision{}
	pageUUID, ok := treepg.ParseUUID(pageID)
	if !ok {
		return out
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, revisionSelectColumns+`
		WHERE r.page_id = $1 ORDER BY r.created_at ASC, r.id ASC
	`, treepg.UUIDVal(pageUUID.String()))
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		if rev, err := scanRevision(pageID, rows); err == nil {
			out = append(out, rev)
		}
	}
	return out
}

// GetRevision loads one revision by ID.
func (s *PostgresPageStore) GetRevision(pageID, revID string) (*Revision, error) {
	pageUUID, ok := treepg.ParseUUID(pageID)
	if !ok {
		return nil, errors.New("revision not found")
	}
	revUUID, ok := treepg.ParseUUID(revID)
	if !ok {
		return nil, errors.New("revision not found")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rev, err := scanRevision(pageID, s.pool.QueryRow(ctx, revisionSelectColumns+`
		WHERE r.page_id = $1 AND r.id = $2
	`, treepg.UUIDVal(pageUUID.String()), treepg.UUIDVal(revUUID.String())))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("revision not found")
		}
		return nil, err
	}
	return rev, nil
}

func (s *PostgresPageStore) pageUUID(ctx context.Context, ns Namespace, title string) (pgtype.UUID, bool) {
	var id pgtype.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM pages WHERE namespace_id = $1 AND title = $2
	`, int(ns), title).Scan(&id)
	if err != nil || !id.Valid {
		return pgtype.UUID{}, false
	}
	return id, true
}

// RebuildBacklinks replaces the outgoing link set of one page.
func (s *PostgresPageStore) RebuildBacklinks(sourceKey string, targets []string) error {
	ns, title, ok := parsePageKey(sourceKey)
	if !ok {
		return nil
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	fromID, ok := s.pageUUID(ctx, ns, title)
	if !ok {
		return nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM page_links WHERE from_page_id = $1`, fromID); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, t := range targets {
		toNS, toTitle, ok := parsePageKey(t)
		if !ok || len([]rune(toTitle)) > 255 {
			continue
		}
		if _, known := GetNamespaceInfo(toNS); !known {
			continue
		}
		dedupe := PageKey(toNS, toTitle)
		if seen[dedupe] {
			continue
		}
		seen[dedupe] = true
		if _, err := tx.Exec(ctx, `
			INSERT INTO page_links (from_page_id, to_namespace_id, to_page_title)
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
		`, fromID, int(toNS), toTitle); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// GetBacklinks returns storage keys of pages linking to the target.
func (s *PostgresPageStore) GetBacklinks(targetKey string) []string {
	out := []string{}
	ns, title, ok := parsePageKey(targetKey)
	if !ok {
		return out
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT p.namespace_id, p.title
		FROM page_links l JOIN pages p ON p.id = l.from_page_id
		WHERE l.to_namespace_id = $1 AND l.to_page_title = $2
		ORDER BY p.title
	`, int(ns), title)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var fromNS int
		var fromTitle string
		if err := rows.Scan(&fromNS, &fromTitle); err == nil {
			out = append(out, PageKey(Namespace(fromNS), fromTitle))
		}
	}
	return out
}

// ensureCategoryTx resolves a category name to its UUID, creating the row.
// ok=false means the name cannot be stored (too long or slug collision) and
// the caller should skip it.
func ensureCategoryTx(ctx context.Context, tx pgx.Tx, name string) (pgtype.UUID, bool, error) {
	if name == "" || len([]rune(name)) > 128 {
		return pgtype.UUID{}, false, nil
	}
	var id pgtype.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM categories WHERE name = $1`, name).Scan(&id)
	if err == nil && id.Valid {
		return id, true, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, false, err
	}
	slug := Slugify(name)
	if slug == "" || len([]rune(slug)) > 128 {
		return pgtype.UUID{}, false, nil
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO categories (name, slug) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING
	`, name, slug)
	if err != nil {
		if treepg.IsUniqueViolation(err) {
			return pgtype.UUID{}, false, nil
		}
		return pgtype.UUID{}, false, err
	}
	err = tx.QueryRow(ctx, `SELECT id FROM categories WHERE name = $1`, name).Scan(&id)
	if err != nil || !id.Valid {
		return pgtype.UUID{}, false, nil
	}
	return id, true, nil
}

// RebuildCategories replaces the category membership of one page.
func (s *PostgresPageStore) RebuildCategories(sourceKey string, cats []string) error {
	ns, title, ok := parsePageKey(sourceKey)
	if !ok {
		return nil
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	pageID, ok := s.pageUUID(ctx, ns, title)
	if !ok {
		return nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM page_categories WHERE page_id = $1`, pageID); err != nil {
		return err
	}
	for _, c := range cats {
		catID, ok, err := ensureCategoryTx(ctx, tx, c)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO page_categories (page_id, category_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, pageID, catID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// GetCategoryMembers returns storage keys of pages in a category.
func (s *PostgresPageStore) GetCategoryMembers(category string) []string {
	out := []string{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT p.namespace_id, p.title
		FROM page_categories pc
		JOIN categories c ON c.id = pc.category_id
		JOIN pages p ON p.id = pc.page_id
		WHERE c.name = $1
		ORDER BY p.title
	`, category)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var ns int
		var title string
		if err := rows.Scan(&ns, &title); err == nil {
			out = append(out, PageKey(Namespace(ns), title))
		}
	}
	return out
}

// ListByNamespace returns storage keys of live pages in a namespace.
func (s *PostgresPageStore) ListByNamespace(ns Namespace) []string {
	out := []string{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT namespace_id, title FROM pages
		WHERE namespace_id = $1 AND is_deleted = FALSE
		ORDER BY title
	`, int(ns))
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var n int
		var title string
		if err := rows.Scan(&n, &title); err == nil {
			out = append(out, PageKey(Namespace(n), title))
		}
	}
	return out
}

// CountAll counts live pages across namespaces.
func (s *PostgresPageStore) CountAll() int {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var n int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM pages WHERE is_deleted = FALSE`).Scan(&n); err != nil {
		return 0
	}
	return n
}

// CountByNamespace counts live pages in one namespace.
func (s *PostgresPageStore) CountByNamespace(ns Namespace) int {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var n int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pages WHERE namespace_id = $1 AND is_deleted = FALSE
	`, int(ns)).Scan(&n); err != nil {
		return 0
	}
	return n
}

// Random returns a random live page, optionally filtered by namespace.
func (s *PostgresPageStore) Random(ns *Namespace) (*Page, error) {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var row pgx.Row
	if ns == nil {
		row = s.pool.QueryRow(ctx, `
			SELECT namespace_id, title FROM pages
			WHERE is_deleted = FALSE ORDER BY RANDOM() LIMIT 1
		`)
	} else {
		row = s.pool.QueryRow(ctx, `
			SELECT namespace_id, title FROM pages
			WHERE namespace_id = $1 AND is_deleted = FALSE ORDER BY RANDOM() LIMIT 1
		`, int(*ns))
	}
	var n int
	var title string
	if err := row.Scan(&n, &title); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	return s.Get(Namespace(n), title)
}
