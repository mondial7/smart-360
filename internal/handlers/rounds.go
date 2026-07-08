package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
)

// RoundsList shows all rounds (admin) or the rounds relevant to the user.
func (h *Handlers) RoundsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)

	users, err := h.allUsersIndex(ctx)
	if err != nil {
		serverError(w, err)
		return
	}

	var rounds []models.FeedbackRound
	if u.Role == models.RoleAdmin {
		rounds, err = h.Repos.Rounds.FindAll(ctx)
	} else {
		rounds, err = h.roundsForMe(ctx, u.ID)
	}
	if err != nil {
		serverError(w, err)
		return
	}

	cards := make([]RoundCard, 0, len(rounds))
	for _, rd := range rounds {
		cards = append(cards, h.toCard(ctx, rd, u.ID, users))
	}

	data := map[string]any{"Cards": cards, "CanCreate": u.Role == models.RoleAdmin || u.Role == models.RoleTeamAdmin}
	h.View.Page(w, http.StatusOK, h.page(r, "Rounds", "rounds", "rounds_content", data))
}

// NewRoundForm renders the round creation form.
func (h *Handlers) NewRoundForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := h.Repos.Users.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	templates, err := h.Repos.Templates.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	data := map[string]any{"Users": users, "Templates": templates}
	h.View.Page(w, http.StatusOK, h.page(r, "New round", "rounds", "round_new_content", data))
}

// CreateRound creates a draft round with the chosen subject, template, deadline,
// and reviewers.
func (h *Handlers) CreateRound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	subjectID := r.FormValue("subject_id")
	if subjectID == "" {
		http.Error(w, "A subject is required", http.StatusBadRequest)
		return
	}
	var templateID *string
	if t := r.FormValue("template_id"); t != "" {
		templateID = &t
	}

	round := &models.FeedbackRound{
		SubjectID:   subjectID,
		CreatedByID: u.ID,
		TemplateID:  templateID,
		Deadline:    parseDate(r.FormValue("deadline")),
		Status:      models.RoundDraft,
	}
	if err := h.Repos.Rounds.Create(ctx, round); err != nil {
		serverError(w, err)
		return
	}
	for _, reviewerID := range r.Form["reviewer_ids"] {
		if reviewerID == subjectID || reviewerID == "" {
			continue // the subject participates via self-assessment, not as a reviewer
		}
		_ = h.Repos.Rounds.AddReviewer(ctx, round.ID, models.RoundReviewer{ReviewerID: reviewerID})
	}

	subject, _ := h.Repos.Users.FindByID(ctx, subjectID)
	h.audit(ctx, auditParams{Action: models.AuditRoundCreated, Actor: u, RoundID: round.ID,
		RoundSubject: nameOf(subject), Description: "Created feedback round"})

	redirect(w, r, "/rounds/"+round.ID)
}

// RoundDetails shows a round with its reviewers, submissions (for managers), and
// lifecycle actions.
func (h *Handlers) RoundDetails(w http.ResponseWriter, r *http.Request) {
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

	users, err := h.allUsersIndex(ctx)
	if err != nil {
		serverError(w, err)
		return
	}

	isManager := u.Role == models.RoleAdmin || u.ID == round.CreatedByID

	type reviewerRow struct {
		Name      string
		Submitted bool
	}
	var reviewers []reviewerRow
	submittedCount := 0
	for _, rev := range round.Reviewers {
		n, _ := h.Repos.Submissions.CountByRoundAndReviewer(ctx, round.ID, rev.ReviewerID)
		if n > 0 {
			submittedCount++
		}
		reviewers = append(reviewers, reviewerRow{Name: users[rev.ReviewerID].Name, Submitted: n > 0})
	}

	// The round owner (and global admins) can read every raw submission. The AI
	// consolidation is an aid, not a replacement for the manager's own reading
	// of the feedback — so we surface the full content here, including reviewer
	// identity (managers see who said what for accountability) and the private
	// manager-only notes.
	var submissions []submissionDetail
	if isManager {
		tmpl, err := h.resolveTemplate(r, round.TemplateID)
		if err != nil {
			serverError(w, err)
			return
		}
		raw, err := h.Repos.Submissions.FindByRoundID(ctx, round.ID)
		if err != nil {
			serverError(w, err)
			return
		}
		submissions = buildSubmissionDetails(raw, tmpl, users)
	}

	_, consErr := h.Repos.Consolidations.FindByRoundID(ctx, round.ID)
	data := map[string]any{
		"Round":            round,
		"SubjectName":      users[round.SubjectID].Name,
		"CreatorName":      users[round.CreatedByID].Name,
		"Reviewers":        reviewers,
		"Submissions":      submissions,
		"SubmittedCount":   submittedCount,
		"IsManager":        isManager,
		"IsAdmin":          u.Role == models.RoleAdmin,
		"Status":           string(round.Status),
		"HasConsolidation": consErr == nil,
	}
	h.View.Page(w, http.StatusOK, h.page(r, "Round details", "rounds", "round_details_content", data))
}

