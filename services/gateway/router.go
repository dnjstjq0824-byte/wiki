package gateway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dnjstjqppk-dev/treewiki/pkg/auth"
	"github.com/dnjstjqppk-dev/treewiki/pkg/db"
	"github.com/dnjstjqppk-dev/treewiki/pkg/middleware"
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
	Database  string    `json:"database"`
}

// ParseRequest represents a wikitext parsing payload
type ParseRequest struct {
	Wikitext string `json:"wikitext"`
}

// NewRouter sets up HTTP routes for Treewiki API Gateway
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	dbStatus := "IN_MEMORY"
	var userStore identity.UserStore = identity.NewMemoryUserStore()
	var wikiStore wiki.PageStorage = wiki.NewPageStore()
	var topicStore topictree.TopicStorage = topictree.NewTopicStore()
	var talkStore talk.TalkStorage = talk.NewTalkStore()
	var searchBackend search.SearchBackend = search.NewSearchIndex()
	var mediaStore search.MediaStorage = search.NewMediaStore()
	var govStore governance.GovernanceStorage = governance.NewGovernanceStore()
	var wlStore watchlist.WatchlistStorage = watchlist.NewWatchlistStore()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		if pgDB, err := db.Connect(dbURL); err == nil {
			if err := pgDB.MigrateAll("database/migrations"); err != nil {
				log.Printf("[Warning] migration failed: %v\n", err)
			}
			userStore = identity.NewPostgresUserStore(pgDB.Pool)
			wikiStore = wiki.NewPostgresPageStore(pgDB.Pool)
			topicStore = topictree.NewPostgresTopicStore(pgDB.Pool)
			talkStore = talk.NewPostgresTalkStore(pgDB.Pool)
			searchBackend = search.NewPostgresSearchBackend(pgDB.Pool)
			mediaStore = search.NewPostgresMediaStore(pgDB.Pool)
			govStore = governance.NewPostgresGovernanceStore(pgDB.Pool)
			wlStore = watchlist.NewPostgresWatchlistStore(pgDB.Pool)
			dbStatus = "POSTGRESQL_CONNECTED"
		} else {
			log.Printf("[Warning] PostgreSQL connection failed, falling back to In-Memory: %v\n", err)
		}
	}

	// Health check
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "OK",
			Version:   appVersion(),
			Timestamp: time.Now().UTC(),
			Database:  dbStatus,
		})
	})

	// Readiness probe
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "READY", "database": dbStatus})
	})

	// Prometheus metrics
	mux.HandleFunc("GET /metrics", middleware.MetricsHandler())

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
	if isProduction() && (jwtSecret == "" || jwtSecret == "treewiki_dev_secret_change_me") {
		log.Fatal("[Treewiki Gateway] JWT_SECRET must be set to a strong value in production")
	}
	svc := identity.NewIdentityService(userStore, jwtSecret)
	svc.ConfigureOAuthFromEnv()
	h := identity.NewHandler(svc)

	mux.HandleFunc("GET /api/v1/auth/oauth/providers", h.OAuthProviders)
	mux.HandleFunc("GET /api/v1/auth/oauth/{provider}/login", h.OAuthLogin)
	mux.HandleFunc("GET /api/v1/auth/oauth/{provider}/callback", h.OAuthCallback)
	mux.HandleFunc("POST /api/v1/auth/oauth/consume", h.OAuthConsume)
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

	// Service wiring (memory or PostgreSQL stores, selected above)
	wikiSvc := wiki.NewWikiService(wikiStore)
	wikiH := wiki.NewHandler(wikiSvc)

	topicSvc := topictree.NewTopicService(topicStore)
	topicH := topictree.NewHandler(topicSvc)

	talkSvc := talk.NewTalkService(talkStore)
	talkH := talk.NewHandler(talkSvc)

	searchH := search.NewHandler(searchBackend, mediaStore)

	govSvc := governance.NewGovernanceService(govStore)
	govH := governance.NewHandler(govSvc)

	wlSvc := watchlist.NewService(wlStore)
	wlH := watchlist.NewHandler(wlSvc)

	// Link wiki updates to search index and governance activity log
	wikiH.SetOnPageSave(func(p *wiki.Page, isUpdate bool) {
		searchBackend.Index(&search.SearchDocument{
			ID:        p.ID,
			Namespace: int(p.Namespace),
			Title:     p.Title,
			Content:   p.Content,
		})
		action := "Created"
		if isUpdate {
			action = "Edited"
		}
		summary := p.Summary
		if summary == "" {
			summary = "Updated page content"
		}
		govSvc.RecordRecentChange(fmt.Sprintf(`{"title":"%s","namespace":%d,"author":"%s","summary":"%s","action":"%s","timestamp":"%s"}`,
			p.Title, p.Namespace, p.AuthorID, summary, action, time.Now().UTC().Format(time.RFC3339)))
	})

	// Seed foundational default data
	seedDefaultData(wikiSvc, topicSvc, searchBackend, talkSvc, govSvc)

	mux.HandleFunc("GET /api/v1/namespaces", wikiH.Namespaces)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/pages", wikiH.ListPages)
	mux.HandleFunc("GET /api/v1/pages/{namespace}/{slug}", wikiH.GetPage)
	mux.HandleFunc("PUT /api/v1/pages/{namespace}/{slug}", wikiH.PutPage)
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
	mux.HandleFunc("GET /api/v1/pages/random", wikiH.RandomPage)

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

	mux.HandleFunc("GET /api/v1/talk/threads", talkH.ListThreads)
	mux.HandleFunc("POST /api/v1/talk/threads", talkH.CreateThread)
	mux.HandleFunc("PUT /api/v1/talk/threads/{id}/status", identity.RequireRole(auth.RoleSysop, talkH.SetStatus))
	mux.HandleFunc("PUT /api/v1/talk/threads/{id}/pin", identity.RequireRole(auth.RoleSysop, talkH.PinThread))
	mux.HandleFunc("PUT /api/v1/talk/threads/{id}/lock", identity.RequireRole(auth.RoleSysop, talkH.LockThread))
	mux.HandleFunc("GET /api/v1/talk/threads/{id}/comments", talkH.GetComments)
	mux.HandleFunc("POST /api/v1/talk/threads/{id}/comments", talkH.AddComment)
	mux.HandleFunc("POST /api/v1/talk/rfc", identity.RequireAuth(talkH.CreateRFC))
	mux.HandleFunc("GET /api/v1/talk/rfc/{id}", talkH.GetRFC)
	mux.HandleFunc("GET /api/v1/talk/threads/{id}/rfc", talkH.GetRFCByThread)
	mux.HandleFunc("POST /api/v1/talk/rfc/{id}/vote", talkH.VoteRFC)
	mux.HandleFunc("POST /api/v1/talk/rfc/{id}/close", identity.RequireRole(auth.RoleSysop, talkH.CloseRFC))
	mux.HandleFunc("POST /api/v1/talk/threads/{id}/archive", identity.RequireRole(auth.RoleSysop, talkH.ArchiveThread))
	mux.HandleFunc("GET /api/v1/talk/archived", talkH.ListArchived)
	mux.HandleFunc("POST /api/v1/talk/arbitration", identity.RequireAuth(talkH.CreateArbitration))
	mux.HandleFunc("GET /api/v1/talk/arbitration", talkH.ListArbitrations)
	mux.HandleFunc("GET /api/v1/talk/sse", talkH.SSE)

	mux.HandleFunc("GET /api/v1/search", searchH.Search)
	mux.HandleFunc("GET /api/v1/search/autocomplete", searchH.Autocomplete)
	mux.HandleFunc("GET /api/v1/media", searchH.ListMedia)
	mux.HandleFunc("POST /api/v1/media", identity.RequireAuth(searchH.UploadMedia))

	mux.HandleFunc("GET /api/v1/governance/filters", govH.ListFilters)
	mux.HandleFunc("POST /api/v1/governance/filters", identity.RequireRole(auth.RoleSysop, govH.AddFilter))
	mux.HandleFunc("POST /api/v1/governance/evaluate", govH.Evaluate)
	mux.HandleFunc("GET /api/v1/governance/blocks", govH.ListBlocks)
	mux.HandleFunc("POST /api/v1/governance/blocks", identity.RequireRole(auth.RoleSysop, govH.AddBlock))
	mux.HandleFunc("GET /api/v1/governance/oversight", govH.ListOversight)
	mux.HandleFunc("POST /api/v1/governance/oversight", identity.RequireRole(auth.RoleOversight, govH.AddOversight))
	mux.HandleFunc("GET /api/v1/governance/takedowns", govH.ListTakedowns)
	mux.HandleFunc("POST /api/v1/governance/takedowns", identity.RequireAuth(govH.AddTakedown))
	mux.HandleFunc("PUT /api/v1/governance/takedowns/{id}", identity.RequireRole(auth.RoleSysop, govH.UpdateTakedownStatus))
	mux.HandleFunc("GET /api/v1/governance/dump", identity.RequireRole(auth.RoleSysop, govH.Dump))
	mux.HandleFunc("GET /api/v1/recentchanges", govH.RecentChanges)

	mux.HandleFunc("GET /api/v1/watchlist", wlH.List)
	mux.HandleFunc("POST /api/v1/watchlist", identity.RequireAuth(wlH.Add))
	mux.HandleFunc("DELETE /api/v1/watchlist", identity.RequireAuth(wlH.Remove))

	// Platform statistics for portals and dashboards
	mux.HandleFunc("GET /api/v1/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"version":   appVersion(),
			"database":  dbStatus,
			"pages":     wikiSvc.CountAll(),
			"topics":    len(topicSvc.AllNodes()),
			"timestamp": time.Now().UTC(),
		})
	})

	authed := identity.AuthMiddleware(jwtSecret)(mux)
	limiter := middleware.NewRateLimiterFromEnv()
	return middleware.Chain(
		authed,
		middleware.RequestID,
		middleware.Recovery,
		middleware.Logger,
		middleware.SecurityHeaders,
		middleware.BodyLimit(40<<20),
		limiter.Middleware,
		func(next http.Handler) http.Handler {
			return middleware.CORS(middleware.CORSConfigFromEnv(), next)
		},
	)
}

