// Package handlers holds the HTTP handlers for the server-rendered app. Each
// handler renders HTML (full pages via the base layout, or fragments for htmx
// swaps). Business logic lives in the repo, ai, and pdf packages.
package handlers

import (
	"errors"
	"net/http"

	"github.com/mondial7/smart-360/internal/auth"
	"github.com/mondial7/smart-360/internal/config"
	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
	"github.com/mondial7/smart-360/internal/view"
)

// Handlers bundles the dependencies every handler needs.
type Handlers struct {
	Repos repo.Repositories
	Auth  *auth.Service
	View  *view.Renderer
	Cfg   *config.Config
}

// New constructs a Handlers.
func New(repos repo.Repositories, authSvc *auth.Service, renderer *view.Renderer, cfg *config.Config) *Handlers {
	return &Handlers{Repos: repos, Auth: authSvc, View: renderer, Cfg: cfg}
}

// page builds the base-layout envelope, pulling the current user and CSRF token
// from the request.
func (h *Handlers) page(r *http.Request, title, active, content string, data any) view.PageData {
	u, _ := auth.UserFrom(r.Context())
	return view.PageData{
		Title:   title,
		Active:  active,
		User:    u,
		CSRF:    h.Auth.CSRFToken(r),
		Content: content,
		Data:    data,
	}
}

// user is a convenience for the authenticated user (handlers behind RequireAuth).
func (h *Handlers) user(r *http.Request) *models.User {
	u, _ := auth.UserFrom(r.Context())
	return u
}

// resolveTemplate loads the round's template, falling back to the default slug
// when the round has no explicit template.
func (h *Handlers) resolveTemplate(r *http.Request, templateID *string) (*models.Template, error) {
	if templateID != nil && *templateID != "" {
		return h.Repos.Templates.FindByID(r.Context(), *templateID)
	}
	t, err := h.Repos.Templates.FindBySlug(r.Context(), models.DefaultTemplateSlug)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, nil
	}
	return t, err
}

// serverError renders a 500 with a generic message; details are the caller's to log.
func serverError(w http.ResponseWriter, err error) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
	_ = err
}

// notFound renders a 404.
func notFound(w http.ResponseWriter) {
	http.Error(w, "Not found", http.StatusNotFound)
}

// forbidden renders a 403.
func forbidden(w http.ResponseWriter) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// userMap indexes users by ID for template lookups (avoids N joins in views).
func userMap(users []models.User) map[string]models.User {
	m := make(map[string]models.User, len(users))
	for _, u := range users {
		m[u.ID] = u
	}
	return m
}
