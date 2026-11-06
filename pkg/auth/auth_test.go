package auth

import (
	"testing"
	"time"
)

func TestPasswordHashing(t *testing.T) {
	weakPasswords := []string{
		"short",
		"alllowercase1",
		"ALLUPPERCASE1",
		"NoNumbersHere",
	}

	for _, pw := range weakPasswords {
		if err := ValidatePasswordStrength(pw); err == nil {
			t.Errorf("expected weak password %q to fail validation", pw)
		}
	}

	validPassword := "SecurePassword123!"
	if err := ValidatePasswordStrength(validPassword); err != nil {
		t.Fatalf("expected valid password, got error: %v", err)
	}

	hash, err := HashPassword(validPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !CheckPasswordHash(validPassword, hash) {
		t.Errorf("expected CheckPasswordHash to succeed for correct password")
	}

	if CheckPasswordHash("WrongPassword123!", hash) {
		t.Errorf("expected CheckPasswordHash to fail for wrong password")
	}
}

func TestJWTLifecycle(t *testing.T) {
	secret := "test_secret_key_1234567890_abcdef"
	userID := "550e8400-e29b-41d4-a716-446655440000"
	username := "TestEditor"
	role := RoleEditor

	pair, err := GenerateTokenPair(userID, username, role, false, secret)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	claims, err := ValidateToken(pair.AccessToken, secret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != userID || claims.Username != username || claims.Role != role {
		t.Errorf("token claims mismatch: %+v", claims)
	}

	// Validate with wrong secret
	if _, err := ValidateToken(pair.AccessToken, "wrong_secret"); err == nil {
		t.Errorf("expected validation to fail with wrong secret")
	}
}

func TestIPMasking(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.50", "192.168.1.xxx"},
		{"203.0.113.195:8080", "203.0.113.xxx"},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:db8:85a3::/64"},
		{"invalid-ip", "xxx.xxx.xxx.xxx"},
	}

	for _, tt := range tests {
		res := MaskIP(tt.input)
		if res != tt.expected {
			t.Errorf("MaskIP(%q) = %q, want %q", tt.input, res, tt.expected)
		}
	}
}

func TestRBACPermissions(t *testing.T) {
	if !HasPermission(RoleSysop, RoleEditor) {
		t.Errorf("SYSOP should have EDITOR permissions")
	}

	if HasPermission(RoleAnonymous, RoleRegistered) {
		t.Errorf("ANONYMOUS should not have REGISTERED permissions")
	}

	if !HasPermission(RoleSystemAdmin, RoleSysop) {
		t.Errorf("SYSTEM_ADMIN should have SYSOP permissions")
	}

	// AutoConfirmed qualification test
	oldDate := time.Now().Add(-8 * 24 * time.Hour)
	recentDate := time.Now().Add(-2 * 24 * time.Hour)

	if !CheckAutoConfirmed(oldDate, 20) {
		t.Errorf("user registered 8 days ago with 20 edits should be AutoConfirmed")
	}

	if CheckAutoConfirmed(oldDate, 5) {
		t.Errorf("user with 5 edits should not be AutoConfirmed")
	}

	if CheckAutoConfirmed(recentDate, 50) {
		t.Errorf("user registered 2 days ago should not be AutoConfirmed")
	}

	// Dynamic effective role
	if EffectiveRole(RoleRegistered, oldDate, 20) != RoleAutoConfirmed {
		t.Errorf("EffectiveRole should promote qualifying user to AUTOCONFIRMED")
	}
}

func TestTOTP(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("failed to generate TOTP secret: %v", err)
	}

	now := time.Now().UTC()
	code, err := GenerateTOTPCode(secret, now)
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %q", code)
	}

	if !ValidateTOTPCode(secret, code, now) {
		t.Errorf("expected valid TOTP code to pass verification")
	}

	if ValidateTOTPCode(secret, "999999", now) && code != "999999" {
		t.Errorf("invalid code should fail verification")
	}
}

func TestAnonymization(t *testing.T) {
	profile := GenerateAnonymizedProfile()
	if profile.Email != nil {
		t.Errorf("anonymized email must be nil")
	}
	if profile.PasswordHash != "" {
		t.Errorf("anonymized password hash must be empty")
	}
	if !profile.IsBlocked {
		t.Errorf("anonymized account must be blocked")
	}
}
