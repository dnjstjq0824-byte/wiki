package identity

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dnjstjqppk-dev/treewiki/pkg/auth"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountBlocked     = errors.New("user account is blocked from signing in")
	ErrInvalidTOTP        = errors.New("invalid 2FA one-time code")
	ErrTOTPRequired       = errors.New("two-factor authentication code is required")
)

type IdentityService struct {
	store     UserStore
	jwtSecret string

	oauthMu           sync.Mutex
	oauthProviders    map[string]*OAuthProvider
	oauthStates       map[string]oauthStatePending
	oauthCodes        map[string]oauthCodePending
	oauthHTTP         *http.Client
	oauthFrontendURL  string
	oauthRedirectBase string
}

func NewIdentityService(store UserStore, jwtSecret string) *IdentityService {
	return &IdentityService{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (s *IdentityService) Register(req RegisterRequest) (*AuthResponse, error) {
	if !auth.ValidateUsername(req.Username) {
		return nil, errors.New("username must be 3-32 alphanumeric characters, dash, dot, or underscore")
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	email := strings.TrimSpace(req.Email)

	user := &User{
		ID:               uuid.New().String(),
		Username:         req.Username,
		Email:            &email,
		PasswordHash:     hash,
		Role:             auth.RoleRegistered,
		EditCount:        0,
		IsBlocked:        false,
		IsBot:            false,
		EmailVerified:    false,
		EmailVerifyToken: uuid.New().String(),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.store.Create(user); err != nil {
		return nil, err
	}

	tokens, err := auth.GenerateTokenPair(user.ID, user.Username, user.Role, user.IsBot, s.jwtSecret)
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

func (s *IdentityService) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.store.GetByUsername(req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.IsBlocked {
		return nil, ErrAccountBlocked
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Verify 2FA TOTP if enabled
	if user.TOTPEnabled {
		if req.TOTPCode == "" {
			return nil, ErrTOTPRequired
		}
		if !auth.ValidateTOTPCode(user.TOTPSecret, req.TOTPCode, time.Now().UTC()) {
			return nil, ErrInvalidTOTP
		}
	}

	// Calculate dynamic effective role (e.g. promoting to AutoConfirmed)
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

func (s *IdentityService) GetProfile(userID string) (*UserPublic, error) {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return nil, err
	}
	pub := toUserPublic(user)
	return &pub, nil
}

func (s *IdentityService) GetPublicProfile(username string) (*UserPublic, error) {
	user, err := s.store.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	pub := toUserPublic(user)
	return &pub, nil
}

func (s *IdentityService) UpdateProfile(userID string, req UpdateProfileRequest) (*UserPublic, error) {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return nil, err
	}

	user.ProfileBio = req.ProfileBio
	user.AvatarURL = req.AvatarURL
	user.UpdatedAt = time.Now().UTC()

	if err := s.store.Update(user); err != nil {
		return nil, err
	}

	pub := toUserPublic(user)
	return &pub, nil
}

func (s *IdentityService) ChangePassword(userID string, req ChangePasswordRequest) error {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return err
	}

	if !auth.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	user.UpdatedAt = time.Now().UTC()

	return s.store.Update(user)
}

func (s *IdentityService) SetupTOTP(userID string) (*TOTPSetupResponse, error) {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return nil, err
	}

	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}

	user.TOTPSecret = secret
	user.UpdatedAt = time.Now().UTC()

	if err := s.store.Update(user); err != nil {
		return nil, err
	}

	return &TOTPSetupResponse{Secret: secret}, nil
}

func (s *IdentityService) VerifyTOTP(userID, code string) error {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return err
	}

	if user.TOTPSecret == "" {
		return errors.New("2FA setup not initiated")
	}

	if !auth.ValidateTOTPCode(user.TOTPSecret, code, time.Now().UTC()) {
		return ErrInvalidTOTP
	}

	user.TOTPEnabled = true
	user.UpdatedAt = time.Now().UTC()

	return s.store.Update(user)
}

func (s *IdentityService) AnonymizeAccount(userID string) error {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return err
	}

	anon := auth.GenerateAnonymizedProfile()
	user.Username = anon.Username
	user.Email = anon.Email
	user.PasswordHash = anon.PasswordHash
	user.ProfileBio = anon.ProfileBio
	user.AvatarURL = anon.AvatarURL
	user.IsBlocked = anon.IsBlocked
	user.UpdatedAt = time.Now().UTC()

	return s.store.Update(user)
}

func (s *IdentityService) Refresh(refreshToken string) (*AuthResponse, error) {
	userID, err := auth.ValidateRefreshToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.store.GetByID(userID)
	if err != nil {
		return nil, ErrInvalidCredentials
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

func (s *IdentityService) IncrementContribution(userID string) error {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return err
	}
	user.ContributionCount++
	user.UpdatedAt = time.Now().UTC()
	// auto badge award
	if user.ContributionCount >= 15 && user.Role == auth.RoleRegistered {
		// promote handled by EffectiveRole at login
	}
	if user.ContributionCount == 1 {
		user.Badges = appendUnique(user.Badges, "FirstEdit")
	}
	if user.ContributionCount >= 50 {
		user.Badges = appendUnique(user.Badges, "ProlificEditor")
	}
	return s.store.Update(user)
}

func (s *IdentityService) AwardBadge(userID, badge string) error {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return err
	}
	user.Badges = appendUnique(user.Badges, badge)
	user.UpdatedAt = time.Now().UTC()
	return s.store.Update(user)
}

func appendUnique(list []string, val string) []string {
	for _, v := range list {
		if v == val {
			return list
		}
	}
	return append(list, val)
}
func (s *IdentityService) VerifyEmail(token string) error {
	user, err := s.store.GetByVerificationToken(token)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}
	user.EmailVerified = true
	user.EmailVerifyToken = ""
	user.UpdatedAt = time.Now().UTC()
	return s.store.Update(user)
}

func (s *IdentityService) ResendEmailVerification(userID string) (string, error) {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return "", err
	}
	if user.EmailVerified {
		return "", errors.New("email already verified")
	}
	user.EmailVerifyToken = uuid.New().String()
	user.UpdatedAt = time.Now().UTC()
	if err := s.store.Update(user); err != nil {
		return "", err
	}
	return user.EmailVerifyToken, nil
}

func (s *IdentityService) GetVerificationToken(userID string) (string, error) {
	user, err := s.store.GetByID(userID)
	if err != nil {
		return "", err
	}
	return user.EmailVerifyToken, nil
}
func toUserPublic(u *User) UserPublic {
	return UserPublic{
		ID:                u.ID,
		Username:          u.Username,
		Role:              auth.EffectiveRole(u.Role, u.CreatedAt, u.EditCount),
		EditCount:         u.EditCount,
		Badges:            u.Badges,
		ContributionCount: u.ContributionCount,
		EmailVerified:     u.EmailVerified,
		ProfileBio:        u.ProfileBio,
		AvatarURL:         u.AvatarURL,
		CreatedAt:         u.CreatedAt,
	}
}
