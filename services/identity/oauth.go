package identity

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dnjstjqppk-dev/treewiki/pkg/auth"
)

var (
	ErrOAuthProviderUnknown = errors.New("unknown oauth provider")
	ErrOAuthNotConfigured   = errors.New("oauth provider is not configured")
	ErrOAuthStateInvalid    = errors.New("invalid or expired oauth state")
	ErrOAuthExchange        = errors.New("oauth code exchange failed")
	ErrOAuthProfile         = errors.New("oauth profile fetch failed")
	ErrOAuthCodeInvalid     = errors.New("invalid or expired oauth code")
	ErrOAuthLinkTaken       = errors.New("oauth identity is already linked to another account")
	ErrOAuthEmailRequired   = errors.New("oauth provider did not return a verified email address")
)

const (
	oauthStateTTL = 10 * time.Minute
	oauthCodeTTL  = 5 * time.Minute
)

// OAuthProvider describes one authorization-code social login provider.
type OAuthProvider struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	EmailsURL    string // optional fallback when UserInfoURL omits email (GitHub)
	Scopes       string
}

// OAuthProfile is the normalized provider identity.
type OAuthProfile struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
	Name           string
	AvatarURL      string
}

// OAuthProviderInfo is returned by the providers listing endpoint.
type OAuthProviderInfo struct {
	Name       string `json:"name"`
	Configured bool   `json:"configured"`
}

type oauthStatePending struct {
	provider  string
	expiresAt time.Time
}

type oauthCodePending struct {
	userID    string
	expiresAt time.Time
}

