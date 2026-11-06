package topictree

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	treepg "github.com/dnjstjqppk-dev/treewiki/pkg/pg"
)

// PostgresTopicStore persists the topic DAG and learning paths in PostgreSQL.
// External node/path IDs map to the slug columns (memory IDs like "cs" are
// not UUIDs); internal UUID primary keys stay hidden from the API.
type PostgresTopicStore struct {
	pool *pgxpool.Pool
}

var _ TopicStorage = (*PostgresTopicStore)(nil)

// NewPostgresTopicStore creates a PostgreSQL-backed topic store.
func NewPostgresTopicStore(pool *pgxpool.Pool) *PostgresTopicStore {
	return &PostgresTopicStore{pool: pool}
}

const nodeSelectColumns = `
	SELECT n.id, n.slug, n.title, n.description, n.difficulty, n.completion,
	       n.prereq_slugs, n.advanced_slugs, n.related_slugs,
	       n.created_at, n.updated_at
	FROM topic_nodes n
	`

func scanNode(row pgx.Row) (id pgtype.UUID, n *TopicNode, err error) {
	var (
		slug     string
		title    string
		desc     pgtype.Text
		diff     int
		compl    float64
		prereq   []string
		advanced []string
		related  []string
		created  time.Time
		updated  time.Time
	)
	n = &TopicNode{}
	if err = row.Scan(&id, &slug, &title, &desc, &diff, &compl,
		&prereq, &advanced, &related, &created, &updated); err != nil {
		return pgtype.UUID{}, nil, err
	}
	n.ID = slug
	n.Name = title
	n.Description = treepg.TextOrEmpty(desc)
	n.Difficulty = diff
	n.Completion = compl
	n.Prerequisites = nonNilStrings(prereq)
	n.Advanced = nonNilStrings(advanced)
	n.Related = nonNilStrings(related)
	n.ParentIDs = []string{}
	n.ChildIDs = []string{}
	n.CreatedAt = created
	n.UpdatedAt = updated
	return id, n, nil
}

func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

