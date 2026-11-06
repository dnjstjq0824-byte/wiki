// Package middleware provides production HTTP middleware for the Treewiki
// API gateway: request IDs, structured access logging, panic recovery,
// security headers, configurable CORS, body size limits, Prometheus-style
// metrics, and an in-memory per-IP rate limiter.
package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/sentry"
)

type contextKey string

const requestIDKey = contextKey("treewiki_request_id")

// RequestIDFromContext returns the request ID attached by RequestID middleware.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// RequestID attaches a unique ID to every request and response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// Logger emits one structured JSON access-log line per request.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		entry := map[string]any{
			"ts":         start.Format(time.RFC3339Nano),
			"level":      "info",
			"request_id": RequestIDFromContext(r.Context()),
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"status":     rec.status,
			"bytes":      rec.bytes,
			"latency_ms": time.Since(start).Milliseconds(),
			"remote_ip":  clientIP(r),
		}
		line, _ := json.Marshal(entry)
		log.Println(string(line))
		metricsInstance.observe(rec.status, time.Since(start))
	})
}

// Recovery converts panics into JSON 500 responses instead of killing the process.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID := RequestIDFromContext(r.Context())
				log.Printf(`{"level":"error","request_id":%q,"msg":"panic recovered","panic":%q,"path":%q}`,
					requestID, toString(rec), r.URL.Path)
				reportPanic(toString(rec), requestID, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// reportPanic forwards recovered panics to Sentry (no-op without SENTRY_DSN).
func reportPanic(value, requestID, path string) {
	client := sentryClient()
	if client == nil {
		return
	}
	go func() {
		_ = client.CaptureException(value, requestID, path)
	}()
}

var (
	sentryOnce   sync.Once
	sentryCached *sentry.Client
)

func sentryClient() *sentry.Client {
	sentryOnce.Do(func() {
		sentryCached = sentry.NewFromEnv()
	})
	return sentryCached
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case error:
		return t.Error()
	default:
		return "unknown panic value"
	}
}

// SecurityHeaders sets baseline response headers (CSP v3 compatible with Nuxt SSR).
func SecurityHeaders(next http.Handler) http.Handler {
	isProd := os.Getenv("ENV") == "production" || os.Getenv("NODE_ENV") == "production"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		// API serves JSON only; lock down framing and object embedding.
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		if isProd {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}
		next.ServeHTTP(w, r)
	})
}

// CORSConfig controls allowed origins for cross-origin requests.
type CORSConfig struct {
	AllowedOrigins []string
	AllowAll       bool
}

// CORSConfigFromEnv builds CORS config from CORS_ALLOWED_ORIGINS (comma-separated).
// Empty value defaults to local dev origins; "*" allows all (non-production only).
func CORSConfigFromEnv() CORSConfig {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "*" {
		return CORSConfig{AllowAll: true}
	}
	if raw == "" {
		return CORSConfig{AllowedOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"}}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if o := strings.TrimSpace(p); o != "" {
			origins = append(origins, o)
		}
	}
	return CORSConfig{AllowedOrigins: origins}
}

// CORS enforces an origin allowlist and handles preflight requests.
func CORS(cfg CORSConfig, next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range cfg.AllowedOrigins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if cfg.AllowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, If-Match, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// BodyLimit caps request body size to mitigate oversized-payload abuse.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil && r.ContentLength > maxBytes {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "request body too large"})
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter is a per-key fixed-window rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	max     int
	window  time.Duration
	limited uint64
}

// NewRateLimiter creates a limiter allowing max requests per window per client IP.
// Non-positive max disables limiting (all requests pass).
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{hits: make(map[string][]time.Time), max: max, window: window}
}

// NewRateLimiterFromEnv builds a limiter from RATE_LIMIT_MAX / RATE_LIMIT_WINDOW_SECONDS.
// Defaults: 300 requests / 60s per IP; RATE_LIMIT_MAX<=0 disables.
func NewRateLimiterFromEnv() *RateLimiter {
	max := 300
	windowSec := 60
	if v := strings.TrimSpace(os.Getenv("RATE_LIMIT_MAX")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			max = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("RATE_LIMIT_WINDOW_SECONDS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			windowSec = n
		}
	}
	return NewRateLimiter(max, time.Duration(windowSec)*time.Second)
}

// Middleware rejects over-limit requests with 429 + Retry-After.
func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l == nil || l.max <= 0 {
			next.ServeHTTP(w, r)
			return
		}
		key := clientIP(r)
		now := time.Now().UTC()

		l.mu.Lock()
		cutoff := now.Add(-l.window)
		kept := l.hits[key][:0]
		for _, t := range l.hits[key] {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		if len(kept) >= l.max {
			l.limited++
			l.hits[key] = kept
			l.mu.Unlock()
			metricsInstance.incRateLimited()
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(int(l.window.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded, retry later"})
			return
		}
		l.hits[key] = append(kept, now)
		l.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// LimitedCount returns total rejected requests (for /metrics).
func (l *RateLimiter) LimitedCount() uint64 {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.limited
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if i := strings.Index(fwd, ","); i >= 0 {
			return strings.TrimSpace(fwd[:i])
		}
		return strings.TrimSpace(fwd)
	}
	if host := strings.TrimSpace(r.RemoteAddr); host != "" {
		if i := strings.LastIndex(host, ":"); i > 0 {
			return host[:i]
		}
		return host
	}
	return "unknown"
}

// Chain applies middlewares outside-in: Chain(h, A, B) == A(B(h)).
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
