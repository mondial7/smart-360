package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/mondial7/smart-360/internal/models"
)

// RequestFeedbackForm lets a member ask for a feedback round on themselves.
func (h *Handlers) RequestFeedbackForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)

	// One open request at a time: if the member already has a draft round as the
	// subject, send them to it rather than letting requests pile up.
	if pending, _ := h.pendingRequestFor(ctx, u.ID); pending != nil {
		http.Redirect(w, r, "/my-feedback?pending=1", http.StatusSeeOther)
		return
	}

	// Everyone except the member can be a reviewer.
	all, err := h.Repos.Users.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	reviewers := make([]models.User, 0, len(all))
	for _, usr := range all {
		if usr.ID != u.ID {
			reviewers = append(reviewers, usr)
		}
	}
	templates, err := h.Repos.Templates.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	data := map[string]any{"Reviewers": reviewers, "Templates": templates}
	h.View.Page(w, http.StatusOK, h.page(r, "Request feedback", "my-feedback", "request_feedback_content", data))
}

// CreateFeedbackRequest creates a self-nominated round. The round is owned by a
// manager who is NOT the subject (the member's team admin, else a global admin),
// so the subject never gains the owner's raw-submission access. It starts as a
// draft for the owner to review and start.
func (h *Handlers) CreateFeedbackRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)

	if pending, _ := h.pendingRequestFor(ctx, u.ID); pending != nil {
		http.Redirect(w, r, "/my-feedback?pending=1", http.StatusSeeOther)
		return
	}

	owner, err := h.resolveOwnerFor(ctx, u)
	if err != nil {
		http.Error(w, "No manager is available to receive your request. Ask an admin to run a round for you.", http.StatusConflict)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	var templateID *string
	if t := r.FormValue("template_id"); t != "" {
		templateID = &t
	}
	round := &models.FeedbackRound{
		SubjectID:   u.ID,
		CreatedByID: owner.ID, // owner ≠ subject: preserves reviewer anonymity
		TemplateID:  templateID,
		Deadline:    parseDate(r.FormValue("deadline")),
		Status:      models.RoundDraft,
	}
	if err := h.Repos.Rounds.Create(ctx, round); err != nil {
		serverError(w, err)
		return
	}
	for _, reviewerID := range r.Form["reviewer_ids"] {
		if reviewerID != "" && reviewerID != u.ID {
			_ = h.Repos.Rounds.AddReviewer(ctx, round.ID, models.RoundReviewer{ReviewerID: reviewerID})
		}
	}
	// Actor is the requester; the round is attributed to them in the trail.
	h.audit(ctx, auditParams{Action: models.AuditRoundRequested, Actor: u, RoundID: round.ID,
		RoundSubject: u.Name, Description: "Requested a feedback round (owner: " + owner.Name + ")"})

	http.Redirect(w, r, "/my-feedback?requested=1", http.StatusSeeOther)
}

// pendingRequestFor returns the member's existing draft round-as-subject, if any.
func (h *Handlers) pendingRequestFor(ctx context.Context, memberID string) (*models.FeedbackRound, error) {
	rounds, err := h.Repos.Rounds.FindBySubjectID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	for i := range rounds {
		if rounds[i].Status == models.RoundDraft {
			return &rounds[i], nil
		}
	}
	return nil, nil
}

// resolveOwnerFor picks a manager to own a member's self-nominated round —
// never the member themselves. Prefers the member's team admin, then the
// configured ADMIN_EMAIL admin, then any other admin.
func (h *Handlers) resolveOwnerFor(ctx context.Context, member *models.User) (*models.User, error) {
	if member.TeamID != nil {
		if team, err := h.Repos.Teams.FindByID(ctx, *member.TeamID); err == nil &&
			team.TeamAdminID != "" && team.TeamAdminID != member.ID {
			if admin, err := h.Repos.Users.FindByID(ctx, team.TeamAdminID); err == nil {
				return admin, nil
			}
		}
	}
	users, err := h.Repos.Users.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var fallback *models.User
	for i := range users {
		if users[i].Role != models.RoleAdmin || users[i].ID == member.ID {
			continue
		}
		if h.Cfg.AdminEmail != "" && strings.EqualFold(users[i].Email, h.Cfg.AdminEmail) {
			return &users[i], nil // prefer the bootstrap owner
		}
		if fallback == nil {
			fallback = &users[i]
		}
	}
	if fallback != nil {
		return fallback, nil
	}
	return nil, errors.New("no eligible manager")
}
