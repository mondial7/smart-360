package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mondial7/smart-360/internal/config"
	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
)

func newTestService(t *testing.T) (*Service, repo.Repositories) {
	t.Helper()
	cfg := &config.Config{SessionSecret: "test-secret-please-change", DevMode: true}
	repos := repo.NewFakes()
	return New(cfg, repos), repos
}

// issueCookie runs issueSession and returns the resulting session cookie.
func issueCookie(t *testing.T, s *Service, userID string) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	if err := s.issueSession(context.Background(), rec, userID); err != nil {
		t.Fatalf("issueSession: %v", err)
	}
	res := rec.Result()
	for _, c := range res.Cookies() {
		if c.Name == sessionCookieName {
			return c
		}
	}
	t.Fatal("session cookie not set")
	return nil
}

func TestSessionRoundTrip(t *testing.T) {
	s, repos := newTestService(t)
	u := &models.User{Email: "a@example.com", Name: "Ann", Role: models.RoleAdmin}
	if err := repos.Users.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	cookie := issueCookie(t, s, u.ID)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	got, ok := s.resolveUser(req.Context(), req)
	if !ok {
		t.Fatal("expected to resolve a user from the session cookie")
	}
	if got.ID != u.ID || got.Role != models.RoleAdmin {
		t.Fatalf("resolved wrong user: %+v", got)
	}
}

func TestSessionRejectsTamperedCookie(t *testing.T) {
	s, repos := newTestService(t)
	u := &models.User{Email: "b@example.com"}
	_ = repos.Users.Create(context.Background(), u)
	cookie := issueCookie(t, s, u.ID)

	// Flip the signature.
	cookie.Value = cookie.Value + "x"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	if _, ok := s.resolveUser(req.Context(), req); ok {
		t.Fatal("expected tampered cookie to be rejected")
	}
}

func TestRequireAuthRedirectsWhenAnonymous(t *testing.T) {
	s, _ := newTestService(t)
	handlerRan := false
	h := s.RequireAuth(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { handlerRan = true }))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))

	if handlerRan {
		t.Fatal("protected handler should not run for anonymous request")
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Fatalf("expected redirect to /login, got %q", loc)
	}
}

func TestCSRFProtection(t *testing.T) {
	s, repos := newTestService(t)
	u := &models.User{Email: "c@example.com"}
	_ = repos.Users.Create(context.Background(), u)
	cookie := issueCookie(t, s, u.ID)

	// Token derived from the session.
	tokReq := httptest.NewRequest(http.MethodGet, "/", nil)
	tokReq.AddCookie(cookie)
	token := s.CSRFToken(tokReq)
	if token == "" {
		t.Fatal("expected a CSRF token for an authenticated request")
	}

	ran := false
	h := s.ProtectCSRF(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { ran = true }))

	// Missing token → rejected.
	rec := httptest.NewRecorder()
	bad := httptest.NewRequest(http.MethodPost, "/rounds", nil)
	bad.AddCookie(cookie)
	h.ServeHTTP(rec, bad)
	if ran || rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing CSRF token, got %d (ran=%v)", rec.Code, ran)
	}

	// Correct token in header → allowed.
	ran = false
	rec = httptest.NewRecorder()
	good := httptest.NewRequest(http.MethodPost, "/rounds", nil)
	good.AddCookie(cookie)
	good.Header.Set(csrfHeader, token)
	h.ServeHTTP(rec, good)
	if !ran {
		t.Fatalf("expected handler to run with valid CSRF token, got %d", rec.Code)
	}

	// GET is exempt.
	ran = false
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/rounds", nil))
	if !ran {
		t.Fatal("GET should bypass CSRF check")
	}
}
