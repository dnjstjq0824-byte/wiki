package pg

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestUUIDValRoundTrip(t *testing.T) {
	v := UUIDVal("550e8400-e29b-41d4-a716-446655440000")
	if !v.Valid {
		t.Fatal("expected valid UUID")
	}
	if got := UUIDString(v); got != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected round trip: %q", got)
	}
	if UUIDVal("not-a-uuid").Valid {
		t.Fatal("expected invalid UUID for garbage")
	}
	if UUIDString(UUIDVal("")) != "" {
		t.Fatal("expected empty string for NULL uuid")
	}
}

func TestTextVal(t *testing.T) {
	if TextVal("").Valid {
		t.Fatal("empty string must map to NULL")
	}
	if got := TextOrEmpty(TextVal("x")); got != "x" {
		t.Fatalf("unexpected text: %q", got)
	}
}

func TestInetVal(t *testing.T) {
	if InetVal("192.0.2.1") == nil {
		t.Fatal("expected IPv4 parse")
	}
	if InetVal("2001:db8::1") == nil {
		t.Fatal("expected IPv6 parse")
	}
	if InetVal("192.0.xxx.xxx") != nil {
		t.Fatal("masked IP must not parse")
	}
	if InetString(nil) != "" {
		t.Fatal("expected empty string for nil inet")
	}
}

func TestDisplayAuthorPriority(t *testing.T) {
	if got := DisplayAuthor("u", "n", "1.1.1.1", "fb"); got != "u" {
		t.Fatalf("username must win, got %q", got)
	}
	if got := DisplayAuthor("", "n", "1.1.1.1", "fb"); got != "n" {
		t.Fatalf("name must win, got %q", got)
	}
	if got := DisplayAuthor("", "", "1.1.1.1", "fb"); got != "1.1.1.1" {
		t.Fatalf("ip must win, got %q", got)
	}
	if got := DisplayAuthor("", "", "", "fb"); got != "fb" {
		t.Fatalf("fallback must win, got %q", got)
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := TruncateRunes("abcdef", 3); got != "abc" {
		t.Fatalf("unexpected truncation: %q", got)
	}
	if got := TruncateRunes("가나다라", 2); got != "가나" {
		t.Fatalf("rune truncation broken: %q", got)
	}
	if got := TruncateRunes("ab", 5); got != "ab" {
		t.Fatalf("short string must pass through: %q", got)
	}
}

func TestViolationClassifiers(t *testing.T) {
	unique := &pgconn.PgError{Code: "23505"}
	fk := &pgconn.PgError{Code: "23503"}
	check := &pgconn.PgError{Code: "23514"}
	if !IsUniqueViolation(unique) || IsFKViolation(unique) || IsCheckViolation(unique) {
		t.Fatal("unique classification wrong")
	}
	if !IsFKViolation(fk) || IsUniqueViolation(fk) || IsCheckViolation(fk) {
		t.Fatal("fk classification wrong")
	}
	if !IsCheckViolation(check) || IsUniqueViolation(check) || IsFKViolation(check) {
		t.Fatal("check classification wrong")
	}
	if IsUniqueViolation(nil) || IsFKViolation(nil) || IsCheckViolation(nil) {
		t.Fatal("nil must not classify")
	}
}
