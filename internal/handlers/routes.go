package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// MountAppRoutes registers every authenticated application route. It runs inside
// the RequireAuth + ProtectCSRF group in the router. submitLimit is a per-IP
// rate-limit middleware applied to the feedback-submission endpoints.
func (h *Handlers) MountAppRoutes(r chi.Router, submitLimit func(http.Handler) http.Handler) {
	// Onboarding
	r.Post("/onboarding/complete", h.CompleteOnboarding)

	// Rounds
	r.Get("/rounds", h.RoundsList)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Get("/rounds/new", h.NewRoundForm)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds", h.CreateRound)
	r.Get("/rounds/{id}", h.RoundDetails)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds/{id}/start", h.StartRound)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds/{id}/close", h.CloseRound)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds/{id}/edit", h.UpdateRoundMeta)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds/{id}/reviewers", h.AddRoundReviewer)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/rounds/{id}/reviewers/{reviewerId}/remove", h.RemoveRoundReviewer)

	// Submissions (writes are rate-limited per IP)
	r.Get("/rounds/{id}/submit", h.SubmitForm)
	r.With(submitLimit).Post("/rounds/{id}/submit", h.CreateSubmission)
	r.Get("/rounds/{id}/submission", h.EditSubmissionForm)
	r.With(submitLimit).Post("/rounds/{id}/submission", h.UpdateSubmission)

	// Consolidation + SSE
	r.With(h.Auth.RequireAdmin).Post("/rounds/{id}/consolidate", h.StartConsolidation)
	r.With(h.Auth.RequireAdmin).Get("/rounds/{id}/consolidate/stream", h.ConsolidationStream)
	r.Get("/rounds/{id}/consolidation", h.ConsolidationView)
	r.With(h.Auth.RequireTeamAdminOrAdmin).Post("/consolidations/{id}", h.UpdateConsolidation)
	r.With(h.Auth.RequireAdmin).Post("/consolidations/{id}/share", h.ShareConsolidation)
	r.Get("/consolidations/{roundId}/pdf", h.DownloadPDF)

	// My feedback
	r.Get("/my-feedback", h.MyFeedback)
	r.Get("/my-feedback/compare", h.CompareRounds)

	// Team directory
	r.Get("/team", h.TeamDirectory)

	// Team management (admin / team admin)
	r.With(h.Auth.RequireAdmin).Get("/teams", h.TeamsList)
	r.With(h.Auth.RequireAdmin).Get("/teams/new", h.NewTeamForm)
	r.With(h.Auth.RequireAdmin).Post("/teams", h.CreateTeam)
	r.With(h.Auth.RequireTeamAdminOrAdmin, h.Auth.RequireTeamScope).Get("/teams/{id}", h.TeamDetails)
	r.With(h.Auth.RequireTeamAdminOrAdmin, h.Auth.RequireTeamScope).Post("/teams/{id}/members", h.AddTeamMember)
	r.With(h.Auth.RequireTeamAdminOrAdmin, h.Auth.RequireTeamScope).Post("/teams/{id}/members/{userId}/remove", h.RemoveTeamMember)
	r.With(h.Auth.RequireTeamAdminOrAdmin, h.Auth.RequireTeamScope).Get("/teams/{id}/create-round", h.NewTeamRoundForm)
	r.With(h.Auth.RequireTeamAdminOrAdmin, h.Auth.RequireTeamScope).Post("/teams/{id}/rounds", h.CreateTeamRounds)

	// Users & roles (admin)
	r.With(h.Auth.RequireAdmin).Get("/users", h.UsersList)
	r.With(h.Auth.RequireAdmin).Post("/users/{id}/role", h.UpdateUserRole)

	// Analytics + audit (admin)
	r.With(h.Auth.RequireAdmin).Get("/analytics", h.Analytics)
	r.With(h.Auth.RequireAdmin).Get("/audit-logs", h.AuditLogs)

	// Live application logs (admin)
	r.With(h.Auth.RequireAdmin).Get("/admin/logs", h.LogsPage)
	r.With(h.Auth.RequireAdmin).Get("/admin/logs/stream", h.LogsStream)
}
