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
	asOwner, err := h.Repos.Rounds.FindByCreatedByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	asSubject, err := h.Repos.Rounds.FindBySubjectID(ctx, userID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(asReviewer)+len(asOwner))
	out := make([]models.FeedbackRound, 0, len(asReviewer)+len(asOwner)+len(asSubject))
	add := func(rounds []models.FeedbackRound) {
		for _, r := range rounds {
			if !seen[r.ID] {
				seen[r.ID] = true
				out = append(out, r)
			}
		}
	}
	// Rounds I review, plus every round I own (a team admin manages their own
	// rounds, including ones members self-nominated to them).
	add(asReviewer)
	add(asOwner)
	// As the subject, I only participate while the round is collecting feedback
	// (self-assessment); I never gain owner access to my own rounds.
	for _, r := range asSubject {
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

// usersForRounds resolves just the users a set of rounds reference (subject +
// creator) in a single batch query, rather than loading the whole user table.
func (h *Handlers) usersForRounds(ctx context.Context, rounds []models.FeedbackRound) (map[string]models.User, error) {
	seen := make(map[string]struct{}, len(rounds)*2)
	ids := make([]string, 0, len(rounds)*2)
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	for _, rd := range rounds {
		add(rd.SubjectID)
		add(rd.CreatedByID)
	}
	users, err := h.Repos.Users.FindByIDs(ctx, ids)
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
