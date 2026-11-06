package identity

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	treepg "github.com/dnjstjqppk-dev/treewiki/pkg/pg"
)

type PostgresUserStore struct {
	pool *pgxpool.Pool
}

func NewPostgresUserStore(pool *pgxpool.Pool) *PostgresUserStore {
	return &PostgresUserStore{pool: pool}
}

const userColumns = `id, username, email, password_hash, role, edit_count, is_blocked, is_bot, profile_bio, avatar_url, created_at, updated_at`

func scanUser(row pgx.Row) (*User, error) {
	var (
		u      User
		id     pgtype.UUID
		email  pgtype.Text
		pwHash pgtype.Text
		bio    pgtype.Text
		avatar pgtype.Text
	)
	err := row.Scan(
		&id,
		&u.Username,
		&email,
		&pwHash,
		&u.Role,
		&u.EditCount,
		&u.IsBlocked,
		&u.IsBot,
		&bio,
		&avatar,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.ID = treepg.UUIDString(id)
	if email.Valid {
		mail := email.String
		u.Email = &mail
	}
	u.PasswordHash = treepg.TextOrEmpty(pwHash)
	u.ProfileBio = treepg.TextOrEmpty(bio)
	u.AvatarURL = treepg.TextOrEmpty(avatar)
	return &u, nil
}

func (s *PostgresUserStore) Create(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (id, username, email, password_hash, role, edit_count, is_blocked, is_bot, profile_bio, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := s.pool.Exec(ctx, query,
		treepg.UUIDVal(user.ID),
		user.Username,
		user.Email,
		treepg.TextVal(user.PasswordHash),
		user.Role,
		user.EditCount,
		user.IsBlocked,
		user.IsBot,
		treepg.TextVal(user.ProfileBio),
		treepg.TextVal(user.AvatarURL),
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

	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	return scanUser(s.pool.QueryRow(ctx, query, treepg.UUIDVal(id)))
}

func (s *PostgresUserStore) GetByUsername(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT ` + userColumns + ` FROM users WHERE LOWER(username) = LOWER($1)`
	return scanUser(s.pool.QueryRow(ctx, query, username))
}

func (s *PostgresUserStore) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT ` + userColumns + ` FROM users WHERE LOWER(email) = LOWER($1)`
	return scanUser(s.pool.QueryRow(ctx, query, email))
}

func (s *PostgresUserStore) GetByVerificationToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if token == "" {
		return nil, ErrUserNotFound
	}

	query := `SELECT ` + userColumns + ` FROM users WHERE profile_bio LIKE $1`
	return scanUser(s.pool.QueryRow(ctx, query, "%token:"+token+"%"))
}

func (s *PostgresUserStore) GetByOAuthLink(provider, providerUserID string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.role, u.edit_count, u.is_blocked, u.is_bot, u.profile_bio, u.avatar_url, u.created_at, u.updated_at
		FROM users u
		JOIN oauth_identities o ON o.user_id = u.id
		WHERE o.provider = $1 AND o.provider_user_id = $2
	`
	return scanUser(s.pool.QueryRow(ctx, query, provider, providerUserID))
}

func (s *PostgresUserStore) CreateOAuthLink(userID, provider, providerUserID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var owner pgtype.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT user_id FROM oauth_identities WHERE provider = $1 AND provider_user_id = $2`,
		provider, providerUserID,
	).Scan(&owner)
	if err == nil {
		if treepg.UUIDString(owner) == userID {
			return nil
		}
		return ErrOAuthLinkTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO oauth_identities (provider, provider_user_id, user_id) VALUES ($1, $2, $3)`,
		provider, providerUserID, treepg.UUIDVal(userID),
	)
	if err != nil {
		if treepg.IsUniqueViolation(err) {
			return ErrOAuthLinkTaken
		}
		if treepg.IsFKViolation(err) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
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
		treepg.UUIDVal(user.ID),
		user.Username,
		user.Email,
		treepg.TextVal(user.PasswordHash),
		user.Role,
		user.EditCount,
		user.IsBlocked,
		user.IsBot,
		treepg.TextVal(user.ProfileBio),
		treepg.TextVal(user.AvatarURL),
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