// ConfigureOAuthFromEnv loads social-login providers from environment variables.
// Supported: GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET, GITHUB_CLIENT_ID/GITHUB_CLIENT_SECRET.
// OAUTH_REDIRECT_BASE is the public gateway base URL used for callback URLs
// (default http://localhost:8080). FRONTEND_URL is where the browser is sent
// after the provider callback (default http://localhost:3000).
func (s *IdentityService) ConfigureOAuthFromEnv() {
	s.ConfigureOAuth(
		envOr("FRONTEND_URL", "http://localhost:3000"),
		envOr("OAUTH_REDIRECT_BASE", "http://localhost:8080"),
		OAuthProvider{
			Name:         "google",
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://openidconnect.googleapis.com/v1/userinfo",
			Scopes:       "openid email profile",
		},
		OAuthProvider{
			Name:         "github",
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
			EmailsURL:    "https://api.github.com/user/emails",
			Scopes:       "read:user user:email",
		},
	)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// ConfigureOAuth installs providers explicitly (used by tests and wiring).
func (s *IdentityService) ConfigureOAuth(frontendURL, redirectBase string, providers ...OAuthProvider) {
	s.oauthMu.Lock()
	defer s.oauthMu.Unlock()
	if s.oauthProviders == nil {
		s.oauthProviders = make(map[string]*OAuthProvider)
	}
	if s.oauthStates == nil {
		s.oauthStates = make(map[string]oauthStatePending)
	}
	if s.oauthCodes == nil {
		s.oauthCodes = make(map[string]oauthCodePending)
	}
	if s.oauthHTTP == nil {
		s.oauthHTTP = &http.Client{Timeout: 10 * time.Second}
	}
	s.oauthFrontendURL = frontendURL
	s.oauthRedirectBase = strings.TrimSuffix(redirectBase, "/")
	for i := range providers {
		p := providers[i]
		s.oauthProviders[p.Name] = &p
	}
}

// OAuthProviders lists known providers and whether each is configured.
func (s *IdentityService) OAuthProviders() []OAuthProviderInfo {
	s.oauthMu.Lock()
	defer s.oauthMu.Unlock()
	out := make([]OAuthProviderInfo, 0, len(s.oauthProviders))
	for _, p := range s.oauthProviders {
		out = append(out, OAuthProviderInfo{
			Name:       p.Name,
			Configured: p.ClientID != "" && p.ClientSecret != "",
		})
	}
	// Stable order: google, github, then anything else alphabetically.
	ordered := make([]OAuthProviderInfo, 0, len(out))
	for _, want := range []string{"google", "github"} {
		for _, info := range out {
			if info.Name == want {
				ordered = append(ordered, info)
			}
		}
	}
	for _, info := range out {
		if info.Name != "google" && info.Name != "github" {
			ordered = append(ordered, info)
		}
	}
	return ordered
}

// OAuthFrontendURL returns the configured browser base URL.
func (s *IdentityService) OAuthFrontendURL() string {
	s.oauthMu.Lock()
	defer s.oauthMu.Unlock()
	if s.oauthFrontendURL == "" {
		return "http://localhost:3000"
	}
	return s.oauthFrontendURL
}

func (s *IdentityService) oauthProviderLocked(name string) (*OAuthProvider, error) {
	p, ok := s.oauthProviders[strings.ToLower(name)]
	if !ok || p == nil {
		return nil, ErrOAuthProviderUnknown
	}
	if p.ClientID == "" || p.ClientSecret == "" {
		return nil, ErrOAuthNotConfigured
	}
	return p, nil
}

func (s *IdentityService) oauthCallbackURL(p *OAuthProvider) string {
	return s.oauthRedirectBase + "/api/v1/auth/oauth/" + p.Name + "/callback"
}

// BeginOAuthLogin creates a single-use state and returns the provider
// authorization URL the browser must visit.
func (s *IdentityService) BeginOAuthLogin(provider string) (string, error) {
	s.oauthMu.Lock()
	defer s.oauthMu.Unlock()
	p, err := s.oauthProviderLocked(provider)
	if err != nil {
		return "", err
	}
	s.purgeOAuthLocked(time.Now().UTC())
	state, err := oauthRandomToken()
	if err != nil {
		return "", err
	}
	s.oauthStates[state] = oauthStatePending{provider: p.Name, expiresAt: time.Now().UTC().Add(oauthStateTTL)}
	q := url.Values{}
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", s.oauthCallbackURL(p))
	q.Set("response_type", "code")
	q.Set("scope", p.Scopes)
	q.Set("state", state)
	return p.AuthURL + "?" + q.Encode(), nil
}

// CompleteOAuthLogin validates the state, exchanges the code, fetches the
// profile, finds or creates the local user, and returns a single-use code the
// frontend exchanges for tokens via ConsumeOAuthCode.
func (s *IdentityService) CompleteOAuthLogin(provider, code, state string) (string, error) {
	s.oauthMu.Lock()
	p, err := s.oauthProviderLocked(provider)
	if err != nil {
		s.oauthMu.Unlock()
		return "", err
	}
	now := time.Now().UTC()
	s.purgeOAuthLocked(now)
	pending, ok := s.oauthStates[state]
	if !ok || pending.provider != p.Name {
		s.oauthMu.Unlock()
		return "", ErrOAuthStateInvalid
	}
	delete(s.oauthStates, state)
	redirectURI := s.oauthCallbackURL(p)
	s.oauthMu.Unlock()

	accessToken, err := s.oauthExchangeCode(p, code, redirectURI)
	if err != nil {
		return "", err
	}
	profile, err := s.oauthFetchProfile(p, accessToken)
	if err != nil {
		return "", err
	}

	user, err := s.oauthFindOrCreateUser(p.Name, profile)
	if err != nil {
		return "", err
	}
	if user.IsBlocked {
		return "", ErrAccountBlocked
	}

	oneTime, err := oauthRandomToken()
	if err != nil {
		return "", err
	}
	s.oauthMu.Lock()
	s.oauthCodes[oneTime] = oauthCodePending{userID: user.ID, expiresAt: time.Now().UTC().Add(oauthCodeTTL)}
	s.oauthMu.Unlock()
	return oneTime, nil
}

// ConsumeOAuthCode exchanges a single-use login code for a token pair.
func (s *IdentityService) ConsumeOAuthCode(code string) (*AuthResponse, error) {
	s.oauthMu.Lock()
	s.purgeOAuthLocked(time.Now().UTC())
	pending, ok := s.oauthCodes[code]
	if !ok {
		s.oauthMu.Unlock()
		return nil, ErrOAuthCodeInvalid
	}
	delete(s.oauthCodes, code)
	s.oauthMu.Unlock()

	user, err := s.store.GetByID(pending.userID)
	if err != nil {
		return nil, ErrOAuthCodeInvalid
	}
	if user.IsBlocked {
		return nil, ErrAccountBlocked
	}
	effectiveRole := auth.EffectiveRole(user.Role, user.CreatedAt, user.EditCount)
	tokens, err := auth.GenerateTokenPair(user.ID, user.Username, effectiveRole, user.IsBot, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{
		User:         toUserPublic(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *IdentityService) purgeOAuthLocked(now time.Time) {
	for k, v := range s.oauthStates {
		if !now.Before(v.expiresAt) {
			delete(s.oauthStates, k)
		}
	}
	for k, v := range s.oauthCodes {
		if !now.Before(v.expiresAt) {
			delete(s.oauthCodes, k)
		}
	}
}

func oauthRandomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *IdentityService) oauthExchangeCode(p *OAuthProvider, code, redirectURI string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)

	req, err := http.NewRequest(http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.oauthHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: provider status %d", ErrOAuthExchange, resp.StatusCode)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("%w: %s", ErrOAuthExchange, payload.Error)
	}
	return payload.AccessToken, nil
}

func (s *IdentityService) oauthFetchProfile(p *OAuthProvider, accessToken string) (*OAuthProfile, error) {
	body, err := s.oauthGetJSON(p.UserInfoURL, accessToken)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthProfile, err)
	}
	profile := &OAuthProfile{}
	switch p.Name {
	case "google":
		profile.ProviderUserID = oauthString(raw["sub"])
		profile.Email = strings.TrimSpace(oauthString(raw["email"]))
		profile.Name = strings.TrimSpace(oauthString(raw["name"]))
		profile.AvatarURL = strings.TrimSpace(oauthString(raw["picture"]))
		if v, ok := raw["email_verified"].(bool); ok {
			profile.EmailVerified = v
		}
	case "github":
		profile.ProviderUserID = oauthString(raw["id"])
		if profile.ProviderUserID == "" {
			if f, ok := raw["id"].(float64); ok {
				profile.ProviderUserID = strings.TrimSuffix(fmt.Sprintf("%f", f), ".000000")
			}
		}
		profile.Email = strings.TrimSpace(oauthString(raw["email"]))
		profile.Name = strings.TrimSpace(oauthString(raw["name"]))
		if profile.Name == "" {
			profile.Name = strings.TrimSpace(oauthString(raw["login"]))
		}
		profile.AvatarURL = strings.TrimSpace(oauthString(raw["avatar_url"]))
		if profile.Email != "" {
			profile.EmailVerified = true
		}
		if p.EmailsURL != "" && (!profile.EmailVerified || profile.Email == "") {
			if email, verified := s.oauthGitHubEmail(p, accessToken); email != "" {
				profile.Email = email
				profile.EmailVerified = verified
			}
		}
	default:
		profile.ProviderUserID = oauthString(raw["id"])
		if profile.ProviderUserID == "" {
			profile.ProviderUserID = oauthString(raw["sub"])
		}
		profile.Email = strings.TrimSpace(oauthString(raw["email"]))
		profile.Name = strings.TrimSpace(oauthString(raw["name"]))
		profile.AvatarURL = strings.TrimSpace(oauthString(raw["avatar_url"]))
		if profile.AvatarURL == "" {
			profile.AvatarURL = strings.TrimSpace(oauthString(raw["picture"]))
		}
		if v, ok := raw["email_verified"].(bool); ok {
			profile.EmailVerified = v
		}
	}
	if profile.ProviderUserID == "" {
		return nil, ErrOAuthProfile
	}
	if profile.Email == "" || !profile.EmailVerified {
		return nil, ErrOAuthEmailRequired
	}
	return profile, nil
}

