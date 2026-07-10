package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
)

const (
	oauthStateCookie     = "s360_oauth_state"
	oauthStateTTLSeconds = 600
)

// StartGoogleLogin sets a state cookie and redirects the browser to Google's
// consent screen. The state cookie defends against login-CSRF.
func (s *Service) StartGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GoogleClientID == "" {
		http.Error(w, "Google login is not configured", http.StatusServiceUnavailable)
		return
	}
	state, err := randomToken()
	if err != nil {
		http.Error(w, "Failed to initialize login", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookies(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   oauthStateTTLSeconds,
	})
	http.Redirect(w, r, s.oauth.AuthCodeURL(state), http.StatusFound)
}

// GoogleCallback completes the OAuth exchange, provisions/updates the user,
// issues a session, and redirects home.
func (s *Service) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validate and always clear the state cookie.
	state := r.URL.Query().Get("state")
	stateCookie, cookieErr := r.Cookie(oauthStateCookie)
	http.SetCookie(w, &http.Cookie{Name: oauthStateCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: s.secureCookies(), SameSite: http.SameSiteLaxMode})
	if cookieErr != nil || state == "" ||
		subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusBadGateway)
		return
	}
	resp, err := s.oauth.Client(ctx, token).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var gu struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil || gu.Email == "" {
		http.Error(w, "Failed to parse user info", http.StatusBadGateway)
		return
	}

	user, err := s.upsertUser(ctx, gu.Email, gu.Name, gu.Picture)
	if err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}
	if err := s.issueSession(ctx, w, user.ID); err != nil {
		http.Error(w, "Failed to start session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// DevLogin bypasses Google OAuth for local development. It is only mounted when
// DEV_MODE=true. Any email logs in; an unknown email is provisioned as a member
// (except the default dev admin), matching the legacy behaviour.
func (s *Service) DevLogin(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.DevMode {
		http.NotFound(w, r)
		return
	}
	ctx := r.Context()
	email := r.URL.Query().Get("email")
	if email == "" {
		email = "dev@example.com"
	}

	user, err := s.repos.Users.FindByEmail(ctx, email)
	if errors.Is(err, repo.ErrNotFound) {
		name := "Dev User"
		if email == "dev@example.com" {
			name = "Dev Admin"
		}
		user, err = s.provisionUser(ctx, email, name, "")
	}
	if err != nil {
		http.Error(w, "Dev login failed", http.StatusInternalServerError)
		return
	}
	_ = s.repos.Users.UpdateLastLogin(ctx, user.ID)
	if err := s.issueSession(ctx, w, user.ID); err != nil {
		http.Error(w, "Failed to start session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// Logout deletes the session and returns to the login page.
func (s *Service) Logout(w http.ResponseWriter, r *http.Request) {
	s.clearSession(r.Context(), w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// upsertUser finds a user by email or creates one, then updates last_login.
func (s *Service) upsertUser(ctx context.Context, email, name, photo string) (*models.User, error) {
	user, err := s.repos.Users.FindByEmail(ctx, email)
	switch {
	case err == nil:
		// keep existing role; refresh display fields is out of scope here
	case errors.Is(err, repo.ErrNotFound):
		user, err = s.provisionUser(ctx, email, name, photo)
		if err != nil {
			return nil, err
		}
	default:
		return nil, err
	}
	_ = s.repos.Users.UpdateLastLogin(ctx, user.ID)
	return user, nil
}

// provisionUser creates a new user. The very first user in the system becomes
// the global admin; everyone else starts as a member.
func (s *Service) provisionUser(ctx context.Context, email, name, photo string) (*models.User, error) {
	role := models.RoleMember
	if existing, err := s.repos.Users.FindAll(ctx); err == nil && len(existing) == 0 {
		role = models.RoleAdmin
	}
	now := time.Now()
	user := &models.User{Email: email, Name: name, PhotoURL: photo, Role: role, LastLogin: &now}
	if err := s.repos.Users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
