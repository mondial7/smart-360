package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"time"

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

	data := map[string]any{
		"Consolidations": cons,
		"Radar":          radar,
		"CanCompare":     len(cons) >= 2,
	}
	pd := h.page(r, "My feedback", "my-feedback", "my_feedback_content", data)
	switch {
	case r.URL.Query().Get("requested") == "1":
		pd.Flash = "Your feedback request was sent — a manager will review and start it."
	case r.URL.Query().Get("pending") == "1":
		pd.Flash = "You already have a feedback request awaiting a manager's approval."
	}
	h.View.Page(w, http.StatusOK, pd)
}

func deltaLen(d *models.SelfVsOthersDelta) int {
	if d == nil {
		return 0
	}
	return len(d.Aligned)
}

// chartPalette cycles line/series colours for the comparison charts.
var chartPalette = []string{"#4f46e5", "#16a34a", "#d97706", "#dc2626", "#2563eb", "#7c3aed", "#0891b2", "#db2777"}

// CompareRounds shows the subject's competency scores trending across their
// shared rounds — the growth-over-time view. Read-only over shared
// consolidations the subject already has access to; no manager-only data.
func (h *Handlers) CompareRounds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)

	cons, err := h.Repos.Consolidations.FindSharedBySubjectID(ctx, u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	// Oldest → newest so the trend reads left to right.
	sort.Slice(cons, func(i, j int) bool {
		return sharedTime(cons[i]).Before(sharedTime(cons[j]))
	})

	if len(cons) < 2 {
		h.View.Page(w, http.StatusOK, h.page(r, "Compare rounds", "my-feedback",
			"compare_content", map[string]any{"Enough": false}))
		return
	}

	xLabels := make([]string, len(cons))
	for i := range cons {
		xLabels[i] = sharedTime(cons[i]).Format("Jan 2006")
	}

	// Union of competencies in first-seen order; per-round "others average".
	type comp struct {
		name   string
		scores []*float64 // one per round
	}
	order := []string{}
	byKey := map[string]*comp{}
	for roundIdx := range cons {
		for _, cr := range cons[roundIdx].CompetencyRatings {
			c, ok := byKey[cr.Key]
			if !ok {
				c = &comp{name: cr.Name, scores: make([]*float64, len(cons))}
				byKey[cr.Key] = c
				order = append(order, cr.Key)
			}
			if cr.OthersAverage != nil {
				v := *cr.OthersAverage
				c.scores[roundIdx] = &v
			}
		}
	}

	var series []view.LineSeries
	type tableRow struct {
		Name   string
		Scores []string
	}
	var rows []tableRow
	for i, key := range order {
		c := byKey[key]
		series = append(series, view.LineSeries{
			Label:  c.name,
			Color:  chartPalette[i%len(chartPalette)],
			Points: c.scores,
		})
		cells := make([]string, len(c.scores))
		for j, s := range c.scores {
			if s == nil {
				cells[j] = "—"
			} else {
				cells[j] = fmt.Sprintf("%.1f", *s)
			}
		}
		rows = append(rows, tableRow{Name: c.name, Scores: cells})
	}

	type timelineEntry struct {
		Date    string
		Summary string
	}
	var timeline []timelineEntry
	for i := len(cons) - 1; i >= 0; i-- { // newest first for the timeline
		timeline = append(timeline, timelineEntry{
			Date:    sharedTime(cons[i]).Format("Jan 2, 2006"),
			Summary: cons[i].ExecutiveSummary,
		})
	}

	data := map[string]any{
		"Enough":   true,
		"XLabels":  xLabels,
		"Series":   series,
		"Rows":     rows,
		"Timeline": timeline,
		"Rounds":   len(cons),
	}
	h.View.Page(w, http.StatusOK, h.page(r, "Compare rounds", "my-feedback", "compare_content", data))
}

// sharedTime returns the consolidation's shared timestamp (falls back to created).
func sharedTime(c models.Consolidation) time.Time {
	if c.SharedAt != nil {
		return *c.SharedAt
	}
	return c.CreatedAt
}
