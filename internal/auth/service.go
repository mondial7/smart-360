// Package auth handles Google OAuth login, server-side sessions stored in
// Postgres with an HttpOnly cookie, role-based middleware, and CSRF protection.
package auth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/mondial7/smart-360/internal/config"
	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
)

type ctxKey int

const userCtxKey ctxKey = iota

// Service owns authentication: OAuth config, the session/user repositories, and
// the session cookie secret.
type Service struct {
	cfg    *config.Config
	repos  repo.Repositories
	oauth  *oauth2.Config
	secret []byte
}

// New builds an auth Service. The OAuth config is derived from the app config;
// it works even with empty Google credentials (only the OAuth routes fail then,
// and dev-login remains available when DEV_MODE=true).
func New(cfg *config.Config, repos repo.Repositories) *Service {
	return &Service{
		cfg:   cfg,
		repos: repos,
		oauth: &oauth2.Config{
			RedirectURL:  cfg.GoogleRedirectURL,
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		secret: []byte(cfg.SessionSecret),
	}
}

// UserFrom returns the authenticated user from a request context, if any.
func UserFrom(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(userCtxKey).(*models.User)
	return u, ok
}

// contextWithUser returns a child context carrying the user.
func contextWithUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// secureCookies reports whether cookies should carry the Secure flag. Relaxed
// in dev so login works over plain http on localhost.
func (s *Service) secureCookies() bool { return !s.cfg.DevMode }

// DevModeEnabled exposes whether dev-login is available, for route wiring.
func (s *Service) DevModeEnabled() bool { return s.cfg.DevMode }

// wantsHTML is a tiny helper for middleware to decide between redirect and
// status responses. Always true here (this is an HTML app) but centralized so
// it can grow (e.g. for htmx partial requests).
func wantsRedirect(r *http.Request) bool {
	// htmx swaps won't follow a 303 body; it needs an HX-Redirect header instead.
	return r.Header.Get("HX-Request") == ""
}
