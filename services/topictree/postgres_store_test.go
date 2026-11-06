package topictree

import (
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/db"
)

func pgTestService(t *testing.T) *TopicService {
	t.Helper()
	url := os.Getenv("TREEWIKI_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TREEWIKI_TEST_DATABASE_URL not set; skipping postgres integration test")
	}
	pgDB, err := db.Connect(url)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	t.Cleanup(func() { pgDB.Close() })
	if err := pgDB.MigrateAll("../../database/migrations"); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return NewTopicService(NewPostgresTopicStore(pgDB.Pool))
}

func TestPostgresTopicDAG(t *testing.T) {
	svc := pgTestService(t)
	suffix := uuid.NewString()[:8]
	root, mid, leaf := "pgroot_"+suffix, "pgmid_"+suffix, "pgleaf_"+suffix

	if err := svc.AddNode(&TopicNode{ID: root, Name: "Root", Description: "r", Difficulty: 1, Completion: 0.5, Prerequisites: []string{"pgpre"}}); err != nil {
		t.Fatalf("AddNode root: %v", err)
	}
	if err := svc.AddNode(&TopicNode{ID: mid, Name: "Mid", Difficulty: 2}); err != nil {
		t.Fatalf("AddNode mid: %v", err)
	}
	if err := svc.AddNode(&TopicNode{ID: leaf, Name: "Leaf", Difficulty: 3}); err != nil {
		t.Fatalf("AddNode leaf: %v", err)
	}
	if err := svc.AddNode(&TopicNode{ID: root, Name: "Dup"}); err != ErrNodeExists {
		t.Fatalf("expected ErrNodeExists, got %v", err)
	}
	if err := svc.Link(root, mid); err != nil {
		t.Fatalf("Link root->mid: %v", err)
	}
	if err := svc.Link(mid, leaf); err != nil {
		t.Fatalf("Link mid->leaf: %v", err)
	}
	if err := svc.Link(leaf, root); err != ErrCycleDetected {
		t.Fatalf("expected cycle detection, got %v", err)
	}

	got, err := svc.GetNode(mid)
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if len(got.ParentIDs) != 1 || got.ParentIDs[0] != root {
		t.Fatalf("parents wrong: %+v", got.ParentIDs)
	}
	if len(got.ChildIDs) != 1 || got.ChildIDs[0] != leaf {
		t.Fatalf("children wrong: %+v", got.ChildIDs)
	}
	rootNode, _ := svc.GetNode(root)
	if len(rootNode.Prerequisites) != 1 || rootNode.Completion != 0.5 {
		t.Fatalf("attributes not persisted: %+v", rootNode)
	}

	desc := svc.Descendants(root)
	if len(desc) != 2 {
		t.Fatalf("expected 2 descendants, got %v", desc)
	}
	anc := svc.Ancestors(leaf)
	if len(anc) != 2 {
		t.Fatalf("expected 2 ancestors, got %v", anc)
	}

	pathID := "pgpath_" + suffix
	if err := svc.CreatePath(&LearningPath{ID: pathID, Name: "PG Path", NodeIDs: []string{root, mid, leaf}}); err != nil {
		t.Fatalf("CreatePath: %v", err)
	}
	path, err := svc.GetPath(pathID)
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if len(path.NodeIDs) != 3 || path.NodeIDs[2] != leaf {
		t.Fatalf("path order wrong: %+v", path.NodeIDs)
	}

	found := false
	for _, n := range svc.AllNodes() {
		if n.ID == leaf {
			found = true
		}
	}
	if !found {
		t.Fatal("AllNodes missing leaf")
	}
	if err := svc.DeleteNode(leaf); err != nil {
		t.Fatalf("DeleteNode: %v", err)
	}
	if _, err := svc.GetNode(leaf); err != ErrNodeNotFound {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}
