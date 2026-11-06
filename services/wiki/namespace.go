package wiki

import "fmt"

// Namespace represents a wiki namespace identifier
type Namespace int

const (
	NSMain          Namespace = 0
	NSTalk          Namespace = 1
	NSUser          Namespace = 2
	NSUserTalk      Namespace = 3
	NSTreewiki      Namespace = 4
	NSTreewikiTalk  Namespace = 5
	NSFile          Namespace = 6
	NSFileTalk      Namespace = 7
	NSTemplate      Namespace = 10
	NSTemplateTalk  Namespace = 11
	NSHelp          Namespace = 12
	NSHelpTalk      Namespace = 13
	NSCategory      Namespace = 14
	NSCategoryTalk  Namespace = 15
	NSTopicTree     Namespace = 20
	NSTopicTreeTalk Namespace = 21
)

// NamespaceInfo holds metadata for a namespace
type NamespaceInfo struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	CanonicalName     string `json:"canonical_name"`
	IsTalk            bool   `json:"is_talk"`
	IsSearchable      bool   `json:"is_searchable"`
	DefaultProtection string `json:"default_protection"`
}

var namespaceRegistry = map[Namespace]NamespaceInfo{
	NSMain:          {0, "Main", "", false, true, "all"},
	NSTalk:          {1, "Talk", "Talk", true, true, "all"},
	NSUser:          {2, "User", "User", false, true, "all"},
	NSUserTalk:      {3, "User_talk", "User_talk", true, true, "all"},
	NSTreewiki:      {4, "Treewiki", "Treewiki", false, true, "editor_only"},
	NSTreewikiTalk:  {5, "Treewiki_talk", "Treewiki_talk", true, true, "all"},
	NSFile:          {6, "File", "File", false, true, "all"},
	NSFileTalk:      {7, "File_talk", "File_talk", true, true, "all"},
	NSTemplate:      {10, "Template", "Template", false, true, "editor_only"},
	NSTemplateTalk:  {11, "Template_talk", "Template_talk", true, true, "all"},
	NSHelp:          {12, "Help", "Help", false, true, "editor_only"},
	NSHelpTalk:      {13, "Help_talk", "Help_talk", true, true, "all"},
	NSCategory:      {14, "Category", "Category", false, true, "all"},
	NSCategoryTalk:  {15, "Category_talk", "Category_talk", true, true, "all"},
	NSTopicTree:     {20, "TopicTree", "TopicTree", false, true, "editor_only"},
	NSTopicTreeTalk: {21, "TopicTree_talk", "TopicTree_talk", true, true, "all"},
}

// NamespaceNameToID maps a namespace name to its ID
var namespaceNameToID = map[string]Namespace{
	"":          NSMain,
	"Main":      NSMain,
	"Talk":      NSTalk,
	"User":      NSUser,
	"User_talk": NSUserTalk,
	"Treewiki":  NSTreewiki,
	"File":      NSFile,
	"Template":  NSTemplate,
	"Help":      NSHelp,
	"Category":  NSCategory,
	"TopicTree": NSTopicTree,
}

// GetNamespaceInfo returns metadata for a namespace ID
func GetNamespaceInfo(ns Namespace) (NamespaceInfo, bool) {
	info, ok := namespaceRegistry[ns]
	return info, ok
}

// ParseNamespace converts a name or id to Namespace
func ParseNamespace(name string) (Namespace, bool) {
	if ns, ok := namespaceNameToID[name]; ok {
		return ns, true
	}
	return NSMain, false
}

// AllNamespaces returns all registered namespaces
func AllNamespaces() []NamespaceInfo {
	out := make([]NamespaceInfo, 0, len(namespaceRegistry))
	for _, info := range namespaceRegistry {
		out = append(out, info)
	}
	return out
}

// String returns namespace canonical name
func (n Namespace) String() string {
	if info, ok := namespaceRegistry[n]; ok {
		if info.CanonicalName != "" {
			return info.CanonicalName
		}
		return info.Name
	}
	return fmt.Sprintf("Namespace(%d)", n)
}
