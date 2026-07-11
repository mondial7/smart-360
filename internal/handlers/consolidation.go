package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/ai"
	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
)

// ConsolidationView shows the consolidation, or (for a manager on a closed
// round without one yet) the generate panel that streams progress over SSE.
func (h *Handlers) ConsolidationView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	id := chi.URLParam(r, "id")

	round, err := h.Repos.Rounds.FindByID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	if !canAccessConsolidation(u, *round) {
		forbidden(w)
		return
	}
	isManager := u.Role == models.RoleAdmin || u.ID == round.CreatedByID

	cons, err := h.Repos.Consolidations.FindByRoundID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		data := map[string]any{
			"Round":       round,
			"IsManager":   isManager,
			"CanGenerate": isManager && round.Status == models.RoundClosed,
		}
		h.View.Page(w, http.StatusOK, h.page(r, "Consolidation", "rounds", "consolidation_empty_content", data))
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}

	// The subject may only see a shared consolidation.
	if u.ID == round.SubjectID && u.Role != models.RoleAdmin && cons.SharedAt == nil {
		http.Error(w, "This consolidation has not been shared yet", http.StatusForbidden)
		return
	}
	// Never expose the manager-only channel to anyone but the manager/admin.
	if !canSeeManagerOnlyChannel(u, *round) {
		cons.ManagerOnlyChannel = nil
	}

	subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
	data := map[string]any{
		"Round":       round,
		"Cons":        cons,
		"SubjectName": nameOf(subject),
		"IsManager":   isManager,
		"CanShare":    isManager && round.Status == models.RoundClosed && cons.SharedAt == nil,
		"Shared":      cons.SharedAt != nil,
	}
	h.View.Page(w, http.StatusOK, h.page(r, "Consolidation", "rounds", "consolidation_content", data))
}

// StartConsolidation (POST) returns the SSE panel that will stream progress.
// The actual generation runs in ConsolidationStream. CSRF-protected + admin.
func (h *Handlers) StartConsolidation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	round, err := h.Repos.Rounds.FindByID(r.Context(), id)
	if err != nil {
		notFound(w)
		return
	}
	if round.Status != models.RoundClosed {
		http.Error(w, "The round must be closed before consolidating", http.StatusBadRequest)
		return
	}
	h.View.Fragment(w, http.StatusOK, "consolidation_stream_panel",
		map[string]any{"RoundID": id, "Token": h.Auth.StreamToken(r)})
}

