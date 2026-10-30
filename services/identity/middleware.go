package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dnjstjqppk-dev/treewiki/pkg/auth"
)

type contextKey string

const userContextKey = contextKey("treewiki_user_context")

type UserContext struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	Role            string `json:"role"`
	IsAuthenticated bool   `json:"is_authenticated"`
	MaskedIP        string `json:"masked_ip"`
	RawIP           string `json:"raw_ip"`
}

// GetUserContext extracts authenticated user context from request
func GetUserContext(r *http.Request) *UserContext {
	if val := r.Context().Value(userContextKey); val != nil {
		if ctx, ok := val.(*UserContext); ok {
			return ctx
		}
	}
	return &UserContext{
		Role:            auth.RoleAnonymous,
		IsAuthenticated: false,
		MaskedIP:        "xxx.xxx.xxx.xxx",
	}
}

// AuthMiddleware inspects Authorization headers and session cookies
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				parts := strings.Split(forwarded, ",")
				rawIP = strings.TrimSpace(parts[0])
			}
			maskedIP := auth.MaskIP(rawIP)

			tokenStr := ""

			// 1. Try Authorization header
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			// 2. Try Cookie
			if tokenStr == "" {
				if cookie, err := r.Cookie("treewiki_session"); err == nil {
					tokenStr = cookie.Value
				}
			}

			userCtx := &UserContext{
				Role:            auth.RoleAnonymous,
				IsAuthenticated: false,
				MaskedIP:        maskedIP,
				RawIP:           rawIP,
			}

			if tokenStr != "" {
				if claims, err := auth.ValidateToken(tokenStr, jwtSecret); err == nil {
					userCtx.UserID = claims.UserID
					userCtx.Username = claims.Username
					userCtx.Role = claims.Role
					userCtx.IsAuthenticated = true
				}
			}

			ctx := context.WithValue(r.Context(), userContextKey, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth blocks unauthenticated requests with 401 Unauthorized
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := GetUserContext(r)
		if !u.IsAuthenticated {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Authentication required to access this resource",
			})
			return
		}
		next(w, r)
	}
}

// RequireRole enforces role hierarchy threshold (403 Forbidden on failure)
func RequireRole(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := GetUserContext(r)
		if !auth.HasPermission(u.Role, requiredRole) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Insufficient permissions: required role " + requiredRole,
			})
			return
		}
		next(w, r)
	}
}

