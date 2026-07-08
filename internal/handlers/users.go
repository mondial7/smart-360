package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/models"
)

// UsersList shows all users with their role and a control to change it (admin).
func (h *Handlers) UsersList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	users, err := h.Repos.Users.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	data := map[string]any{"Users": users, "MeID": u.ID}
	h.View.Page(w, http.StatusOK, h.page(r, "Users", "users", "users_content", data))
}

// UpdateUserRole changes a user's role (admin). It refuses to demote the last
// remaining admin, which would leave the install with no one who can manage it.
func (h *Handlers) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	actor := h.user(r)
	id := chi.URLParam(r, "id")
	_ = r.ParseForm()

	role := models.UserRole(r.FormValue("role"))
	if !role.IsValid() {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	target, err := h.Repos.Users.FindByID(ctx, id)
	if err != nil {
		notFound(w)
		return
	}
	if target.Role == role {
		redirect(w, r, "/users")
		return
	}

	// Guard: don't strip admin from the last admin.
	if target.Role == models.RoleAdmin && role != models.RoleAdmin {
		users, err := h.Repos.Users.FindAll(ctx)
		if err != nil {
			serverError(w, err)
			return
		}
		admins := 0
		for _, usr := range users {
			if usr.Role == models.RoleAdmin {
				admins++
			}
		}
		if admins <= 1 {
			http.Error(w, "Cannot demote the last admin", http.StatusBadRequest)
			return
		}
	}

	if err := h.Repos.Users.UpdateRole(ctx, id, role); err != nil {
		serverError(w, err)
		return
	}
	h.audit(ctx, auditParams{
		Action: models.AuditUserRoleChanged, Actor: actor,
		Description: "Changed role for " + target.Name,
		OldValue:    string(target.Role), NewValue: string(role),
	})
	redirect(w, r, "/users")
}
