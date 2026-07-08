package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

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

	// Per-IP rate limits (in-memory). These are a backstop against brute force
	// and abuse; a production deployment behind a proxy should still throttle
	// there too. Static assets and the health check are intentionally exempt.
	authLimit := httprate.LimitByIP(20, time.Minute)    // login/callback/dev-login
	submitLimit := httprate.LimitByIP(40, time.Minute)  // feedback submission
	appBackstop := httprate.LimitByIP(300, time.Minute) // everything authenticated

	// Static assets (embedded).
	r.Handle("/static/*", http.FileServerFS(web.StaticFS))
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Public routes.
	r.Get("/login", h.LoginPage)
	r.Get("/logout", authSvc.Logout)
	// Auth endpoints are the highest-value brute-force target → tighter limit.
	r.Group(func(r chi.Router) {
		r.Use(authLimit)
		r.Get("/auth/google", authSvc.StartGoogleLogin)
		r.Get("/auth/callback", authSvc.GoogleCallback)
		if cfg.DevMode {
			r.Get("/auth/dev-login", authSvc.DevLogin)
		}
	})

	// Authenticated application routes.
	r.Group(func(r chi.Router) {
		r.Use(appBackstop)
		r.Use(authSvc.RequireAuth)
		r.Use(authSvc.ProtectCSRF)

		r.Get("/", h.Dashboard)
		h.MountAppRoutes(r, submitLimit)
	})

	return r
}
