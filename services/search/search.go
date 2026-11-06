package search

import (
	"sort"
	"strings"
	"sync"
)

type SearchDocument struct {
	ID        string `json:"id"`
	Namespace int    `json:"namespace"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}

type SearchResult struct {
	Document SearchDocument `json:"document"`
	Score    float64        `json:"score"`
	Snippet  string         `json:"snippet"`
}

// SearchBackend abstracts full-text search so handlers run on the in-memory
// SearchIndex (development/tests) or PostgresSearchBackend (production SQL
// full-text over the pages table).
type SearchBackend interface {
	Index(doc *SearchDocument)
	Remove(id string)
	Search(query string, namespace int) []SearchResult
	Autocomplete(prefix string, limit int) []string
	SearchFaceted(query string, namespace int, category string, from, to string) []SearchResult
}

// MediaStorage abstracts media metadata persistence.
type MediaStorage interface {
	Add(a *MediaAsset) error
	Get(id string) (*MediaAsset, bool)
	List() []*MediaAsset
}

type SearchIndex struct {
	mu   sync.RWMutex
	docs map[string]*SearchDocument
}

var (
	_ SearchBackend = (*SearchIndex)(nil)
	_ MediaStorage  = (*MediaStore)(nil)
)

func NewSearchIndex() *SearchIndex {
	return &SearchIndex{docs: make(map[string]*SearchDocument)}
}

func (idx *SearchIndex) Index(doc *SearchDocument) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.docs[doc.ID] = doc
}

func (idx *SearchIndex) Remove(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.docs, id)
}

func (idx *SearchIndex) Search(query string, namespace int) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return []SearchResult{}
	}
	results := []SearchResult{}
	for _, d := range idx.docs {
		if namespace >= 0 && d.Namespace != namespace {
			continue
		}
		text := strings.ToLower(d.Title + " " + d.Content)
		if !strings.Contains(text, q) {
			continue
		}
		score := float64(strings.Count(text, q))
		// title boost
		if strings.Contains(strings.ToLower(d.Title), q) {
			score += 5
		}
		snippet := makeSnippet(d.Content, q)
		results = append(results, SearchResult{Document: *d, Score: score, Snippet: snippet})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

func (idx *SearchIndex) Autocomplete(prefix string, limit int) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	p := strings.ToLower(strings.TrimSpace(prefix))
	if p == "" {
		return []string{}
	}
	out := []string{}
	for _, d := range idx.docs {
		if strings.HasPrefix(strings.ToLower(d.Title), p) {
			out = append(out, d.Title)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}

func (idx *SearchIndex) SearchFaceted(query string, namespace int, category string, from, to string) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	q := strings.ToLower(strings.TrimSpace(query))
	results := []SearchResult{}
	for _, d := range idx.docs {
		if namespace >= 0 && d.Namespace != namespace {
			continue
		}
		text := strings.ToLower(d.Title + " " + d.Content)
		if q != "" && !strings.Contains(text, q) {
			continue
		}
		if category != "" && !strings.Contains(strings.ToLower(d.Content), "category:"+strings.ToLower(category)) {
			// simple category filter via content
		}
		score := float64(0)
		if q != "" {
			score = float64(strings.Count(text, q))
			if strings.Contains(strings.ToLower(d.Title), q) {
				score += 5
			}
		} else {
			score = 1
		}
		snippet := ""
		if q != "" {
			snippet = makeSnippet(d.Content, q)
		} else {
			if len(d.Content) > 120 {
				snippet = d.Content[:120] + "..."
			} else {
				snippet = d.Content
			}
		}
		results = append(results, SearchResult{Document: *d, Score: score, Snippet: snippet})
	}
	return results
}
func makeSnippet(content, q string) string {
	lower := strings.ToLower(content)
	idx := strings.Index(lower, q)
	if idx < 0 {
		if len(content) > 120 {
			return content[:120] + "..."
		}
		return content
	}
	start := idx - 60
	if start < 0 {
		start = 0
	}
	end := idx + len(q) + 60
	if end > len(content) {
		end = len(content)
	}
	return "..." + content[start:end] + "..."
}

// Media asset metadata
type MediaAsset struct {
	ID           string `json:"id"`
	Filename     string `json:"filename"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	FileSize     int64  `json:"file_size_bytes"`
	S3Key        string `json:"s3_key"`
	SHA256       string `json:"sha256_hash"`
	LicenseType  string `json:"license_type"`
	UploaderID   string `json:"uploader_id"`
	CreatedAt    string `json:"created_at"`
}

type MediaStore struct {
	mu     sync.RWMutex
	assets map[string]*MediaAsset
}

func NewMediaStore() *MediaStore {
	return &MediaStore{assets: make(map[string]*MediaAsset)}
}

func (m *MediaStore) Add(a *MediaAsset) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a.ID == "" {
		a.ID = "media-" + a.Filename
	}
	m.assets[a.ID] = a
	return nil
}

func (m *MediaStore) Get(id string) (*MediaAsset, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.assets[id]
	return a, ok
}

func (m *MediaStore) List() []*MediaAsset {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []*MediaAsset{}
	for _, a := range m.assets {
		out = append(out, a)
	}
	return out
}
