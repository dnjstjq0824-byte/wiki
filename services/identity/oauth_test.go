package identity

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dnjstjqppk-dev/treewiki/pkg/auth"
)

func fakeOAuthServer(userinfo, emails string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code") == "bad-code" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-` + r.Form.Get("code") + `"}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer tok-") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(userinfo))
	})
	mux.HandleFunc("/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(emails))
	})
	return httptest.NewServer(mux)
}

func oauthTestService(t *testing.T, srv *httptest.Server, userinfoKind string) *IdentityService {
	t.Helper()
	svc := NewIdentityService(NewMemoryUserStore(), "test-secret")
	provider := OAuthProvider{
		Name:         userinfoKind,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		AuthURL:      srv.URL + "/auth",
		TokenURL:     srv.URL + "/token",
		UserInfoURL:  srv.URL + "/userinfo",
		Scopes:       "test",
	}
	if userinfoKind == "github" {
		provider.EmailsURL = srv.URL + "/emails"
	}
	svc.ConfigureOAuth("http://localhost:3000", "http://localhost:8080", provider,
		OAuthProvider{Name: "other", AuthURL: srv.URL + "/auth", TokenURL: srv.URL + "/token", UserInfoURL: srv.URL + "/userinfo"})
	return svc
}

const googleUserinfo = `{"sub":"google-123","email":"curator@example.com","email_verified":true,"name":"Tree Curator","picture":"https://example.com/avatar.png"}`

func beginState(t *testing.T, svc *IdentityService, provider string) (authURL, state string) {
	t.Helper()
	authURL, err := svc.BeginOAuthLogin(provider)
	if err != nil {
		t.Fatalf("BeginOAuthLogin: %v", err)
	}
	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}
	if u.Query().Get("client_id") == "" || u.Query().Get("redirect_uri") == "" {
		t.Fatalf("auth url missing params: %s", authURL)
	}
	state = u.Query().Get("state")
	if state == "" {
		t.Fatal("auth url missing state")
	}
	return authURL, state
}

func TestOAuthProvidersListing(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	infos := svc.OAuthProviders()
	byName := map[string]bool{}
	for _, info := range infos {
		byName[info.Name] = info.Configured
	}
	if !byName["google"] {
		t.Fatal("google must be configured")
	}
	if byName["other"] {
		t.Fatal("provider without credentials must be unconfigured")
	}
	if _, err := svc.BeginOAuthLogin("other"); err != ErrOAuthNotConfigured {
		t.Fatalf("expected ErrOAuthNotConfigured, got %v", err)
	}
	if _, err := svc.BeginOAuthLogin("nope"); err != ErrOAuthProviderUnknown {
		t.Fatalf("expected ErrOAuthProviderUnknown, got %v", err)
	}
}

