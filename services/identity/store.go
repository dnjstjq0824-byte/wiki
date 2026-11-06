package identity

import (
	"errors"
	"strings"
	"sync"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameDuplicate = errors.New("username is already taken")
	ErrEmailDuplicate    = errors.New("email is already registered")
)

type UserStore interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByVerificationToken(token string) (*User, error)
	GetByOAuthLink(provider, providerUserID string) (*User, error)
	CreateOAuthLink(userID, provider, providerUserID string) error
	Update(user *User) error
}

type MemoryUserStore struct {
	mu          sync.RWMutex
	usersByID   map[string]*User
	usersByName map[string]*User
	usersByMail map[string]*User
	oauthLinks  map[string]string
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		usersByID:   make(map[string]*User),
		usersByName: make(map[string]*User),
		usersByMail: make(map[string]*User),
		oauthLinks:  make(map[string]string),
	}
}

func oauthLinkKey(provider, providerUserID string) string {
	return provider + "\x00" + providerUserID
}

func (s *MemoryUserStore) Create(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	normName := strings.ToLower(user.Username)
	if _, exists := s.usersByName[normName]; exists {
		return ErrUsernameDuplicate
	}

	if user.Email != nil {
		normMail := strings.ToLower(*user.Email)
		if _, exists := s.usersByMail[normMail]; exists {
			return ErrEmailDuplicate
		}
		s.usersByMail[normMail] = user
	}

	s.usersByID[user.ID] = user
	s.usersByName[normName] = user
	return nil
}

func (s *MemoryUserStore) GetByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByID[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return cloneUser(user), nil
}

func (s *MemoryUserStore) GetByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByName[strings.ToLower(username)]
	if !exists {
		return nil, ErrUserNotFound
	}
	return cloneUser(user), nil
}

func (s *MemoryUserStore) GetByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByMail[strings.ToLower(email)]
	if !exists {
		return nil, ErrUserNotFound
	}
	return cloneUser(user), nil
}

func (s *MemoryUserStore) GetByVerificationToken(token string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.usersByID {
		if u.EmailVerifyToken == token && token != "" {
			return cloneUser(u), nil
		}
	}
	return nil, ErrUserNotFound
}
func (s *MemoryUserStore) GetByOAuthLink(provider, providerUserID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, linked := s.oauthLinks[oauthLinkKey(provider, providerUserID)]
	if !linked {
		return nil, ErrUserNotFound
	}
	user, exists := s.usersByID[userID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return cloneUser(user), nil
}

func (s *MemoryUserStore) CreateOAuthLink(userID, provider, providerUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByID[userID]; !exists {
		return ErrUserNotFound
	}
	key := oauthLinkKey(provider, providerUserID)
	if owner, linked := s.oauthLinks[key]; linked {
		if owner == userID {
			return nil
		}
		return ErrOAuthLinkTaken
	}
	s.oauthLinks[key] = userID
	return nil
}

func (s *MemoryUserStore) Update(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.usersByID[user.ID]
	if !exists {
		return ErrUserNotFound
	}

	// Update username index (always refresh pointer, even if unchanged)
	if !strings.EqualFold(existing.Username, user.Username) {
		delete(s.usersByName, strings.ToLower(existing.Username))
	}
	s.usersByName[strings.ToLower(user.Username)] = user

	// Update email index (always refresh pointer, even if unchanged)
	if existing.Email != nil && (user.Email == nil || !strings.EqualFold(*existing.Email, *user.Email)) {
		delete(s.usersByMail, strings.ToLower(*existing.Email))
	}
	if user.Email != nil {
		s.usersByMail[strings.ToLower(*user.Email)] = user
	}

	s.usersByID[user.ID] = user
	return nil
}

func cloneUser(u *User) *User {
	cp := *u
	if u.Email != nil {
		mail := *u.Email
		cp.Email = &mail
	}
	return &cp
}
