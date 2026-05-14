package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"smart360/database"
	"smart360/handlers"
	"smart360/middleware"
	"smart360/models"
	"smart360/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

// Build metadata, populated by `go build -ldflags "-X main.version=..."`.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

// getEnv retrieves an environment variable or returns a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		log.SetFlags(0)
		log.Printf("smart360 %s (commit %s, built %s)", version, commit, buildDate)
		return
	}

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// JWT_SECRET is required: without it, attackers could forge tokens with
	// the previously hardcoded fallback constant.
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Initialize database
	database.InitDB()

	// Round templates are seeded on every boot so the bundled defaults stay
	// in sync with the current code. Cheap (one upsert per template) and
	// safe to run in production.
	database.SeedDefaultTemplates()

	// Seed development data (only in development mode)
	if os.Getenv("DEV_MODE") == "true" {
		log.Println("WARNING: Development mode enabled - seeding test data")
		database.SeedDevData()
	}

	// Initialize OAuth
	handlers.InitOAuthConfig()

	// Setup Gin
	r := gin.Default()

	// CORS is only needed when the SPA is served from a different origin
	// (e.g. `npm run dev` on :5173, or the docker-compose nginx frontend).
	// In single-binary mode the SPA is served from this same process, so we
	// skip the middleware to avoid leaking permissive headers.
	if !web.HasIndex() {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{getEnv("FRONTEND_URL", "http://localhost:5173")},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	}

	// Public routes
	r.GET("/api/auth/google", handlers.GetGoogleAuthURL)
	r.GET("/api/auth/callback", handlers.GoogleCallback)

	// Dev-login endpoint (only in development mode)
	if os.Getenv("DEV_MODE") == "true" {
		r.GET("/api/auth/dev-login", handlers.DevLogin)
		log.Println("WARNING: Dev-login endpoint enabled - NEVER enable this in production!")
	}

	// Protected routes
	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.GET("/me", handlers.GetCurrentUser)
		authorized.GET("/users", handlers.GetUsers)
		authorized.GET("/users/with-feedback-stats", middleware.AdminOnly(), handlers.GetUsersWithFeedbackStats)
		authorized.GET("/users/:id", handlers.GetUserByID)
		authorized.PUT("/users/:id/role", middleware.AdminOnly(), handlers.UpdateUserRole)

		// Teams
		authorized.GET("/teams", middleware.AdminOnly(), handlers.GetTeams)
		authorized.GET("/teams/:id", middleware.TeamAdminOrGlobalAdmin(), handlers.GetTeam)
		authorized.GET("/my-team", handlers.GetMyTeam)
		authorized.POST("/teams", middleware.AdminOnly(), handlers.CreateTeam)
		authorized.PUT("/teams/:id", middleware.TeamAdminOrGlobalAdmin(), handlers.UpdateTeam)
		authorized.DELETE("/teams/:id", middleware.AdminOnly(), handlers.DeleteTeam)
		authorized.POST("/teams/:id/members", middleware.TeamAdminOrGlobalAdmin(), handlers.AddTeamMembers)
		authorized.DELETE("/teams/:id/members/:userId", middleware.TeamAdminOrGlobalAdmin(), handlers.RemoveTeamMember)
		authorized.POST("/teams/:id/rounds/create-batch", middleware.TeamAdminOrGlobalAdmin(), handlers.CreateTeamRounds)

		// Round templates - readable by any authenticated user (question wording
		// is public to reviewers anyway). Mutation handlers are not exposed yet;
		// templates ship via the seeded default(s).
		authorized.GET("/templates", handlers.ListTemplates)
		authorized.GET("/templates/:idOrSlug", handlers.GetTemplate)

		// Rounds - admin and team admin can create / manage
		authorized.POST("/rounds", middleware.AdminOrTeamAdminRole(), handlers.CreateFeedbackRound)
		authorized.GET("/rounds", middleware.AdminOnly(), handlers.GetAllRounds)
		authorized.GET("/rounds/:id", handlers.GetRoundDetails)
		authorized.PUT("/rounds/:id", middleware.AdminOrTeamAdminRole(), handlers.UpdateFeedbackRound)
		authorized.POST("/rounds/:id/reviewers", middleware.AdminOrTeamAdminRole(), handlers.AddReviewersToRound)
		authorized.DELETE("/rounds/:id/reviewers/:reviewerId", middleware.AdminOrTeamAdminRole(), handlers.RemoveReviewerFromRound)
		authorized.POST("/rounds/:id/start", middleware.AdminOrTeamAdminRole(), handlers.StartFeedbackRound)
		authorized.POST("/rounds/:id/close", middleware.AdminOrTeamAdminRole(), handlers.CloseFeedbackRound)
		authorized.GET("/rounds-for-me", handlers.GetRoundsForMe)
		authorized.GET("/my-rounds", handlers.GetMyRounds)

		// Submissions
		authorized.POST("/submissions", handlers.SubmitFeedback)
		authorized.GET("/submissions/check/:roundId", handlers.CheckSubmissionStatus)
		authorized.GET("/submissions/round/:roundId", handlers.GetRoundSubmissions)
		authorized.PUT("/submissions/:id", handlers.UpdateSubmission)
		authorized.GET("/submissions/:submissionId", handlers.GetSubmissionDetails)

		// Consolidations (admin only for generation)
		authorized.POST("/rounds/:id/consolidate", middleware.AdminOnly(), handlers.ConsolidateFeedback)
		authorized.GET("/rounds/:id/moderation-logs", handlers.GetModerationLogsForRound)
		authorized.GET("/consolidations/:roundId", handlers.GetConsolidation)
		authorized.GET("/consolidations/:roundId/pdf", handlers.DownloadConsolidationPDF)
		authorized.PUT("/consolidations/:id", middleware.AdminOnly(), handlers.UpdateConsolidation)
		authorized.PUT("/consolidations/:id/notes", middleware.AdminOnly(), handlers.UpdateConsolidationNotes)
		authorized.POST("/consolidations/:id/share", middleware.AdminOnly(), handlers.ShareConsolidation)

		// Analytics
		authorized.GET("/analytics/me", handlers.GetMyAnalytics)
		authorized.GET("/analytics/admin", middleware.AdminOnly(), handlers.GetAdminAnalytics)

		// Audit logs (admin only)
		authorized.GET("/audit-logs", middleware.AdminOnly(), handlers.GetAuditLogs)
		authorized.GET("/audit-logs/round/:id", middleware.AdminOnly(), handlers.GetRoundAuditLogs)

		// Dashboard
		authorized.GET("/dashboard/stats", handlers.GetDashboardStats)
		authorized.GET("/dashboard/my-rounds", handlers.GetMyRounds)
		authorized.GET("/dashboard/rounds-for-me", handlers.GetRoundsForMe)
		authorized.GET("/dashboard/my-submissions", handlers.GetMySubmissions)
		authorized.GET("/dashboard/my-consolidations", handlers.GetMyConsolidations)

		// Admin debug endpoints
		authorized.GET("/debug/submissions", middleware.AdminOnly(), handlers.DebugSubmissions)
		authorized.GET("/debug/reviewers", middleware.AdminOnly(), func(c *gin.Context) {
			db := database.GetDB()
			ctx := context.Background()

			// Get all reviewers
			cursor, err := db.Collection("round_reviewers").Find(ctx, bson.M{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviewers"})
				return
			}
			defer cursor.Close(ctx)

			var reviewers []models.RoundReviewer
			if err = cursor.All(ctx, &reviewers); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode reviewers"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"total_reviewers": len(reviewers),
				"reviewers":       reviewers,
			})
		})
	}

	// Mount embedded SPA if it was bundled at build time (single-binary
	// distribution). API routes above take precedence; everything else
	// falls back to index.html so client-side routing works.
	if web.HasIndex() {
		assets, err := web.Assets()
		if err != nil {
			log.Fatalf("Failed to load embedded SPA assets: %v", err)
		}
		mountSPA(r, assets)
		log.Println("Serving embedded SPA from /")
	}

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on :%s", port)
	r.Run(":" + port)
}

// mountSPA serves the embedded frontend assets and falls back to index.html
// for any non-API path that does not match a file (so Vue Router controls
// client-side navigation).
func mountSPA(r *gin.Engine, assets fs.FS) {
	fileServer := http.FileServer(http.FS(assets))

	r.NoRoute(func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")

		// API requests that fall through to NoRoute are genuine 404s.
		if strings.HasPrefix(path, "api/") || path == "api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Serve the file if it exists, otherwise serve index.html (SPA fallback).
		if path != "" {
			if f, err := assets.Open(path); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
