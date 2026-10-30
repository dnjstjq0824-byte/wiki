package config

import (
	"os"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	os.Setenv("PORT", "9999")
	os.Setenv("JWT_SECRET", "custom_test_jwt_secret_12345")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	if cfg.Port != "9999" {
		t.Errorf("expected PORT=9999, got %s", cfg.Port)
	}
	if cfg.JWTSecret != "custom_test_jwt_secret_12345" {
		t.Errorf("expected custom JWTSecret, got %s", cfg.JWTSecret)
	}
}

