package wiki

import (
	"testing"
)

func TestWikiCreateAndGet(t *testing.T) {
	store := NewPageStore()
	svc := NewWikiService(store)
	page, err := svc.CreatePage(NSMain, "Hello World", "'''Hello''' [[Category:Test]]", "init", "user1", "REGISTERED")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if page.Title != "Hello_World" {
		t.Errorf("title normalize: %s", page.Title)
	}
	got, err := svc.GetPage(NSMain, "Hello World")
	if err != nil || got.ID != page.ID {
		t.Fatalf("get failed: %v", err)
	}
}

func TestWikiUpdateConflict(t *testing.T) {
	store := NewPageStore()
	svc := NewWikiService(store)
	page, _ := svc.CreatePage(NSMain, "ConflictPage", "v1", "init", "u1", "REGISTERED")
	// first update
	_, err := svc.UpdatePage(NSMain, "ConflictPage", "v2", "upd", "u1", "REGISTERED", page.CurrentRevID, false)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	// stale revision
	_, err = svc.UpdatePage(NSMain, "ConflictPage", "v3", "upd2", "u1", "REGISTERED", page.CurrentRevID, false)
	if err != ErrRevisionConflict {
		t.Errorf("expected conflict, got %v", err)
	}
}

func TestWikiRedirectAndResolve(t *testing.T) {
	store := NewPageStore()
	svc := NewWikiService(store)
	svc.CreatePage(NSMain, "Target", "target content", "init", "u1", "REGISTERED")
	svc.CreatePage(NSMain, "Source", "#REDIRECT [[Target]]", "init", "u1", "REGISTERED")
	resolved, err := svc.ResolvePage(NSMain, "Source")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if resolved.Title != "Target" {
		t.Errorf("resolved to %s", resolved.Title)
	}
}

func TestWikiBacklinksAndCategories(t *testing.T) {
	store := NewPageStore()
	svc := NewWikiService(store)
	svc.CreatePage(NSMain, "A", "links to [[B]] and [[Category:C]]", "init", "u1", "REGISTERED")
	svc.CreatePage(NSMain, "B", "content", "init", "u1", "REGISTERED")
	bl := svc.GetBacklinks(NSMain, "B")
	if len(bl) != 1 {
		t.Errorf("expected 1 backlink, got %d", len(bl))
	}
	members := svc.GetCategoryMembers("C")
	if len(members) != 1 {
		t.Errorf("expected 1 category member, got %d", len(members))
	}
}

func TestWikiMoveAndDelete(t *testing.T) {
	store := NewPageStore()
	svc := NewWikiService(store)
	svc.CreatePage(NSMain, "OldName", "content", "init", "u1", "REGISTERED")
	if err := svc.MovePage(NSMain, "OldName", NSMain, "NewName", "u1", "REGISTERED", true); err != nil {
		t.Fatalf("move failed: %v", err)
	}
	if _, err := svc.GetPage(NSMain, "OldName"); err != nil {
		// redirect page should exist
	}
	// redirect created
	p, err := svc.GetPage(NSMain, "OldName")
	if err != nil || !p.IsRedirect {
		t.Errorf("redirect not created: %v", err)
	}
	// delete
	if err := svc.DeletePage(NSMain, "NewName", "REGISTERED"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := svc.GetPage(NSMain, "NewName"); err != ErrPageDeleted {
		t.Errorf("expected deleted, got %v", err)
	}
	// undelete requires sysop
	if err := svc.UndeletePage(NSMain, "NewName", "SYSOP"); err != nil {
		t.Fatalf("undelete failed: %v", err)
	}
}