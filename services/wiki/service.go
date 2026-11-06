package wiki

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dnjstjqppk-dev/treewiki/pkg/auth"
	"github.com/dnjstjqppk-dev/treewiki/pkg/diff"
)

type WikiService struct {
	store PageStorage
}

func NewWikiService(store PageStorage) *WikiService {
	return &WikiService{store: store}
}

func (s *WikiService) protectionAllowsEdit(level, userRole string) bool {
	switch level {
	case "all":
		return true
	case "registered":
		return auth.HasPermission(userRole, auth.RoleRegistered)
	case "autoconfirmed":
		return auth.HasPermission(userRole, auth.RoleAutoConfirmed)
	case "editor_only", "editor":
		return auth.HasPermission(userRole, auth.RoleEditor)
	case "sysop":
		return auth.HasPermission(userRole, auth.RoleSysop)
	default:
		return auth.HasPermission(userRole, auth.RoleRegistered)
	}
}

func (s *WikiService) CreatePage(ns Namespace, rawTitle, content, summary, authorID, userRole string) (*Page, error) {
	title := NormalizeTitle(rawTitle)
	if title == "" {
		return nil, errors.New("title is required")
	}
	if _, err := s.store.Get(ns, title); err == nil {
		return nil, ErrPageExists
	}
	info, _ := GetNamespaceInfo(ns)
	level := info.DefaultProtection
	if !s.protectionAllowsEdit(level, userRole) {
		return nil, errors.New("insufficient permissions to create page in this namespace")
	}
	target, isRedirect := ParseRedirectTarget(content)
	page := &Page{
		Namespace:       ns,
		Title:           title,
		Content:         content,
		Summary:         summary,
		AuthorID:        authorID,
		ProtectionLevel: level,
		IsRedirect:      isRedirect,
		RedirectTarget:  target,
	}
	if err := s.store.Create(page); err != nil {
		return nil, err
	}
	rev := &Revision{PageID: page.ID, Content: content, Summary: summary, AuthorID: authorID}
	if err := s.store.AddRevision(rev); err != nil {
		return nil, err
	}
	page.CurrentRevID = rev.ID
	_ = s.store.Update(page)
	srcKey := PageKey(ns, title)
	targets := ParseInternalLinks(content)
	tKeys := make([]string, 0, len(targets))
	for _, t := range targets {
		tKeys = append(tKeys, PageKey(NSMain, t))
	}
	if err := s.store.RebuildBacklinks(srcKey, tKeys); err != nil {
		return nil, err
	}
	cats := ParseCategories(content)
	if err := s.store.RebuildCategories(srcKey, cats); err != nil {
		return nil, err
	}
	return s.store.Get(ns, title)
}

// GetPage retrieves a page (non-deleted)
func (s *WikiService) GetPage(ns Namespace, rawTitle string) (*Page, error) {
	title := NormalizeTitle(rawTitle)
	page, err := s.store.Get(ns, title)
	if err != nil {
		return nil, err
	}
	if page.IsDeleted {
		return nil, ErrPageDeleted
	}
	return page, nil
}

// ResolvePage follows redirects with cycle detection
func (s *WikiService) ResolvePage(ns Namespace, rawTitle string) (*Page, error) {
	title := NormalizeTitle(rawTitle)
	visited := map[string]bool{}
	curNS := ns
	curTitle := title
	for i := 0; i < 10; i++ {
		key := PageKey(curNS, curTitle)
		if visited[key] {
			return nil, ErrRedirectCycle
		}
		visited[key] = true
		page, err := s.store.Get(curNS, curTitle)
		if err != nil {
			return nil, err
		}
		if page.IsDeleted {
			return nil, ErrPageDeleted
		}
		if !page.IsRedirect {
			return page, nil
		}
		target := page.RedirectTarget
		nsName, rest := SplitTitleNamespace(target)
		if nsName != "" {
			if n, ok := ParseNamespace(nsName); ok {
				curNS = n
			}
		}
		curTitle = NormalizeTitle(rest)
	}
	return nil, ErrRedirectCycle
}

