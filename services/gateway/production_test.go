package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsEndpoint(t *testing.T) {
	router := NewRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "treewiki_http_requests_total") {
		t.Errorf("expected prometheus counters, got: %s", rec.Body.String())
	}
}

func TestStatsEndpointReportsSeededContent(t *testing.T) {
	router := NewRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode stats: %v", err)
	}
	if body["version"] == "" {
		t.Error("expected version in stats")
	}
	if pages, ok := body["pages"].(float64); !ok || pages < 6 {
		t.Errorf("expected at least 6 seeded pages, got %v", body["pages"])
	}
	if topics, ok := body["topics"].(float64); !ok || topics < 6 {
		t.Errorf("expected at least 6 seeded topics, got %v", body["topics"])
	}
}

func TestRandomPageReturnsSeededArticle(t *testing.T) {
	router := NewRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pages/random", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var page map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("failed to decode page: %v", err)
	}
	if page["title"] == "" || page["content"] == "" {
		t.Errorf("expected title and content, got: %v", page)
	}
}

func TestPaginationEnvelopeOnListPages(t *testing.T) {
	router := NewRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/0/pages?per_page=2", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode envelope: %v", err)
	}
	if _, ok := body["data"]; !ok {
		t.Fatalf("expected paginated envelope with data, got: %v", body)
	}
	if total, ok := body["total"].(float64); !ok || total < 6 {
		t.Errorf("expected total >= 6, got %v", body["total"])
	}
}

func TestListPagesWithoutPaginationStaysArray(t *testing.T) {
	router := NewRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/0/pages", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var arr []any
	if err := json.Unmarshal(rec.Body.Bytes(), &arr); err != nil {
		t.Fatalf("expected backward-compatible JSON array, got: %s", rec.Body.String())
	}
}

func TestSecurityHeadersAndRequestID(t *testing.T) {
	router := NewRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header")
	}
	for _, h := range []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy"} {
		if rec.Header().Get(h) == "" {
			t.Errorf("missing header %s", h)
		}
	}
}

func TestAutocompleteAcceptsPrefixParam(t *testing.T) {
	router := NewRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/autocomplete?prefix=algo&limit=6", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var suggestions []string
	if err := json.Unmarshal(rec.Body.Bytes(), &suggestions); err != nil {
		t.Fatalf("failed to decode suggestions: %v", err)
	}
	found := false
	for _, s := range suggestions {
		if strings.Contains(strings.ToLower(s), "algo") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected algorithm suggestion, got: %v", suggestions)
	}
}

func TestRFCByThreadEndpoint(t *testing.T) {
	router := NewRouter()

	// List seeded threads for the Treewiki page.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/talk/threads?page_key=0:Treewiki", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing threads, got %d", rec.Code)
	}
	var threads []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &threads); err != nil || len(threads) == 0 {
		t.Fatalf("expected seeded threads, got: %s", rec.Body.String())
	}
	threadID, _ := threads[0]["id"].(string)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/talk/threads/"+threadID+"/rfc", nil)
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for thread RFC, got %d", rec2.Code)
	}
	var rfc map[string]any
	if err := json.Unmarshal(rec2.Body.Bytes(), &rfc); err != nil {
		t.Fatalf("failed to decode rfc: %v", err)
	}
	if rfc["question"] == "" {
		t.Errorf("expected RFC question, got: %v", rfc)
	}
}