func (s *IdentityService) oauthGitHubEmail(p *OAuthProvider, accessToken string) (string, bool) {
	body, err := s.oauthGetJSON(p.EmailsURL, accessToken)
	if err != nil {
		return "", false
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", false
	}
	fallback := ""
	for _, e := range emails {
		if !e.Verified || e.Email == "" {
			continue
		}
		if e.Primary {
			return e.Email, true
		}
		if fallback == "" {
			fallback = e.Email
		}
	}
	return fallback, fallback != ""
}

func (s *IdentityService) oauthGetJSON(endpoint, accessToken string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthProfile, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := s.oauthHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthProfile, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: provider status %d", ErrOAuthProfile, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

func oauthString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strings.TrimSuffix(fmt.Sprintf("%f", t), ".000000")
	default:
		return ""
	}
}

// oauthFindOrCreateUser links the provider identity to an existing account
// (by prior link, then by verified email) or creates a new account.
func (s *IdentityService) oauthFindOrCreateUser(provider string, profile *OAuthProfile) (*User, error) {
	if user, err := s.store.GetByOAuthLink(provider, profile.ProviderUserID); err == nil {
		return user, nil
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	if linked, err := s.store.GetByEmail(profile.Email); err == nil {
		if err := s.store.CreateOAuthLink(linked.ID, provider, profile.ProviderUserID); err != nil && !errors.Is(err, ErrOAuthLinkTaken) {
			return nil, err
		}
		return linked, nil
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	username := s.oauthUniqueUsername(deriveOAuthUsername(profile.Name, profile.Email))
	password := "Oa0!" + uuid.NewString() + "Zz9"
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	email := profile.Email
	user := &User{
		ID:            uuid.New().String(),
		Username:      username,
		Email:         &email,
		PasswordHash:  hash,
		Role:          auth.RoleRegistered,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if profile.AvatarURL != "" {
		user.AvatarURL = profile.AvatarURL
	}
	if err := s.store.Create(user); err != nil {
		return nil, err
	}
	if err := s.store.CreateOAuthLink(user.ID, provider, profile.ProviderUserID); err != nil {
		return nil, err
	}
	return user, nil
}

var oauthUsernameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)

// deriveOAuthUsername maps a provider name/email to a valid local username base.
func deriveOAuthUsername(name, email string) string {
	base := strings.TrimSpace(name)
	if base == "" && email != "" {
		if at := strings.Index(email, "@"); at > 0 {
			base = email[:at]
		} else {
			base = email
		}
	}
	if base == "" {
		base = "user"
	}
	base = strings.ReplaceAll(base, " ", ".")
	base = oauthUsernameSanitizer.ReplaceAllString(base, "_")
	base = strings.Trim(base, ".-_")
	if len(base) < 3 {
		base = "user-" + base
	}
	if len(base) < 3 {
		base = "user"
	}
	if len(base) > 28 {
		base = base[:28]
	}
	return base
}

func (s *IdentityService) oauthUniqueUsername(base string) string {
	candidate := base
	for i := 2; ; i++ {
		if _, err := s.store.GetByUsername(candidate); errors.Is(err, ErrUserNotFound) {
			return candidate
		}
		suffix := fmt.Sprintf("-%d", i)
		trimmed := base
		if len(trimmed)+len(suffix) > 32 {
			trimmed = trimmed[:32-len(suffix)]
		}
		candidate = trimmed + suffix
	}
}
