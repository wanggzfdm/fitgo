package coros

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"coros-fit-mcp/pkg/config"
)

func TestBuildLoginFormWithP1P2(t *testing.T) {
	form, err := buildLoginForm(config.CorosConfig{
		P1: "$2b$10$abc",
		P2: "$2b$10$def",
	})
	if err != nil {
		t.Fatalf("buildLoginForm returned error: %v", err)
	}
	if form.P1 != "$2b$10$abc" || form.P2 != "$2b$10$def" {
		t.Fatalf("unexpected form: %#v", form)
	}
	if form.Pwd != "" {
		t.Fatalf("expected empty pwd, got %#v", form)
	}
}

func TestBuildLoginFormWithLegacyPassword(t *testing.T) {
	form, err := buildLoginForm(config.CorosConfig{
		Password: "legacy-md5",
	})
	if err != nil {
		t.Fatalf("buildLoginForm returned error: %v", err)
	}
	if form.Pwd != "legacy-md5" {
		t.Fatalf("unexpected pwd: %#v", form)
	}
}

func TestBuildLoginFormRequiresBothP1P2(t *testing.T) {
	_, err := buildLoginForm(config.CorosConfig{
		P1: "$2b$10$abc",
	})
	if err == nil {
		t.Fatalf("expected error when p2 missing")
	}
}

func TestTokenCacheRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "coros_token.json")
	t.Setenv("COROS_TOKEN_CACHE_PATH", cachePath)

	want := tokenCache{
		Account:     "runner@example.com",
		AccessToken: "token-123",
		ExpiresAt:   time.Now().Add(time.Hour).Round(time.Second),
	}

	if err := saveTokenCache(want); err != nil {
		t.Fatalf("saveTokenCache returned error: %v", err)
	}

	got, err := loadTokenCache("runner@example.com")
	if err != nil {
		t.Fatalf("loadTokenCache returned error: %v", err)
	}

	if got.Account != want.Account || got.AccessToken != want.AccessToken || !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Fatalf("unexpected token cache: got %#v want %#v", got, want)
	}

	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("stat cache path failed: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected cache file mode 0600, got %o", info.Mode().Perm())
	}
}
