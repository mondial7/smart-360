package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/mondial7/smart-360/internal/models"
)

const (
	sessionCookieName = "s360_session"
	sessionTTL        = 7 * 24 * time.Hour
)

// issueSession creates a session row for userID and sets the signed session
// cookie. The cookie value is "<sessionID>.<hmac>"; the opaque session ID is a
// random UUID, and the HMAC lets us reject tampered/forged cookies before any
// database hit.
func (s *Service) issueSession(ctx context.Context, w http.ResponseWriter, userID string) error {
	session := &models.Session{UserID: userID, ExpiresAt: time.Now().Add(sessionTTL)}
	if err := s.repos.Sessions.Create(ctx, session); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    s.signSessionID(session.ID),
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookies(),
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return nil
}

// clearSession deletes the session row (if any) and expires the cookie.
func (s *Service) clearSession(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if id, ok := s.readSessionID(r); ok {
		_ = s.repos.Sessions.Delete(ctx, id)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookies(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// readSessionID extracts and verifies the session ID from the request cookie.
func (s *Service) readSessionID(r *http.Request) (string, bool) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", false
	}
	id, sig, ok := strings.Cut(c.Value, ".")
	if !ok {
		return "", false
	}
	if !hmac.Equal([]byte(sig), []byte(s.sign(id))) {
		return "", false
	}
	return id, true
}

// resolveUser turns a request's session cookie into the current user, or false
// when there is no valid, unexpired session.
func (s *Service) resolveUser(ctx context.Context, r *http.Request) (*models.User, bool) {
	id, ok := s.readSessionID(r)
	if !ok {
		return nil, false
	}
	session, err := s.repos.Sessions.FindByID(ctx, id)
	if err != nil {
		return nil, false
	}
	if time.Now().After(session.ExpiresAt) {
		_ = s.repos.Sessions.Delete(ctx, id)
		return nil, false
	}
	user, err := s.repos.Users.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, false
	}
	return user, true
}

func (s *Service) signSessionID(id string) string { return id + "." + s.sign(id) }

// sign returns a base64url HMAC-SHA256 of msg under the session secret.
func (s *Service) sign(msg string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// constantTimeEqual compares two strings without leaking length-independent
// timing. Used for CSRF token checks.
func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
