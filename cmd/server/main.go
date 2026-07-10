// Command server is the Smart 360 web application: a server-rendered Go app
// (html/template + htmx + SSE) backed by Postgres.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mondial7/smart-360/internal/auth"
	"github.com/mondial7/smart-360/internal/config"
	"github.com/mondial7/smart-360/internal/db"
	"github.com/mondial7/smart-360/internal/handlers"
	"github.com/mondial7/smart-360/internal/repo"
	"github.com/mondial7/smart-360/internal/view"
	"github.com/mondial7/smart-360/web"
)

// Build metadata, injected via -ldflags at release time (see the Makefile).
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-version", "version":
			fmt.Printf("smart360 %s (commit %s, built %s)\n", version, commit, buildDate)
			return
		}
	}
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		return err
	}
	log.Println("migrations applied")

	if err := db.Seed(ctx, pool, cfg.DevMode); err != nil {
		return err
	}

	repos := repo.NewPostgres(pool)
	renderer, err := view.NewRenderer(web.TemplatesFS)
	if err != nil {
		return err
	}
	authSvc := auth.New(cfg, repos)
	h := handlers.New(repos, authSvc, renderer, cfg)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           newRouter(cfg, authSvc, h),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s (dev_mode=%v)", cfg.Port, cfg.DevMode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
