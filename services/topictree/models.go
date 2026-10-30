package topictree

import (
	"time"
)

type TopicNode struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Difficulty  int       `json:"difficulty"`
	Completion  float64   `json:"completion"`
	ParentIDs     []string  `json:"parent_ids"`
	ChildIDs      []string  `json:"child_ids"`
	Prerequisites []string  `json:"prerequisites"`
	Advanced      []string  `json:"advanced"`
	Related       []string  `json:"related"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TopicEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type LearningPath struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	NodeIDs []string `json:"node_ids"`
}

type GraphExport struct {
	Nodes []TopicNode  `json:"nodes"`
	Edges []TopicEdge  `json:"edges"`
}