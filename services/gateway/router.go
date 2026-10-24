package gateway

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dnjstjqppk-dev/treewiki/pkg/parser"
)

// HealthResponse represents system health metadata
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// ParseRequest represents a wikitext parsing payload
type ParseRequest struct {
	Wikitext string `json:"wikitext"`
}

// NewRouter sets up HTTP routes for Treewiki API Gateway
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "OK",
			Version:   "1.0.0",
			Timestamp: time.Now().UTC(),
		})
	})

	// Readiness probe
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "READY"})
	})

	// Wikitext parser API
	mux.HandleFunc("POST /api/v1/parser/preview", func(w http.ResponseWriter, r *http.Request) {
		var req ParseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}

		res := parser.ParseWikitext(req.Wikitext)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, If-Match")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