// UpdatePage updates page content with optimistic concurrency
func (s *WikiService) UpdatePage(ns Namespace, rawTitle, content, summary, authorID, userRole, expectedRevID string, isMinor bool) (*Page, error) {
	title := NormalizeTitle(rawTitle)
	page, err := s.store.Get(ns, title)
	if err != nil {
		return nil, err
	}
	if page.IsDeleted {
		return nil, ErrPageDeleted
	}
	if !s.protectionAllowsEdit(page.ProtectionLevel, userRole) {
		return nil, errors.New("insufficient permissions to edit this page")
	}
	if expectedRevID != "" && page.CurrentRevID != expectedRevID {
		return nil, ErrRevisionConflict
	}
	target, isRedirect := ParseRedirectTarget(content)
	oldRevID := page.CurrentRevID
	page.Content = content
	page.Summary = summary
	page.AuthorID = authorID
	page.IsRedirect = isRedirect
	page.RedirectTarget = target
	if err := s.store.Update(page); err != nil {
		return nil, err
	}
	rev := &Revision{
		PageID:           page.ID,
		Content:          content,
		Summary:          summary,
		AuthorID:         authorID,
		IsMinor:          isMinor,
		ParentRevisionID: oldRevID,
	}
	if err := s.store.AddRevision(rev); err != nil {
		return nil, err
	}
	page.CurrentRevID = rev.ID
	_ = s.store.Update(page)
	srcKey := PageKey(ns, title)
	targets := ParseInternalLinks(content)
	tKeys := make([]string, 0, len(targets))
	for _, t := range targets {
		tKeys = append(tKeys, PageKey(NSMain, t))
	}
	if err := s.store.RebuildBacklinks(srcKey, tKeys); err != nil {
		return nil, err
	}
	cats := ParseCategories(content)
	if err := s.store.RebuildCategories(srcKey, cats); err != nil {
		return nil, err
	}
	return s.store.Get(ns, title)
}

// DeletePage soft-deletes a page
func (s *WikiService) DeletePage(ns Namespace, rawTitle, userRole string) error {
	title := NormalizeTitle(rawTitle)
	page, err := s.store.Get(ns, title)
	if err != nil {
		return err
	}
	if !s.protectionAllowsEdit(page.ProtectionLevel, userRole) {
		if !auth.HasPermission(userRole, auth.RoleSysop) {
			return errors.New("insufficient permissions to delete this page")
		}
	}
	return s.store.Delete(ns, title)
}

// UndeletePage restores a soft-deleted page
func (s *WikiService) UndeletePage(ns Namespace, rawTitle, userRole string) error {
	if !auth.HasPermission(userRole, auth.RoleSysop) {
		return errors.New("insufficient permissions to undelete")
	}
	return s.store.Undelete(ns, rawTitle)
}

// MovePage moves/renames a page and optionally creates redirect
func (s *WikiService) MovePage(oldNS Namespace, oldRaw string, newNS Namespace, newRaw string, authorID, userRole string, createRedirect bool) error {
	oldTitle := NormalizeTitle(oldRaw)
	newTitle := NormalizeTitle(newRaw)
	page, err := s.store.Get(oldNS, oldTitle)
	if err != nil {
		return err
	}
	if !s.protectionAllowsEdit(page.ProtectionLevel, userRole) {
		return errors.New("insufficient permissions to move this page")
	}
	if err := s.store.Move(oldNS, oldTitle, newNS, newTitle); err != nil {
		return err
	}
	if createRedirect {
		redirectContent := "#REDIRECT [[" + newTitle + "]]"
		rp := &Page{
			Namespace:       oldNS,
			Title:           oldTitle,
			Content:         redirectContent,
			Summary:         "moved page redirect",
			AuthorID:        authorID,
			ProtectionLevel: page.ProtectionLevel,
			IsRedirect:      true,
			RedirectTarget:  newTitle,
		}
		if err := s.store.Create(rp); err != nil {
			return err
		}
		rev := &Revision{PageID: rp.ID, Content: redirectContent, Summary: "redirect", AuthorID: authorID}
		if err := s.store.AddRevision(rev); err != nil {
			return err
		}
		rp.CurrentRevID = rev.ID
		_ = s.store.Update(rp)
	}
	return nil
}

// SetProtection changes page protection level
func (s *WikiService) SetProtection(ns Namespace, rawTitle, level, userRole string) error {
	if !auth.HasPermission(userRole, auth.RoleSysop) {
		return errors.New("insufficient permissions to change protection")
	}
	page, err := s.store.Get(ns, rawTitle)
	if err != nil {
		return err
	}
	page.ProtectionLevel = level
	return s.store.Update(page)
}

// GetBacklinks returns pages linking to target
func (s *WikiService) GetBacklinks(ns Namespace, rawTitle string) []*Page {
	title := NormalizeTitle(rawTitle)
	targetKey := PageKey(ns, title)
	srcKeys := s.store.GetBacklinks(targetKey)
	out := []*Page{}
	for _, sk := range srcKeys {
		parts := strings.SplitN(sk, ":", 2)
		if len(parts) != 2 {
			continue
		}
		var n Namespace
		fmt.Sscan(parts[0], &n)
		p, err := s.store.Get(n, parts[1])
		if err == nil && !p.IsDeleted {
			out = append(out, p)
		}
	}
	return out
}

// GetCategoryMembers returns pages in a category
func (s *WikiService) GetCategoryMembers(category string) []*Page {
	cat := NormalizeTitle(category)
	keys := s.store.GetCategoryMembers(cat)
	out := []*Page{}
	for _, sk := range keys {
		parts := strings.SplitN(sk, ":", 2)
		if len(parts) != 2 {
			continue
		}
		var n Namespace
		fmt.Sscan(parts[0], &n)
		p, err := s.store.Get(n, parts[1])
		if err == nil && !p.IsDeleted {
			out = append(out, p)
		}
	}
	return out
}

