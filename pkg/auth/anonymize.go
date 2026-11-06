package auth

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// AnonymizedProfile represents the purged user record structure
type AnonymizedProfile struct {
	Username     string
	Email        *string
	PasswordHash string
	ProfileBio   string
	AvatarURL    string
	IsBlocked    bool
}

// GenerateAnonymizedProfile creates a pseudonymized identity fulfilling GDPR Art. 17 & PIPA
func GenerateAnonymizedProfile() AnonymizedProfile {
	randomUUID := strings.ReplaceAll(uuid.New().String()[:8], "-", "")
	return AnonymizedProfile{
		Username:     fmt.Sprintf("FormerUser_%s", randomUUID),
		Email:        nil,
		PasswordHash: "",
		ProfileBio:   "This user account was permanently anonymized pursuant to privacy regulations.",
		AvatarURL:    "",
		IsBlocked:    true,
	}
}
