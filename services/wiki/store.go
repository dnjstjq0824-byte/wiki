package wiki

import (
	"errors"
	"math/rand/v2"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPageNotFound     = errors.New("page not found")
	ErrPageExists       = errors.New("page already exists")
	ErrRevisionConflict = errors.New("revision conflict: page was modified")
	ErrRedirectCycle    = errors.New("redirect cycle detected")
	ErrPageDeleted      = errors.New("page is deleted")
)

var internalLinkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

// PageStorage abstracts page persistence so the wiki service runs on the
// in-memory PageStore (development/tests) or PostgresPageStore (production).
type PageStorage interface {
	Create(page *Page) error
	Get(ns Namespace, title string) (*Page, error)
	Update(page *Page) error
	Delete(ns Namespace, title string) error
	Undelete(ns Namespace, title string) error
	Move(oldNS Namespace, oldTitle string, newNS Namespace, newTitle string) error
	AddRevision(rev *Revision) error
	GetRevisions(pageID string) []*Revision
	GetRevision(pageID, revID string) (*Revision, error)
	RebuildBacklinks(sourceKey string, targets []string) error
	GetBacklinks(targetKey string) []string
	RebuildCategories(sourceKey string, cats []string) error
	GetCategoryMembers(category string) []string
	ListByNamespace(ns Namespace) []string
	CountAll() int
	CountByNamespace(ns Namespace) int
	Random(ns *Namespace) (*Page, error)
}

type PageStore struct {
	mu         sync.RWMutex
	pages      map[string]*Page
	revisions  map[string][]*Revision
	backlinks  map[string]map[string]bool
	categories map[string]map[string]bool
}

var _ PageStorage = (*PageStore)(nil)

func NewPageStore() *PageStore {
	return &PageStore{
		pages:      make(map[string]*Page),
		revisions:  make(map[string][]*Revision),
		backlinks:  make(map[string]map[string]bool),
		categories: make(map[string]map[string]bool),
	}
}

func (s *PageStore) key(ns Namespace, title string) string {
	return PageKey(ns, title)
}

func clonePage(p *Page) *Page {
	if p == nil {
		return nil
	}
	cp := *p
	if p.DeletedAt != nil {
		t := *p.DeletedAt
		cp.DeletedAt = &t
	}
	return &cp
}

func (s *PageStore) Create(page *Page) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := s.key(page.Namespace, page.Title)
	if _, exists := s.pages[k]; exists {
		return ErrPageExists
	}
	if page.ID == "" {
		page.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	page.CreatedAt = now
	page.UpdatedAt = now
	if page.ProtectionLevel == "" {
		if info, ok := GetNamespaceInfo(page.Namespace); ok {
			page.ProtectionLevel = info.DefaultProtection
		} else {
			page.ProtectionLevel = "all"
		}
	}
	s.pages[k] = page
	return nil
}

func (s *PageStore) Get(ns Namespace, title string) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pages[s.key(ns, title)]
	if !ok {
		return nil, ErrPageNotFound
	}
	return clonePage(p), nil
}

func (s *PageStore) Update(page *Page) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := s.key(page.Namespace, page.Title)
	existing, ok := s.pages[k]
	if !ok {
		return ErrPageNotFound
	}
	page.UpdatedAt = time.Now().UTC()
	page.CreatedAt = existing.CreatedAt
	s.pages[k] = page
	return nil
}

func (s *PageStore) Delete(ns Namespace, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := s.key(ns, title)
	p, ok := s.pages[k]
	if !ok {
		return ErrPageNotFound
	}
	now := time.Now().UTC()
	p.IsDeleted = true
	p.DeletedAt = &now
	p.UpdatedAt = now
	return nil
}

func (s *PageStore) Undelete(ns Namespace, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := s.key(ns, title)
	p, ok := s.pages[k]
	if !ok {
		return ErrPageNotFound
	}
	p.IsDeleted = false
	p.DeletedAt = nil
	p.UpdatedAt = time.Now().UTC()
	return nil
}
func (s *PageStore) Move(oldNS Namespace, oldTitle string, newNS Namespace, newTitle string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	oldK := s.key(oldNS, oldTitle)
	p, ok := s.pages[oldK]
	if !ok {
		return ErrPageNotFound
	}
	newK := s.key(newNS, newTitle)
	if _, exists := s.pages[newK]; exists {
		return ErrPageExists
	}
	delete(s.pages, oldK)
	p.Namespace = newNS
	p.Title = newTitle
	p.UpdatedAt = time.Now().UTC()
	s.pages[newK] = p
	for target, sources := range s.backlinks {
		if sources[oldK] {
			delete(s.backlinks[target], oldK)
			s.backlinks[target][newK] = true
		}
	}
	for cat, members := range s.categories {
		if members[oldK] {
			delete(s.categories[cat], oldK)
			s.categories[cat][newK] = true
		}
	}
	return nil
}

