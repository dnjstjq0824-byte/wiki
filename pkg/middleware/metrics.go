package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// collector aggregates process-level HTTP metrics without external dependencies.
type collector struct {
	mu            sync.Mutex
	requestsTotal uint64
	byStatus      map[int]uint64
	rateLimited   uint64
	latencySumMs  int64
	startedAt     time.Time
}

var metricsInstance = &collector{byStatus: make(map[int]uint64), startedAt: time.Now().UTC()}

func (c *collector) observe(status int, latency time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestsTotal++
	c.byStatus[status]++
	c.latencySumMs += latency.Milliseconds()
}

func (c *collector) incRateLimited() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rateLimited++
}

// Handler exposes Prometheus exposition format at GET /metrics.
func (c *collector) Handler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintln(w, "# HELP treewiki_http_requests_total Total HTTP requests served.")
	fmt.Fprintln(w, "# TYPE treewiki_http_requests_total counter")
	fmt.Fprintf(w, "treewiki_http_requests_total %d\n", c.requestsTotal)
	fmt.Fprintln(w, "# HELP treewiki_http_requests_by_status HTTP requests by status code.")
	fmt.Fprintln(w, "# TYPE treewiki_http_requests_by_status counter")
	for status, n := range c.byStatus {
		fmt.Fprintf(w, "treewiki_http_requests_by_status{status=\"%d\"} %d\n", status, n)
	}
	fmt.Fprintln(w, "# HELP treewiki_http_rate_limited_total Requests rejected by rate limiter.")
	fmt.Fprintln(w, "# TYPE treewiki_http_rate_limited_total counter")
	fmt.Fprintf(w, "treewiki_http_rate_limited_total %d\n", c.rateLimited)
	avg := 0.0
	if c.requestsTotal > 0 {
		avg = float64(c.latencySumMs) / float64(c.requestsTotal)
	}
	fmt.Fprintln(w, "# HELP treewiki_http_latency_avg_ms Average request latency in milliseconds.")
	fmt.Fprintln(w, "# TYPE treewiki_http_latency_avg_ms gauge")
	fmt.Fprintf(w, "treewiki_http_latency_avg_ms %.2f\n", avg)
	fmt.Fprintln(w, "# HELP treewiki_process_uptime_seconds Process uptime in seconds.")
	fmt.Fprintln(w, "# TYPE treewiki_process_uptime_seconds gauge")
	fmt.Fprintf(w, "treewiki_process_uptime_seconds %.0f\n", time.Since(c.startedAt).Seconds())
	fmt.Fprintln(w, "# HELP treewiki_process_goroutines Current goroutine count.")
	fmt.Fprintln(w, "# TYPE treewiki_process_goroutines gauge")
	fmt.Fprintf(w, "treewiki_process_goroutines %d\n", runtime.NumGoroutine())
}

// MetricsHandler returns the shared /metrics handler.
func MetricsHandler() http.HandlerFunc {
	return metricsInstance.Handler
}