// GetRevisions returns revision history
func (s *WikiService) GetRevisions(ns Namespace, rawTitle string) []*Revision {
	title := NormalizeTitle(rawTitle)
	page, err := s.store.Get(ns, title)
	if err != nil {
		return nil
	}
	return s.store.GetRevisions(page.ID)
}

// ListPages returns pages in namespace
func (s *WikiService) ListPages(ns Namespace) []*Page {
	keys := s.store.ListByNamespace(ns)
	out := []*Page{}
	for _, k := range keys {
		parts := strings.SplitN(k, ":", 2)
		if len(parts) != 2 {
			continue
		}
		var n Namespace
		fmt.Sscan(parts[0], &n)
		p, err := s.store.Get(n, parts[1])
		if err == nil {
			out = append(out, p)
		}
	}
	return out
}

// GetContributorStats returns edit counts per author across all pages in a namespace
func (s *WikiService) GetContributorStats(ns Namespace) map[string]int {
	stats := map[string]int{}
	for _, p := range s.ListPages(ns) {
		revs := s.store.GetRevisions(p.ID)
		for _, r := range revs {
			stats[r.AuthorID]++
		}
	}
	return stats
}

// GetTimeline returns revisions sorted by date (already chronological)
func (s *WikiService) GetTimeline(ns Namespace, rawTitle string) []*Revision {
	return s.GetRevisions(ns, rawTitle)
}

// RevertToRevision creates a new revision restoring content from a previous revision
func (s *WikiService) RevertToRevision(ns Namespace, rawTitle, revID, authorID, userRole string) (*Page, error) {
	title := NormalizeTitle(rawTitle)
	page, err := s.store.Get(ns, title)
	if err != nil {
		return nil, err
	}
	if page.IsDeleted {
		return nil, ErrPageDeleted
	}
	if !s.protectionAllowsEdit(page.ProtectionLevel, userRole) {
		return nil, errors.New("insufficient permissions to revert this page")
	}
	targetRev, err := s.store.GetRevision(page.ID, revID)
	if err != nil {
		return nil, err
	}
	oldRevID := page.CurrentRevID
	page.Content = targetRev.Content
	page.Summary = "reverted to revision " + revID
	page.AuthorID = authorID
	target, isRedirect := ParseRedirectTarget(page.Content)
	page.IsRedirect = isRedirect
	page.RedirectTarget = target
	if err := s.store.Update(page); err != nil {
		return nil, err
	}
	rev := &Revision{
		PageID:           page.ID,
		Content:          page.Content,
		Summary:          page.Summary,
		AuthorID:         authorID,
		ParentRevisionID: oldRevID,
	}
	if err := s.store.AddRevision(rev); err != nil {
		return nil, err
	}
	page.CurrentRevID = rev.ID
	_ = s.store.Update(page)
	srcKey := PageKey(ns, title)
	targets := ParseInternalLinks(page.Content)
	tKeys := make([]string, 0, len(targets))
	for _, t := range targets {
		tKeys = append(tKeys, PageKey(NSMain, t))
	}
	if err := s.store.RebuildBacklinks(srcKey, tKeys); err != nil {
		return nil, err
	}
	cats := ParseCategories(page.Content)
	if err := s.store.RebuildCategories(srcKey, cats); err != nil {
		return nil, err
	}
	return s.store.Get(ns, title)
}

// DiffRevisions computes a line diff between two revisions of a page
func (s *WikiService) DiffRevisions(ns Namespace, rawTitle, fromRevID, toRevID string) ([]diff.Chunk, error) {
	title := NormalizeTitle(rawTitle)
	page, err := s.store.Get(ns, title)
	if err != nil {
		return nil, err
	}
	fromRev, err := s.store.GetRevision(page.ID, fromRevID)
	if err != nil {
		return nil, err
	}
	toRev, err := s.store.GetRevision(page.ID, toRevID)
	if err != nil {
		return nil, err
	}
	fromLines := strings.Split(fromRev.Content, "\n")
	toLines := strings.Split(toRev.Content, "\n")
	return diff.DiffLines(fromLines, toLines), nil
}

// CountAll returns the number of live pages across all namespaces.
func (s *WikiService) CountAll() int {
	return s.store.CountAll()
}

// RandomPage returns a random live page, optionally filtered by namespace.
func (s *WikiService) RandomPage(ns *Namespace) (*Page, error) {
	return s.store.Random(ns)
}

// FastRollback reverts to the previous revision
func (s *WikiService) FastRollback(ns Namespace, rawTitle, authorID, userRole string) (*Page, error) {
	revs := s.GetRevisions(ns, rawTitle)
	if len(revs) < 2 {
		return nil, errors.New("no previous revision to rollback to")
	}
	prev := revs[len(revs)-2]
	return s.RevertToRevision(ns, rawTitle, prev.ID, authorID, userRole)
}
