package handlers

import (
	"net/http"

	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/view"
)

// Analytics renders the admin analytics page: headline counters and a donut of
// rounds by status.
func (h *Handlers) Analytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.Repos.Users.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	teams, err := h.Repos.Teams.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	rounds, err := h.Repos.Rounds.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}

	byStatus := map[models.RoundStatus]int{}
	for _, rd := range rounds {
		byStatus[rd.Status]++
	}
	slices := []view.DonutSlice{
		{Label: "Draft", Value: byStatus[models.RoundDraft], Color: "#94a3b8"},
		{Label: "Active", Value: byStatus[models.RoundActive], Color: "#2563eb"},
		{Label: "Closed", Value: byStatus[models.RoundClosed], Color: "#d97706"},
		{Label: "Shared", Value: byStatus[models.RoundShared], Color: "#16a34a"},
	}
	completed := byStatus[models.RoundClosed] + byStatus[models.RoundShared]

	data := map[string]any{
		"TotalUsers":   len(users),
		"TotalTeams":   len(teams),
		"TotalRounds":  len(rounds),
		"Completed":    completed,
		"CompletedPct": pctInt(completed, len(rounds)),
		"StatusSlices": slices,
	}
	h.View.Page(w, http.StatusOK, h.page(r, "Analytics", "analytics", "analytics_content", data))
}

func pctInt(part, total int) int {
	if total == 0 {
		return 0
	}
	return part * 100 / total
}
