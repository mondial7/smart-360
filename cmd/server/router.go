package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/mondial7/smart-360/internal/auth"
	"github.com/mondial7/smart-360/internal/config"
	"github.com/mondial7/smart-360/internal/handlers"
	"github.com/mondial7/smart-360/web"
)

// newRouter wires middleware, static assets, public auth routes, and the
// authenticated application routes.
func newRouter(cfg *config.Config, authSvc *auth.Service, h *handlers.Handlers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Static assets (embedded).
	r.Handle("/static/*", http.FileServerFS(web.StaticFS))
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Public auth routes.
	r.Get("/login", h.LoginPage)
	r.Get("/auth/google", authSvc.StartGoogleLogin)
	r.Get("/auth/callback", authSvc.GoogleCallback)
	r.Get("/logout", authSvc.Logout)
	if cfg.DevMode {
		r.Get("/auth/dev-login", authSvc.DevLogin)
	}

	// Authenticated application routes.
	r.Group(func(r chi.Router) {
		r.Use(authSvc.RequireAuth)
		r.Use(authSvc.ProtectCSRF)

		r.Get("/", h.Dashboard)
		h.MountAppRoutes(r)
	})

	return r
}