// submissionDetail is the manager-facing view of one raw submission.
type submissionDetail struct {
	ReviewerName string
	IsSelf       bool
	Relationship string
	Frequency    string
	Answers      []answerRow
	Ratings      []ratingRow
	PrivateNotes string
	SubmittedAt  time.Time
}

type answerRow struct {
	Label string
	Text  string
}

type ratingRow struct {
	Name          string
	Score         int
	Justification string
}

// buildSubmissionDetails turns raw submissions into manager-facing view models,
// resolving reviewer names, question labels, and competency names from the
// template, and ordering answers by the template's question order.
func buildSubmissionDetails(subs []models.Submission, tmpl *models.Template, users map[string]models.User) []submissionDetail {
	// Question label + order.
	type q struct{ key, label string }
	var questions []q
	if tmpl != nil && len(tmpl.Questions) > 0 {
		for _, tq := range tmpl.Questions {
			label := tq.CardTitle
			if label == "" {
				label = tq.Key
			}
			questions = append(questions, q{tq.Key, label})
		}
	} else {
		for _, k := range []string{"a", "b", "c", "d"} {
			questions = append(questions, q{k, k})
		}
	}
	compName := map[string]string{}
	if tmpl != nil {
		for _, c := range tmpl.Competencies {
			compName[c.Key] = c.Name
		}
	}

	out := make([]submissionDetail, 0, len(subs))
	for _, s := range subs {
		d := submissionDetail{
			ReviewerName: users[s.ReviewerID].Name,
			IsSelf:       s.IsSelf,
			PrivateNotes: s.PrivateNotes,
			SubmittedAt:  s.SubmittedAt,
		}
		if !s.IsSelf {
			d.Relationship = relationshipText(s.Relationship)
			d.Frequency = frequencyText(s.InteractionFrequency)
		}
		for _, qq := range questions {
			if text := s.Responses[qq.key]; text != "" {
				d.Answers = append(d.Answers, answerRow{Label: qq.label, Text: text})
			}
		}
		for _, rt := range s.Ratings {
			name := compName[rt.Key]
			if name == "" {
				name = rt.Key
			}
			d.Ratings = append(d.Ratings, ratingRow{Name: name, Score: rt.Score, Justification: rt.Justification})
		}
		out = append(out, d)
	}
	return out
}

func relationshipText(r models.ReviewerRelationship) string {
	switch r {
	case models.RelationshipManager:
		return "Manager"
	case models.RelationshipReport:
		return "Direct report"
	case models.RelationshipPeer:
		return "Peer"
	case models.RelationshipCrossFunctional:
		return "Cross-functional"
	}
	return "Unspecified"
}

func frequencyText(f models.InteractionFrequency) string {
	switch f {
	case models.InteractionDaily:
		return "Daily"
	case models.InteractionWeekly:
		return "Weekly"
	case models.InteractionMonthly:
		return "Monthly"
	case models.InteractionRarely:
		return "Rarely"
	}
	return ""
}

// StartRound transitions a draft round to active.
func (h *Handlers) StartRound(w http.ResponseWriter, r *http.Request) {
	h.transitionRound(w, r, models.RoundDraft, models.RoundActive, models.AuditRoundStatusChanged)
}

// CloseRound transitions an active round to closed.
func (h *Handlers) CloseRound(w http.ResponseWriter, r *http.Request) {
	h.transitionRound(w, r, models.RoundActive, models.RoundClosed, models.AuditRoundStatusChanged)
}

func (h *Handlers) transitionRound(w http.ResponseWriter, r *http.Request, from, to models.RoundStatus, action models.AuditAction) {
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
	if u.Role != models.RoleAdmin && u.ID != round.CreatedByID {
		forbidden(w)
		return
	}
	if round.Status != from {
		http.Error(w, "Round is not in the required state for this action", http.StatusBadRequest)
		return
	}
	if err := h.Repos.Rounds.UpdateStatus(ctx, id, to); err != nil {
		serverError(w, err)
		return
	}
	subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
	h.audit(ctx, auditParams{Action: action, Actor: u, RoundID: id, RoundSubject: nameOf(subject),
		Description: "Round status changed", OldValue: string(from), NewValue: string(to)})

	redirect(w, r, "/rounds/"+id)
}

func nameOf(u *models.User) string {
	if u == nil {
		return ""
	}
	return u.Name
}
