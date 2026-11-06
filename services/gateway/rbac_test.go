package gateway

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dnjstjqppk-dev/treewiki/pkg/auth"
)

func testJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "treewiki_dev_secret_change_me"
}

func bearerFor(t *testing.T, userID, username, role string) string {
	t.Helper()
	pair, err := auth.GenerateTokenPair(userID, username, role, false, testJWTSecret())
	if err != nil {
		t.Fatalf("mint token: %v", err)
	}
	return "Bearer " + pair.AccessToken
}

func TestGovernanceMutationsRequireSysop(t *testing.T) {
	router := NewRouter()
	body := `{"target_user_id":"spammer-1","reason":"test block"}`

	// Anonymous → 403.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/governance/blocks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("anonymous block must be 403, got %d", rec.Code)
	}

	// Registered user → 403.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/governance/blocks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerFor(t, "user-1", "RegularJoe", auth.RoleRegistered))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("registered block must be 403, got %d", rec.Code)
	}

	// Sysop → success.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/governance/blocks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerFor(t, "sysop-1", "SysopJane", auth.RoleSysop))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("sysop block must succeed, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestTalkModerationRequiresSysop(t *testing.T) {
	router := NewRouter()

	threadsReq := httptest.NewRequest(http.MethodGet, "/api/v1/talk/threads?page_key=0:Treewiki", nil)
	threadsRec := httptest.NewRecorder()
	router.ServeHTTP(threadsRec, threadsReq)
	if threadsRec.Code != http.StatusOK {
		t.Fatalf("seed threads missing: %d", threadsRec.Code)
	}

	// Pin with a registered token → 403 (use a bogus id; auth runs first).
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/talk/threads/bogus/pin", bytes.NewBufferString(`{"pinned":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerFor(t, "user-2", "RegularJoe", auth.RoleRegistered))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("registered pin must be 403, got %d", rec.Code)
	}
}

func TestTakedownFilingStaysOpenToMembers(t *testing.T) {
	router := NewRouter()
	body := `{"page_title":"Treewiki","reason":"test request","requester":"RegularJoe"}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/governance/takedowns", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerFor(t, "user-3", "RegularJoe", auth.RoleRegistered))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("member takedown filing must succeed, got %d %s", rec.Code, rec.Body.String())
	}
}
