package handlers

import (
	"net/http"

	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/view"
)

// MyFeedback lists the shared consolidations for rounds where the user is the
// subject, and shows a radar of the most recent one.
func (h *Handlers) MyFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)

	cons, err := h.Repos.Consolidations.FindSharedBySubjectID(ctx, u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	// The subject never sees the manager-only channel.
	for i := range cons {
		cons[i].ManagerOnlyChannel = nil
	}

	var radar []view.RadarAxis
	if len(cons) > 0 {
		latest := cons[0]
		radar = []view.RadarAxis{
			{Label: "Strengths", Value: float64(len(latest.Strengths))},
			{Label: "Growth areas", Value: float64(len(latest.AreasForImprovement))},
			{Label: "Focus areas", Value: float64(len(latest.ActionableInsights))},
			{Label: "Aligned", Value: float64(deltaLen(latest.SelfVsOthersDelta))},
		}
	}

	data := map[string]any{"Consolidations": cons, "Radar": radar}
	h.View.Page(w, http.StatusOK, h.page(r, "My feedback", "my-feedback", "my_feedback_content", data))
}

func deltaLen(d *models.SelfVsOthersDelta) int {
	if d == nil {
		return 0
	}
	return len(d.Aligned)
}
