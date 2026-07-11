package handlers

import (
	"context"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mondial7/smart-360/internal/models"
)

// htmlEscape escapes text for safe interpolation into server-built HTML strings
// (e.g. SSE event payloads that don't go through html/template).
func htmlEscape(s string) string { return html.EscapeString(s) }

// pageSize is the number of rows per page on paginated lists.
const pageSize = 25

// pageNav is the view-model for the pagination controls.
type pageNav struct {
	Page    int
	HasPrev bool
	HasNext bool
	PrevURL string
	NextURL string
}

// pageParam reads the 1-based ?page query parameter (min 1).
func pageParam(r *http.Request) int {
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if p < 1 {
		return 1
	}
	return p
}

// buildPageNav builds Prev/Next controls, preserving any other query params
// (e.g. the audit action filter). hasNext is determined by fetching pageSize+1
// rows and checking for the extra one.
func buildPageNav(r *http.Request, page int, hasNext bool) pageNav {
	nav := pageNav{Page: page, HasPrev: page > 1, HasNext: hasNext}
	if nav.HasPrev {
		nav.PrevURL = pageURL(r, page-1)
	}
	if nav.HasNext {
		nav.NextURL = pageURL(r, page+1)
	}
	return nav
}

func pageURL(r *http.Request, page int) string {
	q := r.URL.Query()
	q.Set("page", strconv.Itoa(page))
	return r.URL.Path + "?" + q.Encode()
}

// paginate trims a slice fetched with one extra row (pageSize+1) to the page
// size and reports whether a next page exists.
func paginate[T any](rows []T) (page []T, hasNext bool) {
	if len(rows) > pageSize {
		return rows[:pageSize], true
	}
	return rows, false
}

// redirect navigates the browser: htmx requests get an HX-Redirect header (so
// the client does a full navigation), everyone else a 303. The target is forced
// to a same-origin absolute path so it can never become an open redirect.
func redirect(w http.ResponseWriter, r *http.Request, url string) {
	url = safeLocalPath(url)
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther) // #nosec G710 -- url is forced to a same-origin path by safeLocalPath above
}

// safeLocalPath returns p only if it is a same-origin absolute path; anything
// else (empty, scheme-relative "//host", or an absolute URL) collapses to "/".
func safeLocalPath(p string) string {
	if p == "" || p[0] != '/' || strings.HasPrefix(p, "//") {
		return "/"
	}
	return p
}

// parseDate parses an <input type=date> value ("2006-01-02") into a time, or
// nil when empty/invalid.
func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

// auditParams carries the fields for an audit log entry.
type auditParams struct {
	Action       models.AuditAction
	Actor        *models.User
	RoundID      string
	RoundSubject string
	TeamID       string
	TeamName     string
	Description  string
	OldValue     string
	NewValue     string
}

// audit writes an audit log entry, best-effort (a logging failure never fails
// the operation it records).
func (h *Handlers) audit(ctx context.Context, p auditParams) {
	entry := &models.AuditLog{
		Action:       p.Action,
		ActorID:      p.Actor.ID,
		ActorName:    p.Actor.Name,
		ActorEmail:   p.Actor.Email,
		RoundSubject: p.RoundSubject,
		TeamName:     p.TeamName,
		Description:  p.Description,
		OldValue:     p.OldValue,
		NewValue:     p.NewValue,
	}
	if p.RoundID != "" {
		entry.RoundID = &p.RoundID
	}
	if p.TeamID != "" {
		entry.TeamID = &p.TeamID
	}
	_ = h.Repos.Audit.Create(ctx, entry)
}