// ConsolidationStream runs the consolidation pipeline and streams progress as
// Server-Sent Events, then persists the result. Admin only.
func (h *Handlers) ConsolidationStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	id := chi.URLParam(r, "id")

	// The SSE stream generates and persists the consolidation, so it is a
	// state-changing GET. GETs bypass the CSRF middleware and SameSite=Lax lets
	// a cross-site top-level navigation carry the session cookie — so require a
	// session-derived token that a cross-site attacker cannot know.
	if !h.Auth.ValidStreamToken(r, r.URL.Query().Get("t")) {
		forbidden(w)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	round, err := h.Repos.Rounds.FindByID(ctx, id)
	if err != nil || round.Status != models.RoundClosed {
		http.Error(w, "round not ready for consolidation", http.StatusBadRequest)
		return
	}
	submissions, err := h.Repos.Submissions.FindByRoundID(ctx, id)
	if err != nil {
		http.Error(w, "failed to load submissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if len(submissions) == 0 {
		writeSSE(w, flusher, "done", `<div class="card form-error">No submissions to consolidate.</div>`)
		return
	}

	tmpl, err := h.resolveTemplate(r, round.TemplateID)
	if err != nil {
		writeSSE(w, flusher, "done", `<div class="card form-error">Failed to load the round template.</div>`)
		return
	}

	// Decouple the pipeline (which emits progress from several goroutines) from
	// the single writer via a channel.
	type sseMsg struct{ event, data string }
	msgs := make(chan sseMsg, 32)

	go func() {
		defer close(msgs)
		cons, logs, gerr := ai.Consolidate(ctx, ai.Options{
			Submissions:   submissions,
			Template:      tmpl,
			RoundID:       id,
			GeneratedByID: u.ID,
			APIKey:        h.Cfg.GeminiAPIKey,
			Progress: func(ev ai.ProgressEvent) {
				msgs <- sseMsg{"progress", progressHTML(ev)}
			},
		})
		if gerr != nil {
			msgs <- sseMsg{"done", `<div class="card form-error">Consolidation failed. Please try again.</div>`}
			return
		}
		// Persist moderation audit logs (best-effort), then the consolidation.
		for i := range logs {
			_ = h.Repos.Moderation.Create(ctx, &logs[i])
		}
		if existing, ferr := h.Repos.Consolidations.FindByRoundID(ctx, id); ferr == nil {
			cons.ID = existing.ID
			err = h.Repos.Consolidations.Update(ctx, &cons)
		} else {
			err = h.Repos.Consolidations.Create(ctx, &cons)
		}
		if err != nil {
			msgs <- sseMsg{"done", `<div class="card form-error">Failed to save the consolidation.</div>`}
			return
		}
		h.audit(ctx, auditParams{Action: models.AuditConsolidationCreated, Actor: u, RoundID: id,
			Description: "Generated consolidation"})
		// No inline script (CSP script-src 'self'): app.js navigates on [data-redirect].
		msgs <- sseMsg{"done", fmt.Sprintf(
			`<div class="card" data-redirect="/rounds/%s/consolidation"><strong>Consolidation ready.</strong> <a class="btn btn--sm" href="/rounds/%s/consolidation">View feedback</a></div>`,
			id, id)}
	}()

	for m := range msgs {
		writeSSE(w, flusher, m.event, m.data)
		if r.Context().Err() != nil {
			return
		}
	}
}

// ShareConsolidation flips the round to shared and stamps the consolidation.
func (h *Handlers) ShareConsolidation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	id := chi.URLParam(r, "id")

	cons, err := h.Repos.Consolidations.FindByID(ctx, id)
	if err != nil {
		notFound(w)
		return
	}
	round, err := h.Repos.Rounds.FindByID(ctx, cons.RoundID)
	if err != nil {
		notFound(w)
		return
	}
	if u.Role != models.RoleAdmin && u.ID != round.CreatedByID {
		forbidden(w)
		return
	}
	if round.Status != models.RoundClosed {
		http.Error(w, "The round must be closed to share", http.StatusBadRequest)
		return
	}

	now := time.Now()
	cons.SharedAt = &now
	if err := h.Repos.Consolidations.Update(ctx, cons); err != nil {
		serverError(w, err)
		return
	}
	if err := h.Repos.Rounds.UpdateStatus(ctx, round.ID, models.RoundShared); err != nil {
		serverError(w, err)
		return
	}
	subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
	h.audit(ctx, auditParams{Action: models.AuditConsolidationShared, Actor: u, RoundID: round.ID,
		RoundSubject: nameOf(subject), Description: "Shared consolidation with subject",
		OldValue: string(models.RoundClosed), NewValue: string(models.RoundShared)})

	redirect(w, r, "/rounds/"+round.ID+"/consolidation")
}

// UpdateConsolidation lets the round owner correct the AI-generated sections
// and record private admin notes. The AI output is a starting point, not the
// final word — the manager owns the round.
func (h *Handlers) UpdateConsolidation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	id := chi.URLParam(r, "id")

	cons, err := h.Repos.Consolidations.FindByID(ctx, id)
	if err != nil {
		notFound(w)
		return
	}
	round, err := h.Repos.Rounds.FindByID(ctx, cons.RoundID)
	if err != nil {
		notFound(w)
		return
	}
	if u.Role != models.RoleAdmin && u.ID != round.CreatedByID {
		forbidden(w)
		return
	}

	_ = r.ParseForm()
	cons.ExecutiveSummary = strings.TrimSpace(r.FormValue("executive_summary"))
	cons.Strengths = splitLines(r.FormValue("strengths"))
	cons.AreasForImprovement = splitLines(r.FormValue("areas_for_improvement"))
	cons.ActionableInsights = splitLines(r.FormValue("actionable_insights"))
	cons.AdminNotes = strings.TrimSpace(r.FormValue("admin_notes"))

	if err := h.Repos.Consolidations.Update(ctx, cons); err != nil {
		serverError(w, err)
		return
	}
	subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
	h.audit(ctx, auditParams{Action: models.AuditConsolidationEdited, Actor: u, RoundID: round.ID,
		RoundSubject: nameOf(subject), Description: "Edited consolidation"})

	redirect(w, r, "/rounds/"+round.ID+"/consolidation")
}

// splitLines turns a textarea value into a trimmed, non-empty slice of lines.
func splitLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			out = append(out, t)
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}

// progressHTML renders the inner content for the SSE progress element.
func progressHTML(ev ai.ProgressEvent) string {
	msg := ev.Message
	if msg == "" {
		msg = ev.Stage
	}
	return `<span class="spinner"></span><span>` + htmlEscape(msg) + `</span>`
}

// writeSSE writes one Server-Sent Event. Data is collapsed to a single line so
// it stays within one `data:` field.
func writeSSE(w http.ResponseWriter, f http.Flusher, event, data string) {
	data = strings.ReplaceAll(data, "\n", " ")
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	f.Flush()
}
