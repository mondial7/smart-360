package handlers

import "net/http"

// LoginPage renders the login screen. Already-authenticated visitors are sent
// to the dashboard.
func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	if u := h.user(r); u != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := map[string]any{"DevMode": h.Cfg.DevMode}
	h.View.Page(w, http.StatusOK, h.page(r, "Sign in", "", "login_content", data))
}

// Dashboard renders the role-aware landing page.
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)

	users, err := h.allUsersIndex(ctx)
	if err != nil {
		serverError(w, err)
		return
	}

	mine, err := h.roundsForMe(ctx, u.ID)
	if err != nil {
		serverError(w, err)
		return
	}

	pending := make([]RoundCard, 0)
	for _, rd := range mine {
		card := h.toCard(ctx, rd, u.ID, users)
		if !card.SubmittedByMe {
			pending = append(pending, card)
		}
	}

	data := map[string]any{
		"PendingReviews": pending,
		"PendingCount":   len(pending),
	}

	if u.Role == "admin" {
		allRounds, err := h.Repos.Rounds.FindAll(ctx)
		if err != nil {
			serverError(w, err)
			return
		}
		active := 0
		for _, rd := range allRounds {
			if rd.Status == "active" {
				active++
			}
		}
		data["IsAdmin"] = true
		data["TotalUsers"] = len(users)
		data["TotalRounds"] = len(allRounds)
		data["ActiveRounds"] = active
	}

	h.View.Page(w, http.StatusOK, h.page(r, "Dashboard", "dashboard", "dashboard_content", data))
}
