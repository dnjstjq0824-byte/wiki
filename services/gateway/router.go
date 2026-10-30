package gateway

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/dnjstjqppk-dev/treewiki/pkg/parser"
	"github.com/dnjstjqppk-dev/treewiki/services/governance"
	"github.com/dnjstjqppk-dev/treewiki/services/identity"
	"github.com/dnjstjqppk-dev/treewiki/services/search"
	"github.com/dnjstjqppk-dev/treewiki/services/talk"
	"github.com/dnjstjqppk-dev/treewiki/services/topictree"
	"github.com/dnjstjqppk-dev/treewiki/services/watchlist"
	"github.com/dnjstjqppk-dev/treewiki/services/wiki"
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

	// Identity service wiring
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "treewiki_dev_secret_change_me"
	}
	store := identity.NewMemoryUserStore()
	svc := identity.NewIdentityService(store, jwtSecret)
	h := identity.NewHandler(svc)

	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("POST /api/v1/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /api/v1/auth/resend-verification", identity.RequireAuth(h.ResendEmailVerification))
	mux.HandleFunc("GET /api/v1/auth/me", identity.RequireAuth(h.Me))
	mux.HandleFunc("PUT /api/v1/users/me/profile", identity.RequireAuth(h.UpdateProfile))
	mux.HandleFunc("PUT /api/v1/auth/password", identity.RequireAuth(h.ChangePassword))
	mux.HandleFunc("POST /api/v1/auth/totp/setup", identity.RequireAuth(h.SetupTOTP))
	mux.HandleFunc("POST /api/v1/auth/totp/verify", identity.RequireAuth(h.VerifyTOTP))
	mux.HandleFunc("POST /api/v1/auth/anonymize", identity.RequireAuth(h.AnonymizeAccount))
	mux.HandleFunc("GET /api/v1/users/{username}", h.GetPublicProfile)

	// Wiki service wiring
	wikiStore := wiki.NewPageStore()
	wikiSvc := wiki.NewWikiService(wikiStore)
	wikiH := wiki.NewHandler(wikiSvc)

	mux.HandleFunc("GET /api/v1/namespaces", wikiH.Namespaces)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/pages", wikiH.ListPages)
	mux.HandleFunc("GET /api/v1/pages/{namespace}/{slug}", wikiH.GetPage)
	mux.HandleFunc("PUT /api/v1/pages/{namespace}/{slug}", identity.RequireAuth(wikiH.PutPage))
	mux.HandleFunc("DELETE /api/v1/pages/{namespace}/{slug}", identity.RequireAuth(wikiH.DeletePage))
	mux.HandleFunc("POST /api/v1/pages/{namespace}/{slug}/move", identity.RequireAuth(wikiH.MovePage))
	mux.HandleFunc("POST /api/v1/pages/{namespace}/{slug}/undelete", identity.RequireAuth(wikiH.UndeletePage))
	mux.HandleFunc("GET /api/v1/pages/{namespace}/{slug}/backlinks", wikiH.Backlinks)
	mux.HandleFunc("GET /api/v1/pages/{namespace}/{slug}/history", wikiH.Revisions)
	mux.HandleFunc("GET /api/v1/pages/{namespace}/{slug}/resolve", wikiH.ResolvePage)
	mux.HandleFunc("PUT /api/v1/pages/{namespace}/{slug}/protection", identity.RequireAuth(wikiH.SetProtection))
	mux.HandleFunc("POST /api/v1/pages/{namespace}/{slug}/revert", identity.RequireAuth(wikiH.Revert))
	mux.HandleFunc("POST /api/v1/pages/{namespace}/{slug}/rollback", identity.RequireAuth(wikiH.Rollback))
	mux.HandleFunc("GET /api/v1/categories/{category}", wikiH.CategoryMembers)
	mux.HandleFunc("GET /api/v1/pages/{namespace}/{slug}/diff", wikiH.Diff)
	// Topic tree service wiring
	topicStore := topictree.NewTopicStore()
	topicSvc := topictree.NewTopicService(topicStore)
	topicH := topictree.NewHandler(topicSvc)

	mux.HandleFunc("GET /api/v1/topics", topicH.ListNodes)
	mux.HandleFunc("POST /api/v1/topics", identity.RequireAuth(topicH.CreateNode))
	mux.HandleFunc("GET /api/v1/topics/{id}", topicH.GetNode)
	mux.HandleFunc("PUT /api/v1/topics/{id}", identity.RequireAuth(topicH.UpdateNode))
	mux.HandleFunc("DELETE /api/v1/topics/{id}", identity.RequireAuth(topicH.DeleteNode))
	mux.HandleFunc("POST /api/v1/topics/link", identity.RequireAuth(topicH.Link))
	mux.HandleFunc("POST /api/v1/topics/unlink", identity.RequireAuth(topicH.Unlink))
	mux.HandleFunc("GET /api/v1/topics/{id}/ancestors", topicH.Ancestors)
	mux.HandleFunc("GET /api/v1/topics/{id}/descendants", topicH.Descendants)
	mux.HandleFunc("GET /api/v1/topics/export", topicH.Export)
	mux.HandleFunc("GET /api/v1/learning-paths", topicH.ListPaths)
	mux.HandleFunc("POST /api/v1/learning-paths", identity.RequireAuth(topicH.CreatePath))
	mux.HandleFunc("GET /api/v1/learning-paths/{id}", topicH.GetPath)
	// Talk service wiring
	talkStore := talk.NewTalkStore()
	talkSvc := talk.NewTalkService(talkStore)
	talkH := talk.NewHandler(talkSvc)

	mux.HandleFunc("GET /api/v1/talk/threads", talkH.ListThreads)
	mux.HandleFunc("POST /api/v1/talk/threads", identity.RequireAuth(talkH.CreateThread))
	mux.HandleFunc("PUT /api/v1/talk/threads/{id}/status", identity.RequireAuth(talkH.SetStatus))
	mux.HandleFunc("PUT /api/v1/talk/threads/{id}/pin", identity.RequireAuth(talkH.PinThread))
	mux.HandleFunc("PUT /api/v1/talk/threads/{id}/lock", identity.RequireAuth(talkH.LockThread))
	mux.HandleFunc("GET /api/v1/talk/threads/{id}/comments", talkH.GetComments)
	mux.HandleFunc("POST /api/v1/talk/threads/{id}/comments", identity.RequireAuth(talkH.AddComment))
	mux.HandleFunc("POST /api/v1/talk/rfc", identity.RequireAuth(talkH.CreateRFC))
	mux.HandleFunc("GET /api/v1/talk/rfc/{id}", talkH.GetRFC)
	mux.HandleFunc("POST /api/v1/talk/rfc/{id}/vote", identity.RequireAuth(talkH.VoteRFC))
	mux.HandleFunc("POST /api/v1/talk/rfc/{id}/close", identity.RequireAuth(talkH.CloseRFC))
	mux.HandleFunc("POST /api/v1/talk/threads/{id}/archive", identity.RequireAuth(talkH.ArchiveThread))
	mux.HandleFunc("GET /api/v1/talk/archived", talkH.ListArchived)
	mux.HandleFunc("POST /api/v1/talk/arbitration", identity.RequireAuth(talkH.CreateArbitration))
	mux.HandleFunc("GET /api/v1/talk/arbitration", talkH.ListArbitrations)
	mux.HandleFunc("GET /api/v1/talk/sse", talkH.SSE)
	// Search service wiring
	searchIndex := search.NewSearchIndex()
	mediaStore := search.NewMediaStore()
	searchH := search.NewHandler(searchIndex, mediaStore)

	mux.HandleFunc("GET /api/v1/search", searchH.Search)
	mux.HandleFunc("GET /api/v1/search/autocomplete", searchH.Autocomplete)
	mux.HandleFunc("GET /api/v1/media", searchH.ListMedia)
	mux.HandleFunc("POST /api/v1/media", identity.RequireAuth(searchH.UploadMedia))
	// Governance service wiring
	govStore := governance.NewGovernanceStore()
	govSvc := governance.NewGovernanceService(govStore)
	govH := governance.NewHandler(govSvc)

	mux.HandleFunc("GET /api/v1/governance/filters", govH.ListFilters)
	mux.HandleFunc("POST /api/v1/governance/filters", identity.RequireAuth(govH.AddFilter))
	mux.HandleFunc("POST /api/v1/governance/evaluate", govH.Evaluate)
	mux.HandleFunc("GET /api/v1/governance/blocks", govH.ListBlocks)
	mux.HandleFunc("POST /api/v1/governance/blocks", identity.RequireAuth(govH.AddBlock))
	mux.HandleFunc("GET /api/v1/governance/oversight", govH.ListOversight)
	mux.HandleFunc("POST /api/v1/governance/oversight", identity.RequireAuth(govH.AddOversight))
	mux.HandleFunc("GET /api/v1/governance/takedowns", govH.ListTakedowns)
	mux.HandleFunc("POST /api/v1/governance/takedowns", identity.RequireAuth(govH.AddTakedown))
	mux.HandleFunc("PUT /api/v1/governance/takedowns/{id}", identity.RequireAuth(govH.UpdateTakedownStatus))
	mux.HandleFunc("GET /api/v1/governance/dump", identity.RequireAuth(govH.Dump))
	mux.HandleFunc("GET /api/v1/recentchanges", govH.RecentChanges)
	// Watchlist service wiring
	wlStore := watchlist.NewWatchlistStore()
	wlSvc := watchlist.NewService(wlStore)
	wlH := watchlist.NewHandler(wlSvc)

	mux.HandleFunc("GET /api/v1/watchlist", wlH.List)
	mux.HandleFunc("POST /api/v1/watchlist", identity.RequireAuth(wlH.Add))
	mux.HandleFunc("DELETE /api/v1/watchlist", identity.RequireAuth(wlH.Remove))
	authed := identity.AuthMiddleware(jwtSecret)(mux)
	return corsMiddleware(authed)
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

