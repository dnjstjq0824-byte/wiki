package identity

import (
	"time"
)

// User represents the persistent user account entity
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email           *string   `json:"email"`
	PasswordHash    string    `json:"-"`
	Role            string    `json:"role"`
	EditCount       int       `json:"edit_count"`
	IsBlocked       bool      `json:"is_blocked"`
	IsBot           bool      `json:"is_bot"`
	ProfileBio      string    `json:"profile_bio"`
	AvatarURL       string    `json:"avatar_url"`
	EmailVerified   bool      `json:"email_verified"`
	EmailVerifyToken string   `json:"-"`
	Badges            []string  `json:"badges"`
	ContributionCount int       `json:"contribution_count"`
	TOTPSecret      string    `json:"-"`
	TOTPEnabled     bool      `json:"totp_enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"`
}

type AuthResponse struct {
	User         UserPublic `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresIn    int64      `json:"expires_in"`
}

type UserPublic struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Role          string    `json:"role"`
	EditCount     int       `json:"edit_count"`
	Badges        []string  `json:"badges"`
	ContributionCount int     `json:"contribution_count"`
	EmailVerified bool      `json:"email_verified"`
	ProfileBio    string    `json:"profile_bio"`
	AvatarURL     string    `json:"avatar_url"`
	CreatedAt     time.Time `json:"created_at"`
}

type UpdateProfileRequest struct {
	ProfileBio string `json:"profile_bio"`
	AvatarURL  string `json:"avatar_url"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type TOTPSetupResponse struct {
	Secret string `json:"secret"`
}

type TOTPVerifyRequest struct {
	Code string `json:"code"`
}