func (s *PageStore) AddRevision(rev *Revision) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rev.ID == "" {
		rev.ID = uuid.New().String()
	}
	if rev.CreatedAt.IsZero() {
		rev.CreatedAt = time.Now().UTC()
	}
	s.revisions[rev.PageID] = append(s.revisions[rev.PageID], rev)
	return nil
}

func (s *PageStore) GetRevisions(pageID string) []*Revision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Revision, len(s.revisions[pageID]))
	copy(out, s.revisions[pageID])
	return out
}

func (s *PageStore) GetRevision(pageID, revID string) (*Revision, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.revisions[pageID] {
		if r.ID == revID {
			cp := *r
			return &cp, nil
		}
	}
	return nil, errors.New("revision not found")
}

func (s *PageStore) RebuildBacklinks(sourceKey string, targets []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sources := range s.backlinks {
		delete(sources, sourceKey)
	}
	for _, t := range targets {
		if s.backlinks[t] == nil {
			s.backlinks[t] = make(map[string]bool)
		}
		s.backlinks[t][sourceKey] = true
	}
	return nil
}

func (s *PageStore) GetBacklinks(targetKey string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0)
	for src := range s.backlinks[targetKey] {
		out = append(out, src)
	}
	return out
}

func (s *PageStore) RebuildCategories(sourceKey string, cats []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, members := range s.categories {
		delete(members, sourceKey)
	}
	for _, c := range cats {
		if s.categories[c] == nil {
			s.categories[c] = make(map[string]bool)
		}
		s.categories[c][sourceKey] = true
	}
	return nil
}

func (s *PageStore) GetCategoryMembers(category string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0)
	for m := range s.categories[category] {
		out = append(out, m)
	}
	return out
}

func (s *PageStore) ListByNamespace(ns Namespace) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0)
	for k, p := range s.pages {
		if p.Namespace == ns && !p.IsDeleted {
			out = append(out, k)
		}
	}
	return out
}

// CountAll returns the number of non-deleted pages across all namespaces.
func (s *PageStore) CountAll() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, p := range s.pages {
		if !p.IsDeleted {
			n++
		}
	}
	return n
}

// CountByNamespace returns the number of non-deleted pages in one namespace.
func (s *PageStore) CountByNamespace(ns Namespace) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, p := range s.pages {
		if p.Namespace == ns && !p.IsDeleted {
			n++
		}
	}
	return n
}

// Random returns a random non-deleted page. When ns is nil, all namespaces
// are eligible; otherwise only that namespace is considered.
func (s *PageStore) Random(ns *Namespace) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	candidates := make([]*Page, 0)
	for _, p := range s.pages {
		if p.IsDeleted {
			continue
		}
		if ns != nil && p.Namespace != *ns {
			continue
		}
		candidates = append(candidates, p)
	}
	if len(candidates) == 0 {
		return nil, ErrPageNotFound
	}
	return clonePage(candidates[rand.N(len(candidates))]), nil
}

func ParseInternalLinks(content string) []string {
	matches := internalLinkRe.FindAllStringSubmatch(content, -1)
	seen := map[string]bool{}
	out := []string{}
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		raw := m[1]
		if idx := strings.Index(raw, "|"); idx >= 0 {
			raw = raw[:idx]
		}
		if idx := strings.Index(raw, "#"); idx >= 0 {
			raw = raw[:idx]
		}
		target := NormalizeTitle(strings.TrimSpace(raw))
		if target == "" {
			continue
		}
		if !seen[target] {
			seen[target] = true
			out = append(out, target)
		}
	}
	return out
}

func ParseCategories(content string) []string {
	links := ParseInternalLinks(content)
	out := []string{}
	for _, l := range links {
		if strings.HasPrefix(l, "Category:") || strings.HasPrefix(l, "분류:") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) == 2 {
				out = append(out, NormalizeTitle(parts[1]))
			}
		}
	}
	return out
}

func ParseRedirectTarget(content string) (string, bool) {
	trimmed := strings.TrimSpace(content)
	lower := strings.ToLower(trimmed)
	if !strings.HasPrefix(lower, "#redirect") {
		return "", false
	}
	idx := strings.Index(trimmed, "[[")
	if idx < 0 {
		return "", false
	}
	end := strings.Index(trimmed[idx:], "]]")
	if end < 0 {
		return "", false
	}
	raw := trimmed[idx+2 : idx+end]
	if pipe := strings.Index(raw, "|"); pipe >= 0 {
		raw = raw[:pipe]
	}
	return NormalizeTitle(strings.TrimSpace(raw)), true
}
