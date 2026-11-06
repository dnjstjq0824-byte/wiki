// Package sentry provides a dependency-free Sentry-compatible error reporter.
//
// When SENTRY_DSN is set, panics recovered by the gateway are delivered to
// the Sentry store API in a fire-and-forget goroutine with a short timeout.
// When unset, the client is nil and reporting is a no-op.
package sentry

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client delivers events to one Sentry project.
type Client struct {
	endpoint   string
	publicKey  string
	httpClient *http.Client
}

// NewFromEnv builds a client from SENTRY_DSN, or nil when unset/invalid.
func NewFromEnv() *Client {
	return NewFromDSN(os.Getenv("SENTRY_DSN"))
}

// NewFromDSN parses "https://publicKey@host[:port]/projectId".
func NewFromDSN(dsn string) *Client {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil
	}
	u, err := url.Parse(dsn)
	if err != nil || u.User == nil || u.Host == "" {
		return nil
	}
	publicKey := u.User.Username()
	project := strings.Trim(u.Path, "/")
	if publicKey == "" || project == "" || strings.Contains(project, "/") {
		return nil
	}
	scheme := u.Scheme
	if scheme != "https" && scheme != "http" {
		return nil
	}
	return &Client{
		endpoint:   fmt.Sprintf("%s://%s/api/%s/store/", scheme, u.Host, project),
		publicKey:  publicKey,
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

func newEventID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// CaptureException sends a panic/error event synchronously (callers deciding
// latency should wrap it in a goroutine).
func (c *Client) CaptureException(value, requestID, path string) error {
	if c == nil {
		return nil
	}
	payload := map[string]any{
		"event_id":  newEventID(),
		"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05"),
		"level":     "error",
		"logger":    "treewiki-gateway",
		"exception": map[string]any{
			"values": []map[string]any{
				{"type": "panic", "value": value},
			},
		},
		"request": map[string]any{"url": path},
		"tags":    map[string]any{"request_id": requestID},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sentry-Auth",
		"Sentry sentry_version=7, sentry_key="+c.publicKey+", sentry_client=treewiki-go/1.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sentry store returned %d", resp.StatusCode)
	}
	return nil
}
