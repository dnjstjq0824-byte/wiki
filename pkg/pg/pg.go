// Package pg provides shared PostgreSQL helpers for Treewiki stores:
// bounded contexts, UUID/IP parsing, nullable mapping, and
// constraint-violation classification.
package pg

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Ctx returns a bounded context for store queries (5s timeout,
// matching the existing store convention).
func Ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// ParseUUID parses s as a UUID, reporting validity instead of an error.
func ParseUUID(s string) (uuid.UUID, bool) {
	u, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return uuid.Nil, false
	}
	return u, true
}

// UUIDVal maps s to a pgtype.UUID (invalid/NULL when unparseable).
func UUIDVal(s string) pgtype.UUID {
	if u, ok := ParseUUID(s); ok {
		return pgtype.UUID{Bytes: u, Valid: true}
	}
	return pgtype.UUID{}
}

// UUIDString renders a pgtype.UUID ("" when NULL).
func UUIDString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	v := uuid.UUID(u.Bytes)
	return v.String()
}

// MustUUID parses s or returns uuid.Nil (for IDs the store generated).
func MustUUID(s string) uuid.UUID {
	u, _ := ParseUUID(s)
	return u
}

// NewUUID returns a random UUID string for store-generated IDs.
func NewUUID() string {
	return uuid.NewString()
}

// TextVal maps s to a pgtype.Text (NULL when empty).
func TextVal(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// TextOrEmpty renders a pgtype.Text ("" when NULL).
func TextOrEmpty(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// InetVal parses s as an IP network (nil when unparseable, e.g. masked IPs).
func InetVal(s string) *net.IPNet {
	ip := net.ParseIP(strings.TrimSpace(s))
	if ip == nil {
		return nil
	}
	if ip.To4() != nil {
		return &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
	}
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}
}

// InetString renders a scanned inet value ("" when NULL).
func InetString(ip *net.IPNet) string {
	if ip == nil {
		return ""
	}
	return ip.IP.String()
}

// DisplayAuthor picks the human-facing author string with the same priority
// everywhere: registered username, display name, IP, then fallback.
func DisplayAuthor(username, name, ip, fallback string) string {
	if username != "" {
		return username
	}
	if name != "" {
		return name
	}
	if ip != "" {
		return ip
	}
	return fallback
}

// TruncateRunes shortens s to at most max runes for VARCHAR columns.
func TruncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// IsUniqueViolation reports duplicate-key errors (SQLSTATE 23505).
func IsUniqueViolation(err error) bool {
	return pgCode(err) == "23505"
}

// IsFKViolation reports foreign-key errors (SQLSTATE 23503).
func IsFKViolation(err error) bool {
	return pgCode(err) == "23503"
}

// IsCheckViolation reports check-constraint errors (SQLSTATE 23514).
func IsCheckViolation(err error) bool {
	return pgCode(err) == "23514"
}
