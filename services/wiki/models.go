package wiki

import (
	"fmt"
	"time"
)

// Page represents a wiki document
type Page struct {
	ID              string     `json:"id"`
	Namespace       Namespace  `json:"namespace"`
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	Summary         string     `json:"summary"`
	AuthorID        string     `json:"author_id"`
	ProtectionLevel string     `json:"protection_level"`
	IsRedirect      bool       `json:"is_redirect"`
	RedirectTarget  string     `json:"redirect_target,omitempty"`
	IsDeleted       bool       `json:"is_deleted"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CurrentRevID    string     `json:"current_revision_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Revision represents an immutable revision entry
type Revision struct {
	ID               string    `json:"id"`
	PageID           string    `json:"page_id"`
	Content          string    `json:"content"`
	Summary          string    `json:"summary"`
	AuthorID         string    `json:"author_id"`
	IsMinor          bool      `json:"is_minor"`
	ParentRevisionID string    `json:"parent_revision_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// PageKey returns the storage key for a page
func PageKey(ns Namespace, title string) string {
	return fmt.Sprintf("%d:%s", ns, title)
}