package identity

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dnjstjqppk-dev/treewiki/pkg/auth"
)

func setupTestApp() (*Handler, *IdentityService, string) {
	jwtSecret := "super_secret_test_jwt_key_9876543210"
	store := NewMemoryUserStore()
	service := NewIdentityService(store, jwtSecret)
	handler := NewHandler(service)
	return handler, service, jwtSecret
}

func TestRegistrationAndLoginFlow(t *testing.T) {
	handler, _, jwtSecret := setupTestApp()

	// 1. Register User
	regPayload := `{"username":"CuratorAlpha","email":"curator@treewiki.org","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(regPayload))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var authRes AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &authRes); err != nil {
		t.Fatalf("failed to decode auth response: %v", err)
	}

	if authRes.User.Username != "CuratorAlpha" {
		t.Errorf("expected username CuratorAlpha, got %s", authRes.User.Username)
	}

	// 2. Reject Duplicate Registration
	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(regPayload))
	recDup := httptest.NewRecorder()
	handler.Register(recDup, reqDup)

	if recDup.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for duplicate registration, got %d", recDup.Code)
	}

	// 3. Login User
	loginPayload := `{"username":"CuratorAlpha","password":"Password123!"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginPayload))
	recLogin := httptest.NewRecorder()
	handler.Login(recLogin, reqLogin)

	if recLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for login, got %d", recLogin.Code)
	}

	// Check cookie
	cookies := recLogin.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "treewiki_session" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("expected session cookie to be set")
	}

	// Verify token validity
	claims, err := auth.ValidateToken(sessionCookie.Value, jwtSecret)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}

	if claims.Username != "CuratorAlpha" {
		t.Errorf("claims mismatch: %s", claims.Username)
	}
}

func TestRBACMiddlewareEnforcement(t *testing.T) {
	_, service, jwtSecret := setupTestApp()

	// Register normal user
	authRes, err := service.Register(RegisterRequest{
		Username: "NormalUser",
		Email:    "normal@treewiki.org",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Sysop-protected endpoint mock
	protectedHandler := RequireRole(auth.RoleSysop, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access":"granted"}`))
	})

	fullStack := AuthMiddleware(jwtSecret)(protectedHandler)

	// Test 1: Anonymous request should be rejected with 403 (or 401 if RequireAuth used)
	reqAnon := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	recAnon := httptest.NewRecorder()
	fullStack.ServeHTTP(recAnon, reqAnon)

	if recAnon.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for anonymous on sysop route, got %d", recAnon.Code)
	}

	// Test 2: Registered user (below Sysop) should be rejected with 403
	reqUser := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	reqUser.Header.Set("Authorization", "Bearer "+authRes.AccessToken)
	recUser := httptest.NewRecorder()
	fullStack.ServeHTTP(recUser, reqUser)

	if recUser.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for normal user on sysop route, got %d", recUser.Code)
	}

	// Test 3: Sysop user should succeed
	sysopTokens, _ := auth.GenerateTokenPair(authRes.User.ID, authRes.User.Username, auth.RoleSysop, false, jwtSecret)
	reqSysop := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	reqSysop.Header.Set("Authorization", "Bearer "+sysopTokens.AccessToken)
	recSysop := httptest.NewRecorder()
	fullStack.ServeHTTP(recSysop, reqSysop)

	if recSysop.Code != http.StatusOK {
		t.Errorf("expected 200 OK for sysop, got %d: %s", recSysop.Code, recSysop.Body.String())
	}
}

func TestTOTPWorkflow(t *testing.T) {
	handler, service, _ := setupTestApp()

	authRes, err := service.Register(RegisterRequest{
		Username: "SecurityUser",
		Email:    "sec@treewiki.org",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// 1. Setup TOTP
	setupRes, err := service.SetupTOTP(authRes.User.ID)
	if err != nil {
		t.Fatalf("setup TOTP failed: %v", err)
	}

	// 2. Compute current valid code
	code, err := auth.GenerateTOTPCode(setupRes.Secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate code failed: %v", err)
	}

	// 3. Verify TOTP
	if err := service.VerifyTOTP(authRes.User.ID, code); err != nil {
		t.Fatalf("verify TOTP failed: %v", err)
	}

	// 4. Login without TOTP should now be blocked
	loginReqNoTOTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"SecurityUser","password":"Password123!"}`))
	recNoTOTP := httptest.NewRecorder()
	handler.Login(recNoTOTP, loginReqNoTOTP)

	if recNoTOTP.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when missing required TOTP code, got %d", recNoTOTP.Code)
	}

	// 5. Login with valid TOTP should succeed
	loginPayloadWithTOTP := `{"username":"SecurityUser","password":"Password123!","totp_code":"` + code + `"}`
	reqWithTOTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginPayloadWithTOTP))
	recWithTOTP := httptest.NewRecorder()
	handler.Login(recWithTOTP, reqWithTOTP)

	if recWithTOTP.Code != http.StatusOK {
		t.Errorf("expected 200 OK with valid TOTP, got %d: %s", recWithTOTP.Code, recWithTOTP.Body.String())
	}
}

func TestGDPRAnonymization(t *testing.T) {
	_, service, _ := setupTestApp()

	authRes, err := service.Register(RegisterRequest{
		Username: "UserToDelete",
		Email:    "delete_me@treewiki.org",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Execute GDPR erasure
	if err := service.AnonymizeAccount(authRes.User.ID); err != nil {
		t.Fatalf("anonymize failed: %v", err)
	}

	// Try login with old credentials -> must fail
	_, err = service.Login(LoginRequest{
		Username: "UserToDelete",
		Password: "Password123!",
	})

	if err == nil {
		t.Errorf("expected login to fail after account anonymization")
	}
}

