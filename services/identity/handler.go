package identity

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Handler struct {
	service *IdentityService
}

func NewHandler(service *IdentityService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	res, err := h.service.Register(req)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	setSessionCookie(w, res.AccessToken)
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	res, err := h.service.Login(req)
	if err != nil {
		status := http.StatusUnauthorized
		if err == ErrAccountBlocked {
			status = http.StatusForbidden
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	setSessionCookie(w, res.AccessToken)
	respondJSON(w, http.StatusOK, res)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	token := body.RefreshToken
	if token == "" {
		if c, err := r.Cookie("treewiki_refresh"); err == nil {
			token = c.Value
		}
	}

	res, err := h.service.Refresh(token)
	if err != nil {
		status := http.StatusUnauthorized
		if err == ErrAccountBlocked {
			status = http.StatusForbidden
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	setSessionCookie(w, res.AccessToken)
	respondJSON(w, http.StatusOK, res)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "treewiki_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	respondJSON(w, http.StatusOK, map[string]string{"status": "Logged out successfully"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := GetUserContext(r)
	if !ctx.IsAuthenticated {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"is_authenticated": false,
			"role":             ctx.Role,
			"masked_ip":        ctx.MaskedIP,
		})
		return
	}

	profile, err := h.service.GetProfile(ctx.UserID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "User profile not found"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"is_authenticated": true,
		"user":             profile,
		"masked_ip":        ctx.MaskedIP,
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := GetUserContext(r)
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	pub, err := h.service.UpdateProfile(ctx.UserID, req)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, pub)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := GetUserContext(r)
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	if err := h.service.ChangePassword(ctx.UserID, req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "Password updated successfully"})
}

func (h *Handler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	ctx := GetUserContext(r)
	res, err := h.service.SetupTOTP(ctx.UserID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, res)
}

func (h *Handler) VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	ctx := GetUserContext(r)
	var req TOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	if err := h.service.VerifyTOTP(ctx.UserID, req.Code); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "Two-factor authentication enabled successfully"})
}

func (h *Handler) AnonymizeAccount(w http.ResponseWriter, r *http.Request) {
	ctx := GetUserContext(r)
	if err := h.service.AnonymizeAccount(ctx.UserID); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.Logout(w, r)
}

func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Username required"})
		return
	}

	profile, err := h.service.GetPublicProfile(username)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "treewiki_session",
		Value:    token,
		Path:     "/",
		MaxAge:   7200,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}
	if err := h.service.VerifyEmail(body.Token); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "Email verified successfully"})
}

func (h *Handler) OAuthProviders(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"providers": h.service.OAuthProviders(),
	})
}

func (h *Handler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	authURL, err := h.service.BeginOAuthLogin(provider)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrOAuthProviderUnknown {
			status = http.StatusNotFound
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	q := r.URL.Query()
	jsonMode := q.Get("format") == "json"

	fail := func(err error) {
		if jsonMode {
			status := http.StatusBadRequest
			if err == ErrOAuthProviderUnknown {
				status = http.StatusNotFound
			}
			respondJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		target := h.service.OAuthFrontendURL() + "/auth/callback?provider=" + url.QueryEscape(provider) + "&error=" + url.QueryEscape(err.Error())
		http.Redirect(w, r, target, http.StatusFound)
	}

	if providerErr := q.Get("error"); providerErr != "" {
		fail(ErrOAuthExchange)
		return
	}
	code, err := h.service.CompleteOAuthLogin(provider, q.Get("code"), q.Get("state"))
	if err != nil {
		fail(err)
		return
	}
	if jsonMode {
		respondJSON(w, http.StatusOK, map[string]string{"provider": provider, "code": code})
		return
	}
	target := h.service.OAuthFrontendURL() + "/auth/callback?provider=" + url.QueryEscape(provider) + "&code=" + url.QueryEscape(code)
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *Handler) OAuthConsume(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "OAuth code is required"})
		return
	}

	res, err := h.service.ConsumeOAuthCode(body.Code)
	if err != nil {
		status := http.StatusUnauthorized
		if err == ErrAccountBlocked {
			status = http.StatusForbidden
		}
		respondJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	setSessionCookie(w, res.AccessToken)
	respondJSON(w, http.StatusOK, res)
}

func (h *Handler) ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	ctx := GetUserContext(r)
	token, err := h.service.ResendEmailVerification(ctx.UserID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "Verification email sent", "token": token})
}
