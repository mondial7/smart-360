package handlers

import "github.com/go-chi/chi/v5"

// MountAppRoutes registers every authenticated application route. It runs inside
// the RequireAuth + ProtectCSRF group in the router.
func (h *Handlers) MountAppRoutes(r chi.Router) {
	// Rounds
	r.Get("/rounds", h.RoundsList)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Get("/rounds/new", h.NewRoundForm)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds", h.CreateRound)
	r.Get("/rounds/{id}", h.RoundDetails)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds/{id}/start", h.StartRound)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds/{id}/close", h.CloseRound)

	// Submissions
	r.Get("/rounds/{id}/submit", h.SubmitForm)
	r.Post("/rounds/{id}/submit", h.CreateSubmission)
	r.Get("/rounds/{id}/submission", h.EditSubmissionForm)
	r.Post("/rounds/{id}/submission", h.UpdateSubmission)

	// Consolidation + SSE
	r.With(h.Auth.RequireAdmin).Post("/rounds/{id}/consolidate", h.StartConsolidation)
	r.With(h.Auth.RequireAdmin).Get("/rounds/{id}/consolidate/stream", h.ConsolidationStream)
	r.Get("/rounds/{id}/consolidation", h.ConsolidationView)
	r.With(h.Auth.RequireAdmin).Post("/consolidations/{id}/share", h.ShareConsolidation)
	r.Get("/consolidations/{roundId}/pdf", h.DownloadPDF)

	// My feedback
	r.Get("/my-feedback", h.MyFeedback)

	// Team directory
	r.Get("/team", h.TeamDirectory)

	// Team management (admin / team admin)
	r.With(h.Auth.RequireAdmin).Get("/teams", h.TeamsList)
	r.With(h.Auth.RequireAdmin).Get("/teams/new", h.NewTeamForm)
	r.With(h.Auth.RequireAdmin).Post("/teams", h.CreateTeam)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Get("/teams/{id}", h.TeamDetails)
	r.With(h.Auth.RequireTeamAdminOrAdmin, h.Auth.RequireTeamScope).Post("/teams/{id}/members", h.AddTeamMember)
	r.With(h.Auth.RequireTeamAdminOrAdmin, h.Auth.RequireTeamScope).Post("/teams/{id}/members/{userId}/remove", h.RemoveTeamMember)

	// Analytics + audit (admin)
	r.With(h.Auth.RequireAdmin).Get("/analytics", h.Analytics)
	r.With(h.Auth.RequireAdmin).Get("/audit-logs", h.AuditLogs)
}