func (s *PostgresTopicStore) nodeLinks(ctx context.Context, id pgtype.UUID, n *TopicNode) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.slug FROM topic_edges e JOIN topic_nodes p ON p.id = e.source_node_id
		WHERE e.target_node_id = $1 ORDER BY p.slug
	`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var slug string
			if err := rows.Scan(&slug); err == nil {
				n.ParentIDs = append(n.ParentIDs, slug)
			}
		}
	}
	rows, err = s.pool.Query(ctx, `
		SELECT c.slug FROM topic_edges e JOIN topic_nodes c ON c.id = e.target_node_id
		WHERE e.source_node_id = $1 ORDER BY c.slug
	`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var slug string
			if err := rows.Scan(&slug); err == nil {
				n.ChildIDs = append(n.ChildIDs, slug)
			}
		}
	}
}

func (s *PostgresTopicStore) nodeUUID(ctx context.Context, slug string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := s.pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, slug).Scan(&id)
	if err != nil || !id.Valid {
		return pgtype.UUID{}, ErrNodeNotFound
	}
	return id, nil
}

// CreateNode inserts a node; caller-supplied parent/child IDs become edges
// (unknown nodes are skipped, mirroring memory leniency on dangling IDs).
func (s *PostgresTopicStore) CreateNode(n *TopicNode) error {
	if n.Name == "" {
		return errors.New("name required")
	}
	if n.ID == "" {
		n.ID = treepg.NewUUID()
	}
	if len([]rune(n.ID)) > 128 {
		return errors.New("node id too long (max 128 characters)")
	}
	if len([]rune(n.Name)) > 128 {
		return errors.New("node name too long (max 128 characters)")
	}
	now := time.Now().UTC()
	n.CreatedAt = now
	n.UpdatedAt = now
	if n.ParentIDs == nil {
		n.ParentIDs = []string{}
	}
	if n.ChildIDs == nil {
		n.ChildIDs = []string{}
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id pgtype.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO topic_nodes (slug, title, description, difficulty, completion,
		                         prereq_slugs, advanced_slugs, related_slugs,
		                         created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`, n.ID, n.Name, treepg.TextVal(n.Description), n.Difficulty, n.Completion,
		nonNilStrings(n.Prerequisites), nonNilStrings(n.Advanced), nonNilStrings(n.Related),
		now, now).Scan(&id)
	if err != nil {
		if treepg.IsUniqueViolation(err) {
			return ErrNodeExists
		}
		return err
	}
	for _, parentSlug := range n.ParentIDs {
		if err := insertEdgeTx(ctx, tx, s.pool, parentSlug, n.ID); err != nil {
			return err
		}
	}
	for _, childSlug := range n.ChildIDs {
		if err := insertEdgeTx(ctx, tx, s.pool, n.ID, childSlug); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// GetNode loads a node with edges resolved to external IDs.
func (s *PostgresTopicStore) GetNode(slug string) (*TopicNode, error) {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	id, n, err := scanNode(s.pool.QueryRow(ctx, nodeSelectColumns+` WHERE n.slug = $1`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}
	s.nodeLinks(ctx, id, n)
	return n, nil
}

// UpdateNode replaces scalar fields and, when edge lists are provided,
// resyncs edges to match (nil lists preserve existing edges so partial PUTs
// cannot wipe the DAG).
func (s *PostgresTopicStore) UpdateNode(n *TopicNode) error {
	if len([]rune(n.Name)) > 128 {
		return errors.New("node name too long (max 128 characters)")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	id, err := s.nodeUUID(ctx, n.ID)
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	sets := `title = $1, description = $2, difficulty = $3, completion = $4, updated_at = NOW()`
	args := []any{n.Name, treepg.TextVal(n.Description), n.Difficulty, n.Completion}
	next := 5
	if n.Prerequisites != nil {
		sets += `, prereq_slugs = $` + itoa(next)
		args = append(args, n.Prerequisites)
		next++
	}
	if n.Advanced != nil {
		sets += `, advanced_slugs = $` + itoa(next)
		args = append(args, n.Advanced)
		next++
	}
	if n.Related != nil {
		sets += `, related_slugs = $` + itoa(next)
		args = append(args, n.Related)
		next++
	}
	args = append(args, id)
	if _, err := tx.Exec(ctx, `UPDATE topic_nodes SET `+sets+` WHERE id = $`+itoa(next), args...); err != nil {
		return err
	}
	if n.ParentIDs != nil || n.ChildIDs != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM topic_edges WHERE source_node_id = $1 OR target_node_id = $1`, id); err != nil {
			return err
		}
		for _, parentSlug := range n.ParentIDs {
			if err := insertEdgeTx(ctx, tx, s.pool, parentSlug, n.ID); err != nil {
				return err
			}
		}
		for _, childSlug := range n.ChildIDs {
			if err := insertEdgeTx(ctx, tx, s.pool, n.ID, childSlug); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// DeleteNode removes a node; edges cascade.
func (s *PostgresTopicStore) DeleteNode(slug string) error {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	res, err := s.pool.Exec(ctx, `DELETE FROM topic_nodes WHERE slug = $1`, slug)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNodeNotFound
	}
	return nil
}

// AllNodes lists every node ordered by slug.
func (s *PostgresTopicStore) AllNodes() []*TopicNode {
	out := []*TopicNode{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, nodeSelectColumns+` ORDER BY n.slug`)
	if err != nil {
		return out
	}
	defer rows.Close()
	type pair struct {
		id pgtype.UUID
		n  *TopicNode
	}
	var pairs []pair
	for rows.Next() {
		if id, n, err := scanNode(rows); err == nil {
			pairs = append(pairs, pair{id, n})
		}
	}
	rows.Close()
	for _, p := range pairs {
		s.nodeLinks(ctx, p.id, p.n)
		out = append(out, p.n)
	}
	return out
}

func insertEdgeTx(ctx context.Context, tx pgx.Tx, pool *pgxpool.Pool, parentSlug, childSlug string) error {
	if parentSlug == "" || childSlug == "" || parentSlug == childSlug {
		return nil
	}
	var parentID, childID pgtype.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, parentSlug).Scan(&parentID); err != nil {
		return nil
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, childSlug).Scan(&childID); err != nil {
		return nil
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO topic_edges (source_node_id, target_node_id, relationship_type)
		VALUES ($1, $2, 'subtopic') ON CONFLICT DO NOTHING
	`, parentID, childID)
	return err
}

// AddEdge links parent -> child with cycle detection.
func (s *PostgresTopicStore) AddEdge(parentID, childID string) error {
	if parentID == childID {
		return ErrCycleDetected
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var parentUUID, childUUID pgtype.UUID
	err := s.pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, parentID).Scan(&parentUUID)
	if err != nil || !parentUUID.Valid {
		return ErrNodeNotFound
	}
	err = s.pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, childID).Scan(&childUUID)
	if err != nil || !childUUID.Valid {
		return ErrNodeNotFound
	}
	for _, ancestor := range s.ancestorsOf(ctx, parentUUID) {
		if ancestor == childID {
			return ErrCycleDetected
		}
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO topic_edges (source_node_id, target_node_id, relationship_type)
		VALUES ($1, $2, 'subtopic') ON CONFLICT DO NOTHING
	`, parentUUID, childUUID)
	return err
}

// RemoveEdge unlinks parent -> child.
func (s *PostgresTopicStore) RemoveEdge(parentID, childID string) error {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var parentUUID, childUUID pgtype.UUID
	err := s.pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, parentID).Scan(&parentUUID)
	if err != nil || !parentUUID.Valid {
		return ErrNodeNotFound
	}
	err = s.pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, childID).Scan(&childUUID)
	if err != nil || !childUUID.Valid {
		return ErrNodeNotFound
	}
	_, err = s.pool.Exec(ctx, `
		DELETE FROM topic_edges WHERE source_node_id = $1 AND target_node_id = $2
	`, parentUUID, childUUID)
	return err
}

func (s *PostgresTopicStore) ancestorsOf(ctx context.Context, id pgtype.UUID) []string {
	out := []string{}
	rows, err := s.pool.Query(ctx, `
		WITH RECURSIVE up(id) AS (
			SELECT e.source_node_id FROM topic_edges e WHERE e.target_node_id = $1
			UNION
			SELECT e.source_node_id FROM topic_edges e JOIN up ON up.id = e.target_node_id
		)
		SELECT n.slug FROM up JOIN topic_nodes n ON n.id = up.id ORDER BY n.slug
	`, id)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err == nil {
			out = append(out, slug)
		}
	}
	return out
}

func (s *PostgresTopicStore) descendantsOf(ctx context.Context, id pgtype.UUID) []string {
	out := []string{}
	rows, err := s.pool.Query(ctx, `
		WITH RECURSIVE down(id) AS (
			SELECT e.target_node_id FROM topic_edges e WHERE e.source_node_id = $1
			UNION
			SELECT e.target_node_id FROM topic_edges e JOIN down ON down.id = e.source_node_id
		)
		SELECT n.slug FROM down JOIN topic_nodes n ON n.id = down.id ORDER BY n.slug
	`, id)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err == nil {
			out = append(out, slug)
		}
	}
	return out
}

// GetAncestors returns ancestor slugs (set semantics; ordered by slug).
func (s *PostgresTopicStore) GetAncestors(slug string) []string {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var id pgtype.UUID
	if err := s.pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, slug).Scan(&id); err != nil {
		return []string{}
	}
	return s.ancestorsOf(ctx, id)
}

// GetDescendants returns descendant slugs (set semantics; ordered by slug).
func (s *PostgresTopicStore) GetDescendants(slug string) []string {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var id pgtype.UUID
	if err := s.pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, slug).Scan(&id); err != nil {
		return []string{}
	}
	return s.descendantsOf(ctx, id)
}

// ExportGraph dumps nodes and parent->child edges with external IDs.
func (s *PostgresTopicStore) ExportGraph() GraphExport {
	ge := GraphExport{Nodes: []TopicNode{}, Edges: []TopicEdge{}}
	for _, n := range s.AllNodes() {
		ge.Nodes = append(ge.Nodes, *n)
		for _, c := range n.ChildIDs {
			ge.Edges = append(ge.Edges, TopicEdge{From: n.ID, To: c})
		}
	}
	return ge
}

// CreatePath upserts a learning path (memory overwrites on ID reuse).
func (s *PostgresTopicStore) CreatePath(p *LearningPath) error {
	if p.ID == "" {
		p.ID = treepg.NewUUID()
	}
	if len([]rune(p.ID)) > 128 {
		return errors.New("path id too long (max 128 characters)")
	}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var pathID pgtype.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO learning_paths (slug, name) VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, p.ID, p.Name).Scan(&pathID)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM learning_path_nodes WHERE path_id = $1`, pathID); err != nil {
		return err
	}
	pos := 0
	for _, slug := range p.NodeIDs {
		var nodeID pgtype.UUID
		err := s.pool.QueryRow(ctx, `SELECT id FROM topic_nodes WHERE slug = $1`, slug).Scan(&nodeID)
		if err != nil || !nodeID.Valid {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO learning_path_nodes (path_id, node_id, position)
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
		`, pathID, nodeID, pos); err != nil {
			return err
		}
		pos++
	}
	return tx.Commit(ctx)
}

// GetPath loads a path with node slugs in order.
func (s *PostgresTopicStore) GetPath(slug string) (*LearningPath, error) {
	ctx, cancel := treepg.Ctx()
	defer cancel()
	var (
		id   pgtype.UUID
		name string
	)
	err := s.pool.QueryRow(ctx, `SELECT id, name FROM learning_paths WHERE slug = $1`, slug).Scan(&id, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPathNotFound
		}
		return nil, err
	}
	return &LearningPath{ID: slug, Name: name, NodeIDs: s.pathNodes(ctx, id)}, nil
}

func (s *PostgresTopicStore) pathNodes(ctx context.Context, pathID pgtype.UUID) []string {
	out := []string{}
	rows, err := s.pool.Query(ctx, `
		SELECT n.slug FROM learning_path_nodes l
		JOIN topic_nodes n ON n.id = l.node_id
		WHERE l.path_id = $1 ORDER BY l.position
	`, pathID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err == nil {
			out = append(out, slug)
		}
	}
	return out
}

// AllPaths lists every learning path ordered by slug.
func (s *PostgresTopicStore) AllPaths() []*LearningPath {
	out := []*LearningPath{}
	ctx, cancel := treepg.Ctx()
	defer cancel()
	rows, err := s.pool.Query(ctx, `SELECT id, slug, name FROM learning_paths ORDER BY slug`)
	if err != nil {
		return out
	}
	defer rows.Close()
	type row struct {
		id   pgtype.UUID
		slug string
		name string
	}
	var list []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.slug, &r.name); err == nil {
			list = append(list, r)
		}
	}
	rows.Close()
	for _, r := range list {
		out = append(out, &LearningPath{ID: r.slug, Name: r.name, NodeIDs: s.pathNodes(ctx, r.id)})
	}
	return out
}
