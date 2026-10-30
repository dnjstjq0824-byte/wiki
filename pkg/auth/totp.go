package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"time"
)

// GenerateTOTPSecret creates a 160-bit random Base32 encoded secret for 2FA
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// GenerateTOTPCode computes the current 6-digit code for a given secret at time t
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", err
	}

	timestep := uint64(t.Unix() / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, timestep)

	h := hmac.New(sha1.New, key)
	h.Write(buf)
	sum := h.Sum(nil)

	offset := sum[len(sum)-1] & 0x0F
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7FFFFFFF
	code = code % 1000000

	return fmt.Sprintf("%06d", code), nil
}

// ValidateTOTPCode checks if the provided 6-digit code matches current or adjacent time window
func ValidateTOTPCode(secret, code string, t time.Time) bool {
	// Check current, previous, and next timestep to accommodate slight clock drift (+-30s)
	for _, offset := range []int64{-1, 0, 1} {
		shifted := t.Add(time.Duration(offset*30) * time.Second)
		expected, err := GenerateTOTPCode(secret, shifted)
		if err == nil && expected == code {
			return true
		}
	}
	return false
}

