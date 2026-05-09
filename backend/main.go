package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"smart360/database"
	"smart360/handlers"
	"smart360/middleware"
	"smart360/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

// getEnv retrieves an environment variable or returns a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	database.InitDB()

	// Seed development data (only in development mode)
	if os.Getenv("DEV_MODE") == "true" {
		log.Println("WARNING: Development mode enabled - seeding test data")
		database.SeedDevData()
	}

	// Initialize OAuth
	handlers.InitOAuthConfig()

	// Setup Gin
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{getEnv("FRONTEND_URL", "http://localhost:5173")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

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

		// Rounds - admin and team admin can create
		authorized.POST("/rounds", handlers.CreateFeedbackRound)
		authorized.GET("/rounds", middleware.AdminOnly(), handlers.GetAllRounds)
		authorized.GET("/rounds/:id", handlers.GetRoundDetails)
		authorized.PUT("/rounds/:id", handlers.UpdateFeedbackRound)
		authorized.POST("/rounds/:id/reviewers", handlers.AddReviewersToRound)
		authorized.DELETE("/rounds/:id/reviewers/:reviewerId", handlers.RemoveReviewerFromRound)
		authorized.POST("/rounds/:id/start", handlers.StartFeedbackRound)
		authorized.POST("/rounds/:id/close", handlers.CloseFeedbackRound)
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

	// Start server
	log.Println("Server starting on :8080")
	r.Run(":8080")
}
