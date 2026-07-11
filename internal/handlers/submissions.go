package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/ai"
	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
)

// submitView is the view-model for the feedback form.
type submitView struct {
	Round       models.FeedbackRound
	Template    *models.Template
	IsSelf      bool
	SubjectName string
	Existing    *models.Submission
	Values      map[string]string // prefilled responses by question key
	Ratings     map[string]models.CompetencyRating
	Error       string
}

// participantRole checks the user's relationship to the round and returns
// (isSelf, isReviewer).
func participantRole(round *models.FeedbackRound, userID string) (isSelf, isReviewer bool) {
	if round.SubjectID == userID {
		isSelf = true
	}
	for _, rev := range round.Reviewers {
		if rev.ReviewerID == userID {
			isReviewer = true
		}
	}
	return isSelf, isReviewer
}

// SubmitForm renders the feedback form for a new submission.
func (h *Handlers) SubmitForm(w http.ResponseWriter, r *http.Request) {
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
	isSelf, isReviewer := participantRole(round, u.ID)
	if !isSelf && !isReviewer {
		forbidden(w)
		return
	}
	if round.Status != models.RoundActive {
		http.Error(w, "This round is not collecting feedback right now", http.StatusBadRequest)
		return
	}
	// Already submitted → send to the edit view.
	if n, _ := h.Repos.Submissions.CountByRoundAndReviewer(ctx, id, u.ID); n > 0 {
		http.Redirect(w, r, safeLocalPath("/rounds/"+id+"/submission"), http.StatusSeeOther) // #nosec G710 -- safeLocalPath guarantees a same-origin path
		return
	}

	tmpl, err := h.resolveTemplate(r, round.TemplateID)
	if err != nil {
		serverError(w, err)
		return
	}
	subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
	h.renderSubmitForm(w, r, submitView{
		Round: *round, Template: tmpl, IsSelf: isSelf, SubjectName: nameOf(subject),
		Values: map[string]string{}, Ratings: map[string]models.CompetencyRating{},
	}, http.StatusOK)
}

// CreateSubmission validates and stores a new submission.
func (h *Handlers) CreateSubmission(w http.ResponseWriter, r *http.Request) {
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
	isSelf, isReviewer := participantRole(round, u.ID)
	if !isSelf && !isReviewer {
		forbidden(w)
		return
	}
	if round.Status != models.RoundActive {
		http.Error(w, "This round is not collecting feedback right now", http.StatusBadRequest)
		return
	}

	tmpl, err := h.resolveTemplate(r, round.TemplateID)
	if err != nil {
		serverError(w, err)
		return
	}

	sub, verr := h.parseSubmission(r, round, tmpl, isSelf)
	if verr != "" {
		subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
		h.renderSubmitForm(w, r, submitView{
			Round: *round, Template: tmpl, IsSelf: isSelf, SubjectName: nameOf(subject),
			Values: sub.Responses, Ratings: ratingsMap(sub.Ratings), Error: verr,
		}, http.StatusUnprocessableEntity)
		return
	}
	sub.ReviewerID = u.ID
	if err := h.Repos.Submissions.Create(ctx, sub); err != nil {
		serverError(w, err)
		return
	}
	redirect(w, r, "/rounds/"+id)
}

// EditSubmissionForm renders the form prefilled with the user's submission.
func (h *Handlers) EditSubmissionForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	id := chi.URLParam(r, "id")

	round, err := h.Repos.Rounds.FindByID(ctx, id)
	if err != nil {
		notFound(w)
		return
	}
	existing, err := h.Repos.Submissions.FindByRoundAndReviewer(ctx, id, u.ID)
	if errors.Is(err, repo.ErrNotFound) {
		http.Redirect(w, r, safeLocalPath("/rounds/"+id+"/submit"), http.StatusSeeOther) // #nosec G710 -- safeLocalPath guarantees a same-origin path
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	tmpl, err := h.resolveTemplate(r, round.TemplateID)
	if err != nil {
		serverError(w, err)
		return
	}
	subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
	h.renderSubmitForm(w, r, submitView{
		Round: *round, Template: tmpl, IsSelf: existing.IsSelf, SubjectName: nameOf(subject),
		Existing: existing, Values: existing.Responses, Ratings: ratingsMap(existing.Ratings),
	}, http.StatusOK)
}

