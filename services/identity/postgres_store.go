package identity

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserStore struct {
	pool *pgxpool.Pool
}

func NewPostgresUserStore(pool *pgxpool.Pool) *PostgresUserStore {
	return &PostgresUserStore{pool: pool}
}

func (s *PostgresUserStore) Create(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (id, username, email, password_hash, role, edit_count, is_blocked, is_bot, profile_bio, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := s.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.EditCount,
		user.IsBlocked,
		user.IsBot,
		user.ProfileBio,
		user.AvatarURL,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_username_key") {
			return ErrUsernameDuplicate
		}
		if strings.Contains(err.Error(), "users_email_key") {
			return ErrEmailDuplicate
		}
		return err
	}
	return nil
}

func (s *PostgresUserStore) GetByID(id string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, role, edit_count, is_blocked, is_bot, profile_bio, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u User
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.EditCount,
		&u.IsBlocked,
		&u.IsBot,
		&u.ProfileBio,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *PostgresUserStore) GetByUsername(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, role, edit_count, is_blocked, is_bot, profile_bio, avatar_url, created_at, updated_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
	`
	var u User
	err := s.pool.QueryRow(ctx, query, username).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.EditCount,
		&u.IsBlocked,
		&u.IsBot,
		&u.ProfileBio,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *PostgresUserStore) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, role, edit_count, is_blocked, is_bot, profile_bio, avatar_url, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
	var u User
	err := s.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.EditCount,
		&u.IsBlocked,
		&u.IsBot,
		&u.ProfileBio,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *PostgresUserStore) GetByVerificationToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if token == "" {
		return nil, ErrUserNotFound
	}

	query := `
		SELECT id, username, email, password_hash, role, edit_count, is_blocked, is_bot, profile_bio, avatar_url, created_at, updated_at
		FROM users
		WHERE profile_bio LIKE $1
	`
	var u User
	err := s.pool.QueryRow(ctx, query, "%token:"+token+"%").Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.EditCount,
		&u.IsBlocked,
		&u.IsBot,
		&u.ProfileBio,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *PostgresUserStore) Update(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, role = $5, edit_count = $6, is_blocked = $7, is_bot = $8, profile_bio = $9, avatar_url = $10, updated_at = $11
		WHERE id = $1
	`
	res, err := s.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.EditCount,
		user.IsBlocked,
		user.IsBot,
		user.ProfileBio,
		user.AvatarURL,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
