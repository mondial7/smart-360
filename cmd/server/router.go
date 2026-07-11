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
	r.Use(securityHeaders)

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

// contentSecurityPolicy is intentionally strict: all scripts are self-hosted
// (htmx, sse, app.js), so no 'unsafe-inline' for scripts. Inline style="…"
// attributes in templates require 'unsafe-inline' for styles only. SSE is
// same-origin (connect-src 'self').
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: https:; " +
	"connect-src 'self'; " +
	"font-src 'self'; " +
	"form-action 'self'; " +
	"base-uri 'self'; " +
	"frame-ancestors 'none'"

// securityHeaders sets defensive response headers on every response.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		next.ServeHTTP(w, r)
	})
}