func TestOAuthGoogleFullFlow(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	authURL, state := beginState(t, svc, "google")
	if !strings.HasPrefix(authURL, srv.URL+"/auth?") {
		t.Fatalf("unexpected auth url: %s", authURL)
	}

	code, err := svc.CompleteOAuthLogin("google", "good-code", state)
	if err != nil {
		t.Fatalf("CompleteOAuthLogin: %v", err)
	}
	if code == "" {
		t.Fatal("one-time code must not be empty")
	}

	res, err := svc.ConsumeOAuthCode(code)
	if err != nil {
		t.Fatalf("ConsumeOAuthCode: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("token pair must be issued")
	}
	if !auth.ValidateUsername(res.User.Username) {
		t.Fatalf("derived username invalid: %q", res.User.Username)
	}
	storedEmail, err := svc.store.GetByID(res.User.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if storedEmail.Email == nil || *storedEmail.Email != "curator@example.com" {
		t.Fatalf("email mismatch: %+v", storedEmail.Email)
	}

	stored, err := svc.store.GetByOAuthLink("google", "google-123")
	if err != nil {
		t.Fatalf("GetByOAuthLink: %v", err)
	}
	if stored.ID != res.User.ID || !stored.EmailVerified {
		t.Fatalf("link mismatch: %+v", stored)
	}

	// Second login with the same provider identity must reuse the account.
	_, state2 := beginState(t, svc, "google")
	code2, err := svc.CompleteOAuthLogin("google", "good-code-2", state2)
	if err != nil {
		t.Fatalf("second CompleteOAuthLogin: %v", err)
	}
	res2, err := svc.ConsumeOAuthCode(code2)
	if err != nil {
		t.Fatalf("second ConsumeOAuthCode: %v", err)
	}
	if res2.User.ID != res.User.ID {
		t.Fatal("repeat OAuth login must return the same user")
	}
}

func TestOAuthLinksExistingEmailAccount(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	local, err := svc.Register(RegisterRequest{Username: "localcurator", Email: "curator@example.com", Password: "Str0ngPass!x"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, state := beginState(t, svc, "google")
	code, err := svc.CompleteOAuthLogin("google", "link-code", state)
	if err != nil {
		t.Fatalf("CompleteOAuthLogin: %v", err)
	}
	res, err := svc.ConsumeOAuthCode(code)
	if err != nil {
		t.Fatalf("ConsumeOAuthCode: %v", err)
	}
	if res.User.ID != local.User.ID {
		t.Fatal("OAuth with matching email must link the existing account")
	}
	if _, err := svc.store.GetByOAuthLink("google", "google-123"); err != nil {
		t.Fatalf("link must be stored: %v", err)
	}
}

func TestOAuthStateIsSingleUse(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	_, state := beginState(t, svc, "google")
	if _, err := svc.CompleteOAuthLogin("google", "code-a", state); err != nil {
		t.Fatalf("first complete: %v", err)
	}
	if _, err := svc.CompleteOAuthLogin("google", "code-b", state); err != ErrOAuthStateInvalid {
		t.Fatalf("expected ErrOAuthStateInvalid on reuse, got %v", err)
	}
	if _, err := svc.CompleteOAuthLogin("google", "code-c", "bogus-state"); err != ErrOAuthStateInvalid {
		t.Fatalf("expected ErrOAuthStateInvalid for bogus state, got %v", err)
	}
}

func TestOAuthExchangeFailure(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	_, state := beginState(t, svc, "google")
	if _, err := svc.CompleteOAuthLogin("google", "bad-code", state); err == nil || !strings.Contains(err.Error(), ErrOAuthExchange.Error()) {
		t.Fatalf("expected exchange error, got %v", err)
	}
}

func TestOAuthGitHubEmailsFallback(t *testing.T) {
	userinfo := `{"id":4242,"login":"octo-cat","email":null,"name":null,"avatar_url":"https://example.com/octo.png"}`
	emails := `[{"email":"unverified@example.com","primary":false,"verified":false},{"email":"octo@example.com","primary":true,"verified":true}]`
	srv := fakeOAuthServer(userinfo, emails)
	defer srv.Close()
	svc := oauthTestService(t, srv, "github")

	_, state := beginState(t, svc, "github")
	code, err := svc.CompleteOAuthLogin("github", "gh-code", state)
	if err != nil {
		t.Fatalf("CompleteOAuthLogin: %v", err)
	}
	res, err := svc.ConsumeOAuthCode(code)
	if err != nil {
		t.Fatalf("ConsumeOAuthCode: %v", err)
	}
	storedGH, err := svc.store.GetByID(res.User.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if storedGH.Email == nil || *storedGH.Email != "octo@example.com" {
		t.Fatalf("must use verified primary email, got %+v", storedGH.Email)
	}
	if !auth.ValidateUsername(res.User.Username) {
		t.Fatalf("derived username invalid: %q", res.User.Username)
	}
}

func TestOAuthUnverifiedEmailRejected(t *testing.T) {
	userinfo := `{"sub":"google-999","email":"shady@example.com","email_verified":false,"name":"Shady"}`
	srv := fakeOAuthServer(userinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	_, state := beginState(t, svc, "google")
	if _, err := svc.CompleteOAuthLogin("google", "code-x", state); err != ErrOAuthEmailRequired {
		t.Fatalf("expected ErrOAuthEmailRequired, got %v", err)
	}
}

func TestOAuthConsumeSingleUse(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	_, state := beginState(t, svc, "google")
	code, err := svc.CompleteOAuthLogin("google", "code-1", state)
	if err != nil {
		t.Fatalf("CompleteOAuthLogin: %v", err)
	}
	if _, err := svc.ConsumeOAuthCode(code); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if _, err := svc.ConsumeOAuthCode(code); err != ErrOAuthCodeInvalid {
		t.Fatalf("expected ErrOAuthCodeInvalid on reuse, got %v", err)
	}
	if _, err := svc.ConsumeOAuthCode("bogus"); err != ErrOAuthCodeInvalid {
		t.Fatalf("expected ErrOAuthCodeInvalid for bogus code, got %v", err)
	}
}

func TestOAuthCodeExpiry(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	_, state := beginState(t, svc, "google")
	code, err := svc.CompleteOAuthLogin("google", "code-e", state)
	if err != nil {
		t.Fatalf("CompleteOAuthLogin: %v", err)
	}
	svc.oauthMu.Lock()
	pending := svc.oauthCodes[code]
	pending.expiresAt = time.Now().UTC().Add(-time.Minute)
	svc.oauthCodes[code] = pending
	svc.oauthMu.Unlock()

	if _, err := svc.ConsumeOAuthCode(code); err != ErrOAuthCodeInvalid {
		t.Fatalf("expected ErrOAuthCodeInvalid for expired code, got %v", err)
	}
}

func TestOAuthBlockedUser(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")

	_, state := beginState(t, svc, "google")
	code, err := svc.CompleteOAuthLogin("google", "code-b1", state)
	if err != nil {
		t.Fatalf("CompleteOAuthLogin: %v", err)
	}
	res, err := svc.ConsumeOAuthCode(code)
	if err != nil {
		t.Fatalf("ConsumeOAuthCode: %v", err)
	}
	stored, _ := svc.store.GetByID(res.User.ID)
	stored.IsBlocked = true
	if err := svc.store.Update(stored); err != nil {
		t.Fatalf("Update: %v", err)
	}

	_, state2 := beginState(t, svc, "google")
	if _, err := svc.CompleteOAuthLogin("google", "code-b2", state2); err != ErrAccountBlocked {
		t.Fatalf("expected ErrAccountBlocked, got %v", err)
	}
}

func TestDeriveOAuthUsername(t *testing.T) {
	cases := []struct{ name, email string }{
		{"Tree Curator", "curator@example.com"},
		{"", "curator@example.com"},
		{"한글 이름 !@#", "user@example.com"},
		{"ab", "ab@example.com"},
		{"", ""},
		{"a-very-long-display-name-that-exceeds-thirty-two-chars", "long@example.com"},
	}
	for _, c := range cases {
		got := deriveOAuthUsername(c.name, c.email)
		if !auth.ValidateUsername(got) {
			t.Fatalf("deriveOAuthUsername(%q,%q) = %q, invalid", c.name, c.email, got)
		}
	}
}

func TestOAuthHTTPHandlers(t *testing.T) {
	srv := fakeOAuthServer(googleUserinfo, `[]`)
	defer srv.Close()
	svc := oauthTestService(t, srv, "google")
	h := NewHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/auth/oauth/providers", h.OAuthProviders)
	mux.HandleFunc("GET /api/v1/auth/oauth/{provider}/login", h.OAuthLogin)
	mux.HandleFunc("GET /api/v1/auth/oauth/{provider}/callback", h.OAuthCallback)
	mux.HandleFunc("POST /api/v1/auth/oauth/consume", h.OAuthConsume)

	// Providers listing.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/providers", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("providers status = %d", rec.Code)
	}
	var listing struct {
		Providers []OAuthProviderInfo `json:"providers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listing); err != nil {
		t.Fatalf("decode listing: %v", err)
	}
	if len(listing.Providers) == 0 {
		t.Fatal("providers listing must not be empty")
	}

	// Login redirects to the provider.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/google/login", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("login status = %d", rec.Code)
	}
	loc, _ := url.Parse(rec.Header().Get("Location"))
	state := loc.Query().Get("state")
	if state == "" {
		t.Fatal("login redirect must carry state")
	}

	// Unknown provider login is 404.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/nope/login", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown provider status = %d", rec.Code)
	}

	// Callback in JSON mode returns the one-time code.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/google/callback?code=cb-code&state="+state+"&format=json", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("callback status = %d: %s", rec.Code, rec.Body.String())
	}
	var cb struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cb); err != nil || cb.Code == "" {
		t.Fatalf("callback must return code: %s", rec.Body.String())
	}

	// Consume returns the token pair and sets the session cookie.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/consume", strings.NewReader(`{"code":"`+cb.Code+`"}`))
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("consume status = %d: %s", rec.Code, rec.Body.String())
	}
	var authRes AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &authRes); err != nil || authRes.AccessToken == "" {
		t.Fatalf("consume must return tokens: %s", rec.Body.String())
	}
	if rec.Header().Get("Set-Cookie") == "" {
		t.Fatal("consume must set the session cookie")
	}

	// Browser callback redirects to the frontend with the code.
	_, state2 := beginState(t, svc, "google")
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/google/callback?code=cb-2&state="+state2, nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("browser callback status = %d", rec.Code)
	}
	if !strings.HasPrefix(rec.Header().Get("Location"), "http://localhost:3000/auth/callback?") {
		t.Fatalf("unexpected callback target: %s", rec.Header().Get("Location"))
	}
}
