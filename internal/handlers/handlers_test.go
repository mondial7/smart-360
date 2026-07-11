package handlers_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/auth"
	"github.com/mondial7/smart-360/internal/config"
	"github.com/mondial7/smart-360/internal/handlers"
	"github.com/mondial7/smart-360/internal/logstream"
	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
	"github.com/mondial7/smart-360/internal/view"
	"github.com/mondial7/smart-360/web"
)

// newTestServer builds the full handler stack backed by in-memory fakes, so
// these tests exercise routing + middleware + templates without a database.
func newTestServer(t *testing.T) (*httptest.Server, *http.Client, repo.Repositories) {
	t.Helper()
	cfg := &config.Config{DevMode: true, SessionSecret: "test-secret", AdminEmail: "admin@example.com"}
	repos := repo.NewFakes()
	renderer, err := view.NewRenderer(web.TemplatesFS)
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	authSvc := auth.New(cfg, repos)
	h := handlers.New(repos, authSvc, renderer, cfg, logstream.New(50))

	r := chi.NewRouter()
	r.Get("/login", h.LoginPage)
	r.Get("/auth/dev-login", authSvc.DevLogin)
	r.Group(func(r chi.Router) {
		r.Use(authSvc.RequireAuth)
		r.Use(authSvc.ProtectCSRF)
		r.Get("/", h.Dashboard)
		// No-op rate limiter for tests.
		h.MountAppRoutes(r, func(next http.Handler) http.Handler { return next })
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

// csrfToken pulls the per-session CSRF token from a rendered page's meta tag.
func csrfToken(t *testing.T, c *http.Client, base string) string {
	t.Helper()
	_, body := get(t, c, base+"/my-feedback")
	m := regexp.MustCompile(`name="csrf-token" content="([^"]*)"`).FindStringSubmatch(body)
	if m == nil {
		t.Fatal("no csrf-token meta on page")
	}
	return m[1]
}

// postForm submits a form with the CSRF token in the header (as htmx does).
func postForm(t *testing.T, c *http.Client, base, path, token string, form url.Values) (int, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, base+path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CSRF-Token", token)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestSelfNominationOwnedByManagerNotSubject(t *testing.T) {
	srv, admin, repos := newTestServer(t)
	ctx := t.Context()
	// admin@example.com matches ADMIN_EMAIL → the eligible owner.
	_, _ = get(t, admin, srv.URL+"/auth/dev-login?email=admin@example.com")
	adminUser, _ := repos.Users.FindByEmail(ctx, "admin@example.com")

	// A member requests feedback on themselves.
	jar, _ := cookiejar.New(nil)
	member := &http.Client{Jar: jar}
	member.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	_, _ = get(t, member, srv.URL+"/auth/dev-login?email=mia@example.com")
	memberUser, _ := repos.Users.FindByEmail(ctx, "mia@example.com")
	token := csrfToken(t, member, srv.URL)

	code, _ := postForm(t, member, srv.URL, "/request-feedback", token, url.Values{"reviewer_ids": {adminUser.ID}})
	if code != http.StatusSeeOther {
		t.Fatalf("expected 303 after request, got %d", code)
	}

	rounds, _ := repos.Rounds.FindBySubjectID(ctx, memberUser.ID)
	if len(rounds) != 1 {
		t.Fatalf("expected exactly one requested round, got %d", len(rounds))
	}
	rd := rounds[0]
	if rd.SubjectID != memberUser.ID {
		t.Fatalf("subject should be the member")
	}
	// The invariant: the owner (creator, who gets raw-submission access) is NOT
	// the subject — otherwise the member could de-anonymize their reviewers.
	if rd.CreatedByID == memberUser.ID {
		t.Fatal("SECURITY: self-nominated round must not be owned by its subject")
	}
	if rd.CreatedByID != adminUser.ID {
		t.Fatalf("expected the admin to own the round, got %q", rd.CreatedByID)
	}
	if rd.Status != models.RoundDraft {
		t.Fatalf("expected draft (awaiting owner approval), got %q", rd.Status)
	}

	// A second request is blocked while one is pending.
	code, _ = postForm(t, member, srv.URL, "/request-feedback", token, url.Values{"reviewer_ids": {adminUser.ID}})
	if code != http.StatusSeeOther {
		t.Fatalf("expected redirect on duplicate request, got %d", code)
	}
	if again, _ := repos.Rounds.FindBySubjectID(ctx, memberUser.ID); len(again) != 1 {
		t.Fatalf("duplicate request should not create a second round; got %d", len(again))
	}
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

	// dev-login provisions the ADMIN_EMAIL user as admin and sets the session cookie.
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

func TestUsersListPaginates(t *testing.T) {
	srv, client, repos := newTestServer(t)
	ctx := t.Context()
	_, _ = get(t, client, srv.URL+"/auth/dev-login?email=admin@example.com")

	// Create enough users to span two pages (pageSize is 25).
	for i := 0; i < 30; i++ {
		_ = repos.Users.Create(ctx, &models.User{Email: fmt.Sprintf("u%02d@example.com", i), Name: fmt.Sprintf("User %02d", i)})
	}

	// Page 1 offers a Next link; no Prev.
	_, body := get(t, client, srv.URL+"/users")
	if !strings.Contains(body, "page=2") {
		t.Fatal("page 1 should link to page 2 (Next)")
	}
	if strings.Contains(body, "page=0") {
		t.Fatal("page 1 should not offer a Prev link")
	}

	// Page 2 offers a Prev link back to page 1.
	_, body = get(t, client, srv.URL+"/users?page=2")
	if !strings.Contains(body, "page=1") {
		t.Fatal("page 2 should link back to page 1 (Prev)")
	}
}

func TestOnboardingShownUntilCompleted(t *testing.T) {
	srv, client, repos := newTestServer(t)
	ctx := t.Context()
	if code, _ := get(t, client, srv.URL+"/auth/dev-login?email=newbie@example.com"); code != http.StatusOK {
		t.Fatalf("dev-login: %d", code)
	}
	user, _ := repos.Users.FindByEmail(ctx, "newbie@example.com")

	// First login: the tour overlay is present.
	_, body := get(t, client, srv.URL+"/")
	if !strings.Contains(body, `id="onboarding"`) {
		t.Fatal("expected onboarding overlay on first login")
	}

	// After completing (marked seen), it no longer renders.
	if err := repos.Users.MarkOnboarded(ctx, user.ID); err != nil {
		t.Fatalf("mark onboarded: %v", err)
	}
	_, body = get(t, client, srv.URL+"/")
	if strings.Contains(body, `id="onboarding"`) {
		t.Fatal("onboarding overlay should not render after completion")
	}
}

func TestConsolidationStreamRequiresToken(t *testing.T) {
	srv, client, repos := newTestServer(t)
	ctx := t.Context()
	_, _ = get(t, client, srv.URL+"/auth/dev-login?email=admin@example.com")
	admin, _ := repos.Users.FindByEmail(ctx, "admin@example.com")
	round := &models.FeedbackRound{SubjectID: admin.ID, CreatedByID: admin.ID, Status: models.RoundClosed}
	_ = repos.Rounds.Create(ctx, round)

	// The SSE stream is a state-changing GET; without the session-derived token
	// it must be refused (CSRF defense), even for an authenticated admin.
	code, _ := get(t, client, srv.URL+"/rounds/"+round.ID+"/consolidate/stream")
	if code != http.StatusForbidden {
		t.Fatalf("stream without token: expected 403, got %d", code)
	}
	code, _ = get(t, client, srv.URL+"/rounds/"+round.ID+"/consolidate/stream?t=bogus")
	if code != http.StatusForbidden {
		t.Fatalf("stream with bad token: expected 403, got %d", code)
	}
}

func TestNonAdminCannotAccessTeams(t *testing.T) {
	srv, client, repos := newTestServer(t)
	// admin@example.com matches ADMIN_EMAIL → admin; member@example.com → member.
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
