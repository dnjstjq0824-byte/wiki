package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestRequestIDGeneratesAndPropagates(t *testing.T) {
	var got string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}
	if got == "" || got != rec.Header().Get("X-Request-ID") {
		t.Fatalf("context request id %q must match header", got)
	}
}

func TestRequestIDHonorsClientValue(t *testing.T) {
	h := RequestID(okHandler())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "client-123")
	h.ServeHTTP(rec, req)
	if rec.Header().Get("X-Request-ID") != "client-123" {
		t.Fatalf("expected echoed request id, got %q", rec.Header().Get("X-Request-ID"))
	}
}

func TestRecoveryConvertsPanicTo500(t *testing.T) {
	panicker := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	Recovery(panicker).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON body: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected error message in body")
	}
}

func TestSecurityHeadersPresent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(okHandler()).ServeHTTP(rec, req)

	for _, h := range []string{
		"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy",
		"Permissions-Policy", "Content-Security-Policy",
	} {
		if rec.Header().Get(h) == "" {
			t.Errorf("missing security header %s", h)
		}
	}
}

func TestCORSAllowlistedOrigin(t *testing.T) {
	cfg := CORSConfig{AllowedOrigins: []string{"http://localhost:3000"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pages/0/X", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	CORS(cfg, okHandler()).ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatal("expected allowlisted origin to be reflected")
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("expected credentials header for allowlisted origin")
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	cfg := CORSConfig{AllowedOrigins: []string{"http://localhost:3000"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example")
	CORS(cfg, okHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "https://evil.example" {
		t.Fatal("must not reflect unlisted origin")
	}
}

func TestCORSPreflight(t *testing.T) {
	cfg := CORSConfig{AllowedOrigins: []string{"http://localhost:3000"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	CORS(cfg, okHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 preflight, got %d", rec.Code)
	}
}

func TestBodyLimitRejectsOversized(t *testing.T) {
	h := BodyLimit(16)(okHandler())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 64)))
	req.ContentLength = 64
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestRateLimiterAllowsThenRejects(t *testing.T) {
	l := NewRateLimiter(3, time.Minute)
	h := l.Middleware(okHandler())

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.9:1234"
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d should pass, got %d", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:1234"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header on 429")
	}
	if l.LimitedCount() != 1 {
		t.Fatalf("expected 1 limited request, got %d", l.LimitedCount())
	}
}

func TestRateLimiterIsolatesClients(t *testing.T) {
	l := NewRateLimiter(1, time.Minute)
	h := l.Middleware(okHandler())

	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.1:1"
	h.ServeHTTP(rec1, req1)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.2:1"
	h.ServeHTTP(rec2, req2)

	if rec1.Code != http.StatusOK || rec2.Code != http.StatusOK {
		t.Fatalf("different IPs must have independent buckets: %d, %d", rec1.Code, rec2.Code)
	}
}

func TestRateLimiterDisabledPasses(t *testing.T) {
	l := NewRateLimiter(0, time.Minute)
	h := l.Middleware(okHandler())
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("disabled limiter must pass, got %d", rec.Code)
		}
	}
}

func TestMetricsHandlerExposesCounters(t *testing.T) {
	c := &collector{byStatus: map[int]uint64{}, startedAt: time.Now().UTC()}
	c.observe(200, 5*time.Millisecond)
	c.observe(404, 2*time.Millisecond)
	c.incRateLimited()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	c.Handler(rec, req)

	body := rec.Body.String()
	for _, want := range []string{
		"treewiki_http_requests_total 2",
		`treewiki_http_requests_by_status{status="200"} 1`,
		"treewiki_http_rate_limited_total 1",
		"treewiki_process_goroutines",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q\n%s", want, body)
		}
	}
}
