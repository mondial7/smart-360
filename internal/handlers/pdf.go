package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/pdf"
)

// DownloadPDF streams the consolidation as a PDF. Access mirrors the on-screen
// view: admin, round creator, or the subject after sharing. The manager-only
// channel is stripped for the subject.
func (h *Handlers) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	roundID := chi.URLParam(r, "roundId")

	round, err := h.Repos.Rounds.FindByID(ctx, roundID)
	if err != nil {
		notFound(w)
		return
	}
	if !canAccessConsolidation(u, *round) {
		forbidden(w)
		return
	}
	cons, err := h.Repos.Consolidations.FindByRoundID(ctx, roundID)
	if err != nil {
		notFound(w)
		return
	}
	if u.ID == round.SubjectID && u.Role != models.RoleAdmin && cons.SharedAt == nil {
		forbidden(w)
		return
	}
	if !canSeeManagerOnlyChannel(u, *round) {
		cons.ManagerOnlyChannel = nil
	}

	subject, _ := h.Repos.Users.FindByID(ctx, round.SubjectID)
	subjectModel := models.User{}
	if subject != nil {
		subjectModel = *subject
	}

	bytes, err := pdf.Render(subjectModel, *round, *cons)
	if err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename=%q`, pdf.Filename(subjectModel.Name, round.CreatedAt.Format("2006-01-02"))))
	_, _ = w.Write(bytes)
}
