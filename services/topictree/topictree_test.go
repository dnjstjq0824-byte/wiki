package topictree

import (
	"testing"
)

func TestTopicTreeDAGAndCycleDetection(t *testing.T) {
	store := NewTopicStore()
	svc := NewTopicService(store)

	// 1. Create Nodes
	err := svc.AddNode(&TopicNode{
		ID:          "cs-root",
		Name:        "Computer Science",
		Description: "Foundational study of computation and information",
		Difficulty:  1,
	})
	if err != nil {
		t.Fatalf("Create root failed: %v", err)
	}

	err = svc.AddNode(&TopicNode{
		ID:          "algorithms",
		Name:        "Algorithms & Data Structures",
		Description: "Step-by-step procedures for calculation",
		Difficulty:  2,
	})
	if err != nil {
		t.Fatalf("Create child failed: %v", err)
	}

	err = svc.AddNode(&TopicNode{
		ID:          "quantum-algos",
		Name:        "Quantum Algorithms",
		Description: "Algorithms operating on quantum computing substrates (Shor, Grover)",
		Difficulty:  4,
	})
	if err != nil {
		t.Fatalf("Create leaf failed: %v", err)
	}

	// 2. Link Edges
	if err := svc.Link("cs-root", "algorithms"); err != nil {
		t.Fatalf("Link cs-root -> algorithms failed: %v", err)
	}
	if err := svc.Link("algorithms", "quantum-algos"); err != nil {
		t.Fatalf("Link algorithms -> quantum-algos failed: %v", err)
	}

	// 3. Test Cycle Detection
	// Linking quantum-algos -> cs-root should FAIL with ErrCycleDetected
	err = svc.Link("quantum-algos", "cs-root")
	if err != ErrCycleDetected {
		t.Errorf("expected cycle detection error, got: %v", err)
	}

	// 4. Test Ancestors and Descendants
	descendants := svc.Descendants("cs-root")
	if len(descendants) != 2 {
		t.Errorf("expected 2 descendants for cs-root, got: %+v", descendants)
	}

	ancestors := svc.Ancestors("quantum-algos")
	if len(ancestors) != 2 {
		t.Errorf("expected 2 ancestors for quantum-algos, got: %+v", ancestors)
	}

	// 5. Test Graph Export
	export := svc.Export()
	if len(export.Nodes) != 3 || len(export.Edges) != 2 {
		t.Errorf("unexpected graph export: nodes=%d, edges=%d", len(export.Nodes), len(export.Edges))
	}

	// 6. Test Learning Path
	path := &LearningPath{
		ID:      "path-1",
		Name:    "Quantum Journey",
		NodeIDs: []string{"cs-root", "algorithms", "quantum-algos"},
	}
	if err := svc.CreatePath(path); err != nil {
		t.Fatalf("CreatePath failed: %v", err)
	}
	fetchedPath, err := svc.GetPath("path-1")
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}
	if len(fetchedPath.NodeIDs) != 3 {
		t.Errorf("expected 3 nodes in learning path, got: %d", len(fetchedPath.NodeIDs))
	}
}
