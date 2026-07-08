package handlers_test

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/auth"
	"github.com/mondial7/smart-360/internal/config"
	"github.com/mondial7/smart-360/internal/handlers"
	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
	"github.com/mondial7/smart-360/internal/view"
	"github.com/mondial7/smart-360/web"
)

// newTestServer builds the full handler stack backed by in-memory fakes, so
// these tests exercise routing + middleware + templates without a database.
func newTestServer(t *testing.T) (*httptest.Server, *http.Client, repo.Repositories) {
	t.Helper()
	cfg := &config.Config{DevMode: true, SessionSecret: "test-secret"}
	repos := repo.NewFakes()
	renderer, err := view.NewRenderer(web.TemplatesFS)
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	authSvc := auth.New(cfg, repos)
	h := handlers.New(repos, authSvc, renderer, cfg)

	r := chi.NewRouter()
	r.Get("/login", h.LoginPage)
	r.Get("/auth/dev-login", authSvc.DevLogin)
	r.Group(func(r chi.Router) {
		r.Use(authSvc.RequireAuth)
		r.Use(authSvc.ProtectCSRF)
		r.Get("/", h.Dashboard)
		h.MountAppRoutes(r)
	})

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	return srv, client, repos
}

func get(t *testing.T, c *http.Client, url string) (int, string) {
	t.Helper()
	resp, err := c.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestUnauthenticatedRedirectsToLogin(t *testing.T) {
	srv, client, _ := newTestServer(t)
	// Don't follow redirects, so we can assert the 303 → /login.
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	resp, err := client.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/login" {
		t.Fatalf("expected redirect to /login, got %q", loc)
	}
}

func TestDevLoginThenDashboard(t *testing.T) {
	srv, client, repos := newTestServer(t)

	// dev-login provisions the first user as admin and sets the session cookie.
	if code, _ := get(t, client, srv.URL+"/auth/dev-login?email=admin@example.com"); code != http.StatusOK {
		t.Fatalf("dev-login final status %d", code)
	}
	if u, err := repos.Users.FindByEmail(t.Context(), "admin@example.com"); err != nil || u.Role != "admin" {
		t.Fatalf("expected admin user provisioned, got %+v err=%v", u, err)
	}

	code, body := get(t, client, srv.URL+"/")
	if code != http.StatusOK {
		t.Fatalf("dashboard status %d", code)
	}
	if !strings.Contains(body, "<h1>Dashboard</h1>") {
		t.Fatalf("dashboard did not render expected heading")
	}

	// Admin-only pages render for an admin.
	for _, path := range []string{"/rounds", "/rounds/new", "/teams", "/analytics", "/audit-logs", "/my-feedback", "/team"} {
		if code, _ := get(t, client, srv.URL+path); code != http.StatusOK {
			t.Fatalf("GET %s: expected 200, got %d", path, code)
		}
	}
}

func TestRoundOwnerSeesRawSubmissionsButReviewerDoesNot(t *testing.T) {
	srv, client, repos := newTestServer(t)
	ctx := t.Context()

	// admin (round owner) logs in.
	_, _ = get(t, client, srv.URL+"/auth/dev-login?email=admin@example.com")
	admin, _ := repos.Users.FindByEmail(ctx, "admin@example.com")

	subject := &models.User{Email: "subject@example.com", Name: "Subject"}
	_ = repos.Users.Create(ctx, subject)
	reviewer := &models.User{Email: "rev@example.com", Name: "Reviewer Rita"}
	_ = repos.Users.Create(ctx, reviewer)

	round := &models.FeedbackRound{SubjectID: subject.ID, CreatedByID: admin.ID, Status: models.RoundClosed}
	_ = repos.Rounds.Create(ctx, round)
	_ = repos.Rounds.AddReviewer(ctx, round.ID, models.RoundReviewer{ReviewerID: reviewer.ID})
	_ = repos.Submissions.Create(ctx, &models.Submission{
		RoundID:      round.ID,
		ReviewerID:   reviewer.ID,
		Relationship: models.RelationshipPeer,
		Responses:    map[string]string{"a": "unblocks the team constantly"},
		Ratings:      []models.CompetencyRating{{Key: "execution", Score: 4, Justification: "ships reliably"}},
		PrivateNotes: "confidential note for the manager",
	})

	// Owner sees the raw content.
	code, body := get(t, client, srv.URL+"/rounds/"+round.ID)
	if code != http.StatusOK {
		t.Fatalf("owner round details: %d", code)
	}
	for _, want := range []string{"Feedback submissions", "unblocks the team constantly", "ships reliably", "confidential note for the manager"} {
		if !strings.Contains(body, want) {
			t.Fatalf("owner should see %q in round details", want)
		}
	}

	// The reviewer (a participant but not the round owner) must not see the raw
	// submissions section or anyone's private notes on the round details page.
	jar2, _ := cookiejar.New(nil)
	rita := &http.Client{Jar: jar2}
	if code, _ := get(t, rita, srv.URL+"/auth/dev-login?email=rev@example.com"); code != http.StatusOK {
		t.Fatalf("reviewer dev-login: %d", code)
	}
	code, body = get(t, rita, srv.URL+"/rounds/"+round.ID)
	if code != http.StatusOK {
		t.Fatalf("reviewer round details: %d", code)
	}
	if strings.Contains(body, "confidential note for the manager") || strings.Contains(body, "Feedback submissions") {
		t.Fatal("a non-owner must not see raw submissions or private notes")
	}
}

func TestNonAdminCannotAccessTeams(t *testing.T) {
	srv, client, repos := newTestServer(t)
	// First user (admin) provisioned separately so the second is a member.
	_, _ = get(t, client, srv.URL+"/auth/dev-login?email=admin@example.com")

	jar2, _ := cookiejar.New(nil)
	member := &http.Client{Jar: jar2}
	if code, _ := get(t, member, srv.URL+"/auth/dev-login?email=member@example.com"); code != http.StatusOK {
		t.Fatalf("member dev-login: %d", code)
	}
	if u, _ := repos.Users.FindByEmail(t.Context(), "member@example.com"); u.Role != "member" {
		t.Fatalf("expected member role, got %q", u.Role)
	}
	if code, _ := get(t, member, srv.URL+"/teams"); code != http.StatusForbidden {
		t.Fatalf("expected 403 for member on /teams, got %d", code)
	}
}
