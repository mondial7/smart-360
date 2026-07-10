package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvDoesNotOverrideExisting(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "" +
		"# a comment\n" +
		"FOO=from_file\n" +
		"BAR=\"quoted value\"\n" +
		"PRESET=from_file\n" +
		"\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PRESET", "from_env") // already set → must win
	os.Unsetenv("FOO")
	os.Unsetenv("BAR")
	t.Cleanup(func() { os.Unsetenv("FOO"); os.Unsetenv("BAR") })

	loadDotEnv(envPath)

	if got := os.Getenv("FOO"); got != "from_file" {
		t.Errorf("FOO = %q, want from_file", got)
	}
	if got := os.Getenv("BAR"); got != "quoted value" {
		t.Errorf("BAR = %q, want unquoted value", got)
	}
	if got := os.Getenv("PRESET"); got != "from_env" {
		t.Errorf("PRESET = %q, want from_env (env must override file)", got)
	}
}

func TestLoadRequiresSessionSecretInProd(t *testing.T) {
	// Run in a temp dir so a stray ./.env doesn't interfere.
	t.Chdir(t.TempDir())
	t.Setenv("DEV_MODE", "false")
	t.Setenv("SESSION_SECRET", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected an error when SESSION_SECRET is unset in production")
	}

	t.Setenv("DEV_MODE", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("dev mode should supply a fallback secret: %v", err)
	}
	if cfg.SessionSecret == "" {
		t.Fatal("expected a dev fallback session secret")
	}
}
