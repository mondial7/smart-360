package handlers

import (
	"context"
	"html"
	"net/http"
	"time"

	"github.com/mondial7/smart-360/internal/models"
)

// htmlEscape escapes text for safe interpolation into server-built HTML strings
// (e.g. SSE event payloads that don't go through html/template).
func htmlEscape(s string) string { return html.EscapeString(s) }

// redirect navigates the browser: htmx requests get an HX-Redirect header (so
// the client does a full navigation), everyone else a 303.
func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
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
