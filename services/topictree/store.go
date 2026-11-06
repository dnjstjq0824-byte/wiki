package topictree

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNodeNotFound  = errors.New("node not found")
	ErrNodeExists    = errors.New("node already exists")
	ErrCycleDetected = errors.New("cycle detected")
	ErrPathNotFound  = errors.New("learning path not found")
)

// TopicStorage abstracts topic persistence so the service runs on the
// in-memory TopicStore (development/tests) or PostgresTopicStore (production).
type TopicStorage interface {
	CreateNode(n *TopicNode) error
	GetNode(id string) (*TopicNode, error)
	UpdateNode(n *TopicNode) error
	DeleteNode(id string) error
	AllNodes() []*TopicNode
	AddEdge(parentID, childID string) error
	RemoveEdge(parentID, childID string) error
	CreatePath(p *LearningPath) error
	GetPath(id string) (*LearningPath, error)
	AllPaths() []*LearningPath
	GetDescendants(id string) []string
	GetAncestors(id string) []string
	ExportGraph() GraphExport
}

type TopicStore struct {
	mu    sync.RWMutex
	nodes map[string]*TopicNode
	paths map[string]*LearningPath
}

var _ TopicStorage = (*TopicStore)(nil)

func NewTopicStore() *TopicStore {
	return &TopicStore{nodes: make(map[string]*TopicNode), paths: make(map[string]*LearningPath)}
}

func cloneNode(n *TopicNode) *TopicNode {
	if n == nil {
		return nil
	}
	cp := *n
	cp.ParentIDs = append([]string{}, n.ParentIDs...)
	cp.ChildIDs = append([]string{}, n.ChildIDs...)
	cp.Prerequisites = append([]string{}, n.Prerequisites...)
	cp.Advanced = append([]string{}, n.Advanced...)
	cp.Related = append([]string{}, n.Related...)
	return &cp
}

func (s *TopicStore) CreateNode(n *TopicNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	if _, exists := s.nodes[n.ID]; exists {
		return ErrNodeExists
	}
	if n.Name == "" {
		return errors.New("name required")
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
	s.nodes[n.ID] = n
	return nil
}

func (s *TopicStore) GetNode(id string) (*TopicNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.nodes[id]
	if !ok {
		return nil, ErrNodeNotFound
	}
	return cloneNode(n), nil
}

func (s *TopicStore) UpdateNode(n *TopicNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.nodes[n.ID]
	if !ok {
		return ErrNodeNotFound
	}
	n.CreatedAt = existing.CreatedAt
	n.UpdatedAt = time.Now().UTC()
	s.nodes[n.ID] = n
	return nil
}

func (s *TopicStore) DeleteNode(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[id]; !ok {
		return ErrNodeNotFound
	}
	delete(s.nodes, id)
	for _, n := range s.nodes {
		filtered := []string{}
		for _, c := range n.ChildIDs {
			if c != id {
				filtered = append(filtered, c)
			}
		}
		n.ChildIDs = filtered
		filteredP := []string{}
		for _, p := range n.ParentIDs {
			if p != id {
				filteredP = append(filteredP, p)
			}
		}
		n.ParentIDs = filteredP
	}
	return nil
}

func (s *TopicStore) AllNodes() []*TopicNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*TopicNode, 0, len(s.nodes))
	for _, n := range s.nodes {
		out = append(out, cloneNode(n))
	}
	return out
}
func (s *TopicStore) AddEdge(parentID, childID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	parent, ok := s.nodes[parentID]
	if !ok {
		return ErrNodeNotFound
	}
	child, ok := s.nodes[childID]
	if !ok {
		return ErrNodeNotFound
	}
	if parentID == childID {
		return ErrCycleDetected
	}
	visited := map[string]bool{}
	var dfs func(id string) bool
	dfs = func(id string) bool {
		if id == childID {
			return true
		}
		if visited[id] {
			return false
		}
		visited[id] = true
		n, ok := s.nodes[id]
		if !ok {
			return false
		}
		for _, p := range n.ParentIDs {
			if dfs(p) {
				return true
			}
		}
		return false
	}
	if dfs(parentID) {
		return ErrCycleDetected
	}
	for _, p := range child.ParentIDs {
		if p == parentID {
			s.nodes[parentID] = parent
			s.nodes[childID] = child
			return nil
		}
	}
	child.ParentIDs = append(child.ParentIDs, parentID)
	for _, c := range parent.ChildIDs {
		if c == childID {
			s.nodes[parentID] = parent
			s.nodes[childID] = child
			return nil
		}
	}
	parent.ChildIDs = append(parent.ChildIDs, childID)
	s.nodes[parentID] = parent
	s.nodes[childID] = child
	return nil
}

func (s *TopicStore) RemoveEdge(parentID, childID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	parent, ok := s.nodes[parentID]
	if !ok {
		return ErrNodeNotFound
	}
	child, ok := s.nodes[childID]
	if !ok {
		return ErrNodeNotFound
	}
	filtered := []string{}
	for _, p := range child.ParentIDs {
		if p != parentID {
			filtered = append(filtered, p)
		}
	}
	child.ParentIDs = filtered
	filteredC := []string{}
	for _, c := range parent.ChildIDs {
		if c != childID {
			filteredC = append(filteredC, c)
		}
	}
	parent.ChildIDs = filteredC
	s.nodes[parentID] = parent
	s.nodes[childID] = child
	return nil
}

func (s *TopicStore) CreatePath(p *LearningPath) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	s.paths[p.ID] = p
	return nil
}

func (s *TopicStore) GetPath(id string) (*LearningPath, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.paths[id]
	if !ok {
		return nil, ErrPathNotFound
	}
	cp := *p
	cp.NodeIDs = append([]string{}, p.NodeIDs...)
	return &cp, nil
}

func (s *TopicStore) AllPaths() []*LearningPath {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*LearningPath, 0, len(s.paths))
	for _, p := range s.paths {
		cp := *p
		cp.NodeIDs = append([]string{}, p.NodeIDs...)
		out = append(out, &cp)
	}
	return out
}

func (s *TopicStore) GetDescendants(id string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []string{}
	visited := map[string]bool{}
	var dfs func(cur string)
	dfs = func(cur string) {
		n, ok := s.nodes[cur]
		if !ok {
			return
		}
		for _, c := range n.ChildIDs {
			if !visited[c] {
				visited[c] = true
				out = append(out, c)
				dfs(c)
			}
		}
	}
	dfs(id)
	return out
}

func (s *TopicStore) GetAncestors(id string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []string{}
	visited := map[string]bool{}
	var dfs func(cur string)
	dfs = func(cur string) {
		n, ok := s.nodes[cur]
		if !ok {
			return
		}
		for _, p := range n.ParentIDs {
			if !visited[p] {
				visited[p] = true
				out = append(out, p)
				dfs(p)
			}
		}
	}
	dfs(id)
	return out
}

func (s *TopicStore) ExportGraph() GraphExport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ge := GraphExport{Nodes: []TopicNode{}, Edges: []TopicEdge{}}
	for _, n := range s.nodes {
		ge.Nodes = append(ge.Nodes, *cloneNode(n))
	}
	for _, n := range s.nodes {
		for _, c := range n.ChildIDs {
			ge.Edges = append(ge.Edges, TopicEdge{From: n.ID, To: c})
		}
	}
	return ge
}
