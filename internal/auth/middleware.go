package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/models"
)

// RequireAuth loads the session user into the request context, or sends the
// visitor to /login. htmx requests get an HX-Redirect header instead of a body
// redirect so the client navigates rather than swapping the login page in.
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := s.resolveUser(r.Context(), r)
		if !ok {
			s.redirectToLogin(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(contextWithUser(r.Context(), user)))
	})
}

// RequireAdmin gates a route on the global admin role. Assumes RequireAuth ran.
func (s *Service) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFrom(r.Context())
		if !ok {
			s.redirectToLogin(w, r)
			return
		}
		if user.Role != models.RoleAdmin {
			forbidden(w, "Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireTeamAdminOrAdmin gates a route on the admin OR team_admin role, without
// per-team scoping (handlers that key on a team must check ownership too).
func (s *Service) RequireTeamAdminOrAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFrom(r.Context())
		if !ok {
			s.redirectToLogin(w, r)
			return
		}
		if user.Role != models.RoleAdmin && user.Role != models.RoleTeamAdmin {
			forbidden(w, "Admin or team admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireTeamScope gates a :id team route: global admins pass; a team admin
// passes only for their own team. Use after RequireTeamAdminOrAdmin.
func (s *Service) RequireTeamScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFrom(r.Context())
		if !ok {
			s.redirectToLogin(w, r)
			return
		}
		if user.Role == models.RoleAdmin {
			next.ServeHTTP(w, r)
			return
		}
		teamID := chi.URLParam(r, "id")
		if user.Role == models.RoleTeamAdmin && user.TeamID != nil && *user.TeamID == teamID {
			next.ServeHTTP(w, r)
			return
		}
		forbidden(w, "Not authorized to manage this team")
	})
}

func (s *Service) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	if wantsRedirect(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusUnauthorized)
}

func forbidden(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusForbidden)
}