func isProduction() bool {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("NODE_ENV")
	}
	return env == "production"
}

func appVersion() string {
	if v := strings.TrimSpace(os.Getenv("APP_VERSION")); v != "" {
		return v
	}
	return "1.0.0"
}

func seedDefaultData(
	wikiSvc *wiki.WikiService,
	topicSvc *topictree.TopicService,
	searchIdx search.SearchBackend,
	talkSvc *talk.TalkService,
	govSvc *governance.GovernanceService,
) {
	// 1. Foundational Encyclopedic Pages
	pages := []struct {
		title   string
		summary string
		content string
	}{
		{
			title:   "Treewiki",
			summary: "Initial Treewiki landing documentation",
			content: `= Welcome to Treewiki =

'''Treewiki''' is a next-generation collaborative knowledge platform combining a **Myers Diff** version-controlled wiki engine with a **Directed Acyclic Graph (DAG)** knowledge taxonomy.

== Core Architectural Pillars ==
* **Hybrid Lexer & Parser**: Seamlessly transforms MediaWiki syntax alongside GitHub Flavored Markdown and mathematical KaTeX equations.
* **Knowledge DAG Taxonomy**: Multi-parent prerequisite graphs replacing rigid Wikipedia categories with clear learning trajectories.
* **Decentralized Consensus**: Threaded talk pages with formal [[RFC]] consensus polling, quorum thresholds, and dispute arbitration.
* **Cryptographic Ledger**: SHA-256 revision hash chaining with three-way merge conflict resolution.

== Explore Foundational Topics ==
* [[Computer Science]] — The mathematical study of computation and systemic information.
* [[Data Structures]] — Fundamental memory layouts, binary search trees, and DAGs.
* [[Algorithms]] — Rigorous step-by-step methods and asymptotic complexity.
* [[Graph Theory]] — Mathematical foundations of vertices, edges, and topological orderings.
* [[Myers Diff]] — The gold standard shortest edit script diff algorithm.

[[Category:Documentation]]
[[Category:Computer Science]]`,
		},
		{
			title:   "Computer_Science",
			summary: "Overview of Computer Science discipline",
			content: `= Computer Science =

'''Computer Science''' is the study of computation, information processing, and algorithmic systems. It spans theoretical disciplines to applied fields including software engineering, artificial intelligence, and distributed systems.

== Foundational Subfields ==
* [[Algorithms]] — Asymptotic complexity, sorting algorithms, and shortest path graph search.
* [[Data Structures]] — Dynamic arrays, hash tables, B-trees, and Directed Acyclic Graphs.
* [[Graph Theory]] — Graph representations, adjacency matrices, and topological sorting.
* [[Myers Diff]] — The sequence comparison algorithm powering modern version control.

[[Category:Computer Science]]`,
		},
		{
			title:   "Algorithms",
			summary: "Core algorithms and computational complexity",
			content: `= Algorithms =

An '''algorithm''' is an unambiguous, finite specification of how to solve a class of problems. Algorithms perform calculation, data processing, and automated reasoning tasks.

== Fundamental Classification ==
* **Divide and Conquer**: Merge Sort, Fast Fourier Transform (FFT).
* **Dynamic Programming**: Bellman-Ford, Floyd-Warshall, Knapsack.
* **Greedy Heuristics**: Dijkstra's algorithm, Kruskal's MST, [[Myers Diff]].

== Asymptotic Analysis ==
In algorithm analysis, Big-O notation characterizes performance in terms of input growth $N$:
* $O(1)$: Constant time hash map lookup.
* $O(\log N)$: Binary search on ordered arrays.
* $O(N \log N)$: Optimal comparison sort (Merge Sort, Heapsort).
* $O(N \cdot D)$: Myers Diff algorithm where $D$ is edit distance.

[[Category:Algorithms]]
[[Category:Computer Science]]`,
		},
		{
			title:   "Data_Structures",
			summary: "Taxonomy of data structures",
			content: `= Data Structures =

A '''data structure''' provides specialized formatting for organizing, processing, retrieving, and storing data.

== Linear vs Hierarchical Primitives ==
* **Linear**: Arrays, Doubly Linked Lists, Stacks (LIFO), and Queues (FIFO).
* **Trees**: AVL Trees, Red-Black Trees, Heaps, and B-Trees.
* **Graphs**: Directed Acyclic Graphs (DAG), powering Treewiki's topic taxonomy.

[[Category:Data Structures]]
[[Category:Computer Science]]`,
		},
		{
			title:   "Graph_Theory",
			summary: "Foundations of graph theory and DAGs",
			content: `= Graph Theory =

'''Graph Theory''' is the mathematical study of graphs — structures used to model pairwise relations between objects. A graph $G = (V, E)$ consists of vertices $V$ connected by edges $E$.

== Directed Acyclic Graphs (DAG) ==
A **Directed Acyclic Graph** has oriented edges with no directed cycles. In Treewiki, topic trees are modeled as DAGs to allow topics to have multiple prerequisites without circular deadlocks.

[[Category:Graph Theory]]
[[Category:Mathematics]]`,
		},
		{
			title:   "Myers_Diff",
			summary: "Myers diff algorithm specifications",
			content: `= Myers Diff Algorithm =

The '''Myers Diff Algorithm''' was introduced by Eugene W. Myers in 1986. It calculates the Shortest Edit Script (SES) between two sequences using graph search over an edit grid.

== Algorithm Complexity ==
* Time: $O(ND)$ where $N = |A| + |B|$ and $D$ is the length of the minimal edit script.
* Space: $O(N + D^2)$, reducible to linear $O(N)$ with divide-and-conquer recursion.

Treewiki utilizes Myers diff to compute token-level inline and side-by-side revision deltas.

[[Category:Algorithms]]
[[Category:Diff]]`,
		},
	}

	for _, p := range pages {
		if page, err := wikiSvc.CreatePage(wiki.NSMain, p.title, p.content, p.summary, "SystemArchitect", "sysop"); err == nil {
			searchIdx.Index(&search.SearchDocument{
				ID:        page.ID,
				Namespace: int(page.Namespace),
				Title:     page.Title,
				Content:   page.Content,
			})
			govSvc.RecordRecentChange(fmt.Sprintf(`{"title":"%s","namespace":%d,"author":"%s","summary":"%s","action":"Created","timestamp":"%s"}`,
				page.Title, page.Namespace, page.AuthorID, page.Summary, time.Now().UTC().Format(time.RFC3339)))
		}
	}

	// 2. Topic DAG nodes & edges
	_ = topicSvc.AddNode(&topictree.TopicNode{
		ID:          "cs",
		Name:        "Computer Science",
		Description: "Foundational theory of computation and information",
		Difficulty:  1,
	})
	_ = topicSvc.AddNode(&topictree.TopicNode{
		ID:          "algo",
		Name:        "Algorithms",
		Description: "Finite instructions and computational complexity",
		Difficulty:  2,
	})
	_ = topicSvc.AddNode(&topictree.TopicNode{
		ID:          "ds",
		Name:        "Data Structures",
		Description: "Memory primitives and data organization formats",
		Difficulty:  2,
	})
	_ = topicSvc.AddNode(&topictree.TopicNode{
		ID:          "graph",
		Name:        "Graph Theory",
		Description: "Vertices, edges, and Directed Acyclic Graphs",
		Difficulty:  3,
	})
	_ = topicSvc.AddNode(&topictree.TopicNode{
		ID:          "diff",
		Name:        "Myers Diff",
		Description: "Shortest edit script algorithm for version control",
		Difficulty:  4,
	})
	_ = topicSvc.AddNode(&topictree.TopicNode{
		ID:          "ai",
		Name:        "Artificial Intelligence",
		Description: "Autonomous agents, neural networks, and generative models",
		Difficulty:  3,
	})

	_ = topicSvc.Link("cs", "algo")
	_ = topicSvc.Link("cs", "ds")
	_ = topicSvc.Link("cs", "ai")
	_ = topicSvc.Link("algo", "graph")
	_ = topicSvc.Link("ds", "graph")
	_ = topicSvc.Link("algo", "diff")

	// 3. Discussion Thread & RFC on Treewiki page (seed once; the guard keeps
	// PostgreSQL restarts from duplicating threads).
	if len(talkSvc.ListThreads("0:Treewiki")) == 0 {
		if thread, err := talkSvc.CreateThread("0:Treewiki", "RFC-1: Standardizing Wikitext and Markdown Interop", "CuratorAlpha"); err == nil {
			_, _ = talkSvc.AddComment(thread.ID, "", "CuratorAlpha", "Welcome to Treewiki! We invite the community to review the hybrid parser specification. Do you agree with prioritising CommonMark for tabular data while keeping Wikitext for templates?")
			_, _ = talkSvc.AddComment(thread.ID, "", "FoundingEditor", "Strong support. CommonMark tables are much easier to read and edit across mobile and desktop devices.")
			if rfc, err := talkSvc.CreateRFC(thread.ID, "Should CommonMark syntax be standardized for tabular and code-fence data?", 3); err == nil {
				_ = talkSvc.VoteRFC(rfc.ID, "FoundingEditor", "SUPPORT")
				_ = talkSvc.VoteRFC(rfc.ID, "CuratorAlpha", "SUPPORT")
			}
		}
	}
}
