package topictree

type TopicService struct {
	store *TopicStore
}

func NewTopicService(store *TopicStore) *TopicService {
	return &TopicService{store: store}
}

func (s *TopicService) AddNode(n *TopicNode) error {
	return s.store.CreateNode(n)
}

func (s *TopicService) GetNode(id string) (*TopicNode, error) {
	return s.store.GetNode(id)
}

func (s *TopicService) UpdateNode(n *TopicNode) error {
	return s.store.UpdateNode(n)
}

func (s *TopicService) DeleteNode(id string) error {
	return s.store.DeleteNode(id)
}

func (s *TopicService) Link(parentID, childID string) error {
	return s.store.AddEdge(parentID, childID)
}

func (s *TopicService) Unlink(parentID, childID string) error {
	return s.store.RemoveEdge(parentID, childID)
}

func (s *TopicService) Ancestors(id string) []string {
	return s.store.GetAncestors(id)
}

func (s *TopicService) Descendants(id string) []string {
	return s.store.GetDescendants(id)
}

func (s *TopicService) CreatePath(p *LearningPath) error {
	return s.store.CreatePath(p)
}

func (s *TopicService) GetPath(id string) (*LearningPath, error) {
	return s.store.GetPath(id)
}

func (s *TopicService) AllPaths() []*LearningPath {
	return s.store.AllPaths()
}

func (s *TopicService) Export() GraphExport {
	return s.store.ExportGraph()
}

func (s *TopicService) AllNodes() []*TopicNode {
	return s.store.AllNodes()
}

func (s *TopicService) ReorderNode(nodeID, oldParentID, newParentID string) error {
	if oldParentID != "" {
		if err := s.store.RemoveEdge(oldParentID, nodeID); err != nil {
			return err
		}
	}
	if newParentID != "" {
		return s.store.AddEdge(newParentID, nodeID)
	}
	return nil
}