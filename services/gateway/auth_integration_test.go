package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthIntegrationFlow(t *testing.T) {
	router := NewRouter()

	regPayload := `{"username":"GatewayUser","email":"gw@treewiki.org","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(regPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", rec.Code, rec.Body.String())
	}

	var authRes map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &authRes); err != nil {
		t.Fatalf("decode register: %v", err)
	}
	refreshToken, _ := authRes["refresh_token"].(string)
	if refreshToken == "" {
		t.Fatalf("missing refresh token")
	}

	var sessionCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "treewiki_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("missing session cookie")
	}

	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMe.AddCookie(sessionCookie)
	recMe := httptest.NewRecorder()
	router.ServeHTTP(recMe, reqMe)
	if recMe.Code != http.StatusOK {
		t.Fatalf("me failed: %d %s", recMe.Code, recMe.Body.String())
	}
	var meRes map[string]interface{}
	_ = json.Unmarshal(recMe.Body.Bytes(), &meRes)
	if meRes["is_authenticated"] != true {
		t.Fatalf("expected authenticated, got %v", meRes["is_authenticated"])
	}

	reqPub := httptest.NewRequest(http.MethodGet, "/api/v1/users/GatewayUser", nil)
	recPub := httptest.NewRecorder()
	router.ServeHTTP(recPub, reqPub)
	if recPub.Code != http.StatusOK {
		t.Fatalf("public profile failed: %d %s", recPub.Code, recPub.Body.String())
	}

	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	reqRef := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
	reqRef.Header.Set("Content-Type", "application/json")
	recRef := httptest.NewRecorder()
	router.ServeHTTP(recRef, reqRef)
	if recRef.Code != http.StatusOK {
		t.Fatalf("refresh failed: %d %s", recRef.Code, recRef.Body.String())
	}
}