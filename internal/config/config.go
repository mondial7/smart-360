// Package config loads runtime configuration from the environment.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration. Values are read once at startup.
type Config struct {
	// Port is the TCP port the HTTP server listens on.
	Port string
	// DatabaseURL is the Postgres connection string (pgx format).
	DatabaseURL string
	// SessionSecret signs/authenticates session cookies. Required.
	SessionSecret string
	// AppURL is the externally reachable base URL (used for OAuth redirects).
	AppURL string

	// AdminEmail bootstraps the owner: the user who signs in with this email is
	// (and stays) a global admin. Deterministic and race-free, replacing the old
	// "first user to sign in becomes admin" heuristic. When empty, no admin is
	// auto-assigned.
	AdminEmail string

	// GeminiAPIKey enables the Gemini moderation + synthesis passes.
	// When empty, consolidation falls back to non-AI aggregation.
	GeminiAPIKey string

	// LogFormat selects the log handler: "text" (default) or "json".
	LogFormat string

	// Google OAuth credentials.
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// DevMode enables /auth/dev-login, relaxes the Secure cookie flag, and
	// seeds development data. Never enable in production.
	DevMode bool
}

// Load reads configuration from the environment, applying defaults for local
// development. A `.env` file in the working directory is loaded first (without
// overriding variables already set in the environment). It returns an error
// only for values that have no safe default.
func Load() (*Config, error) {
	loadDotEnv(".env")

	c := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://smart360:smart360@localhost:5432/smart360?sslmode=disable"),
		SessionSecret:      os.Getenv("SESSION_SECRET"),
		AppURL:             getEnv("APP_URL", "http://localhost:8080"),
		AdminEmail:         strings.TrimSpace(os.Getenv("ADMIN_EMAIL")),
		GeminiAPIKey:       os.Getenv("GEMINI_API_KEY"),
		LogFormat:          getEnv("LOG_FORMAT", "text"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		DevMode:            getBool("DEV_MODE", false),
	}

	if c.SessionSecret == "" {
		if !c.DevMode {
			return nil, fmt.Errorf("SESSION_SECRET is required")
		}
		// Deterministic dev fallback so local restarts don't invalidate sessions.
		c.SessionSecret = "dev-insecure-session-secret-change-me"
	}

	return c, nil
}

// loadDotEnv reads simple KEY=VALUE lines from path into the process
// environment, skipping blanks, comments, and keys already set. It is a minimal
// convenience loader for local development, not a full dotenv implementation.
func loadDotEnv(path string) {
	f, err := os.Open(path) // #nosec G304 -- path is the constant ".env" from Load, never user input
	if err != nil {
		return // no .env is fine
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return b
}