// UpdateSubmission validates and saves edits to an existing submission (only
// while the round is still active).
func (h *Handlers) UpdateSubmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	id := chi.URLParam(r, "id")

	round, err := h.Repos.Rounds.FindByID(ctx, id)
	if err != nil {
		notFound(w)
		return
	}
	existing, err := h.Repos.Submissions.FindByRoundAndReviewer(ctx, id, u.ID)
	if err != nil {
		notFound(w)
		return
	}
	if round.Status != models.RoundActive {
		http.Error(w, "This round is closed to edits", http.StatusBadRequest)
		return
	}
	tmpl, err := h.resolveTemplate(r, round.TemplateID)
	if err != nil {
		serverError(w, err)
		return
	}

	sub, verr := h.parseSubmission(r, round, tmpl, existing.IsSelf)
	if verr != "" {
		subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
		h.renderSubmitForm(w, r, submitView{
			Round: *round, Template: tmpl, IsSelf: existing.IsSelf, SubjectName: nameOf(subject),
			Existing: existing, Values: sub.Responses, Ratings: ratingsMap(sub.Ratings), Error: verr,
		}, http.StatusUnprocessableEntity)
		return
	}
	sub.ID = existing.ID
	sub.ReviewerID = u.ID
	if err := h.Repos.Submissions.Update(ctx, sub); err != nil {
		serverError(w, err)
		return
	}
	redirect(w, r, "/rounds/"+id)
}

// parseSubmission reads the form into a Submission and validates it. Returns a
// non-empty error string on validation failure (the partial submission is still
// returned so the form can be re-rendered with the user's input).
func (h *Handlers) parseSubmission(r *http.Request, round *models.FeedbackRound, tmpl *models.Template, isSelf bool) (*models.Submission, string) {
	_ = r.ParseForm()

	responses := map[string]string{}
	if tmpl != nil {
		for _, q := range tmpl.Questions {
			responses[q.Key] = strings.TrimSpace(r.FormValue("response_" + q.Key))
		}
	} else {
		for _, k := range []string{"a", "b", "c", "d"} {
			responses[k] = strings.TrimSpace(r.FormValue("response_" + k))
		}
	}

	sub := &models.Submission{RoundID: round.ID, Responses: responses, IsSelf: isSelf}

	// Ratings (if the template has competencies).
	if tmpl != nil && len(tmpl.Competencies) > 0 {
		for _, comp := range tmpl.Competencies {
			scoreStr := r.FormValue("score_" + comp.Key)
			if scoreStr == "" {
				continue
			}
			score, _ := strconv.Atoi(scoreStr)
			sub.Ratings = append(sub.Ratings, models.CompetencyRating{
				Key:           comp.Key,
				Score:         score,
				Justification: strings.TrimSpace(r.FormValue("justification_" + comp.Key)),
			})
		}
	}

	if !isSelf {
		sub.Relationship = models.ReviewerRelationship(r.FormValue("relationship"))
		sub.InteractionFrequency = models.InteractionFrequency(r.FormValue("interaction_frequency"))
		sub.PrivateNotes = strings.TrimSpace(r.FormValue("private_notes"))
		if !sub.Relationship.IsValid() {
			return sub, "Please select your relationship to the subject."
		}
		if !sub.InteractionFrequency.IsValid() {
			return sub, "Please select how often you interacted."
		}
	}

	// At least one response must be non-empty.
	answered := false
	for _, v := range responses {
		if v != "" {
			answered = true
			break
		}
	}
	if !answered {
		return sub, "Please answer at least one question."
	}

	if err := ai.ValidateRatings(sub.Ratings, tmpl); err != nil {
		return sub, err.Error()
	}
	return sub, ""
}

func (h *Handlers) renderSubmitForm(w http.ResponseWriter, r *http.Request, v submitView, status int) {
	h.View.Page(w, status, h.page(r, "Give feedback", "rounds", "submit_content", v))
}

func ratingsMap(ratings []models.CompetencyRating) map[string]models.CompetencyRating {
	m := make(map[string]models.CompetencyRating, len(ratings))
	for _, rt := range ratings {
		m[rt.Key] = rt
	}
	return m
}
