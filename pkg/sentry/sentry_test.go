package sentry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewFromDSNParsing(t *testing.T) {
	if NewFromDSN("") != nil {
		t.Fatal("empty DSN must yield nil client")
	}
	if NewFromDSN("not-a-dsn") != nil {
		t.Fatal("garbage DSN must yield nil client")
	}
	c := NewFromDSN("https://publickey123@sentry.example.com/42")
	if c == nil {
		t.Fatal("valid DSN must yield client")
	}
	if c.endpoint != "https://sentry.example.com/api/42/store/" {
		t.Fatalf("bad endpoint: %s", c.endpoint)
	}
	if c.publicKey != "publickey123" {
		t.Fatalf("bad key: %s", c.publicKey)
	}
}

func TestCaptureExceptionDeliversPayload(t *testing.T) {
	var gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("X-Sentry-Auth")
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dsn := strings.Replace(srv.URL, "http://", "http://testkey@", 1) + "/7"
	c := NewFromDSN(dsn)
	if c == nil {
		t.Fatal("test DSN must parse")
	}
	if err := c.CaptureException("boom", "req-1", "/api/v1/pages"); err != nil {
		t.Fatalf("capture: %v", err)
	}
	if !strings.Contains(gotAuth, "sentry_key=testkey") {
		t.Fatalf("missing auth header: %q", gotAuth)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["level"] != "error" {
		t.Fatalf("level wrong: %v", payload["level"])
	}
	if !strings.Contains(gotBody, "boom") || !strings.Contains(gotBody, "req-1") {
		t.Fatalf("payload lacks context: %s", gotBody)
	}
}

func TestNilClientIsNoOp(t *testing.T) {
	var c *Client
	if err := c.CaptureException("x", "y", "z"); err != nil {
		t.Fatalf("nil client must no-op, got %v", err)
	}
}
