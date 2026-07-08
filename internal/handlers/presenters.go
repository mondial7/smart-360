package handlers

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

// RoundCard is the view-model for a round shown in a list or on the dashboard.
type RoundCard struct {
	Round            models.FeedbackRound
	SubjectName      string
	CreatorName      string
	ReviewerCount    int
	SubmittedByMe    bool
	IsSubject        bool
	HasConsolidation bool
}

// roundsForMe returns the rounds a user must engage with: rounds where they are
// a reviewer, plus their own active rounds (for self-assessment). Deduplicated.
func (h *Handlers) roundsForMe(ctx context.Context, userID string) ([]models.FeedbackRound, error) {
	asReviewer, err := h.Repos.Rounds.FindByReviewerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	asSubject, err := h.Repos.Rounds.FindBySubjectID(ctx, userID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(asReviewer))
	out := make([]models.FeedbackRound, 0, len(asReviewer)+len(asSubject))
	for _, r := range asReviewer {
		if !seen[r.ID] {
			seen[r.ID] = true
			out = append(out, r)
		}
	}
	for _, r := range asSubject {
		// The subject participates only while the round is collecting feedback.
		if r.Status == models.RoundActive && !seen[r.ID] {
			seen[r.ID] = true
			out = append(out, r)
		}
	}
	return out, nil
}

// toCard builds a RoundCard, resolving names and the current user's submission
// state. users is an index built by userMap.
func (h *Handlers) toCard(ctx context.Context, round models.FeedbackRound, currentUserID string, users map[string]models.User) RoundCard {
	card := RoundCard{
		Round:         round,
		SubjectName:   users[round.SubjectID].Name,
		CreatorName:   users[round.CreatedByID].Name,
		ReviewerCount: len(round.Reviewers),
		IsSubject:     round.SubjectID == currentUserID,
	}
	if n, err := h.Repos.Submissions.CountByRoundAndReviewer(ctx, round.ID, currentUserID); err == nil {
		card.SubmittedByMe = n > 0
	}
	if _, err := h.Repos.Consolidations.FindByRoundID(ctx, round.ID); err == nil {
		card.HasConsolidation = true
	}
	return card
}

// allUsersIndex loads every user into an ID→user map for name resolution.
func (h *Handlers) allUsersIndex(ctx context.Context) (map[string]models.User, error) {
	users, err := h.Repos.Users.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return userMap(users), nil
}

// canSeeManagerOnlyChannel reports whether the caller may see the private
// manager-only synthesis: global admins and the round creator, never the subject.
func canSeeManagerOnlyChannel(user *models.User, round models.FeedbackRound) bool {
	if user.Role == models.RoleAdmin {
		return true
	}
	return user.ID == round.CreatedByID
}

// canAccessConsolidation mirrors the legacy rule: admin, round creator, or the
// subject (subject only after it has been shared, enforced separately).
func canAccessConsolidation(user *models.User, round models.FeedbackRound) bool {
	return user.Role == models.RoleAdmin || user.ID == round.CreatedByID || user.ID == round.SubjectID
}
