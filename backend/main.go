package main

import (
	"log"
	"smart360/database"
	"smart360/handlers"
	"smart360/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	database.InitDB()

	// Seed development data
	database.SeedData()

	// Initialize OAuth
	handlers.InitOAuthConfig()

	// Setup Gin
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public routes
	r.GET("/api/auth/google", handlers.GetGoogleAuthURL)
	r.GET("/api/auth/callback", handlers.GoogleCallback)
	r.GET("/api/auth/dev-login", handlers.DevLogin) // Development only

	// Protected routes
	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.GET("/me", handlers.GetCurrentUser)
		authorized.GET("/users", handlers.GetUsers)
		authorized.GET("/users/:id", handlers.GetUserByID)
		authorized.PUT("/users/:id/role", middleware.AdminOnly(), handlers.UpdateUserRole)

		// Rounds - admin only for creation
		authorized.POST("/rounds", middleware.AdminOnly(), handlers.CreateFeedbackRound)
		authorized.GET("/rounds", handlers.GetMyRounds)
		authorized.GET("/rounds/:id", handlers.GetRoundDetails)
		authorized.POST("/rounds/:id/reviewers", handlers.AddReviewersToRound)
		authorized.POST("/rounds/:id/start", handlers.StartFeedbackRound)
		authorized.POST("/rounds/:id/close", handlers.CloseFeedbackRound)
		authorized.GET("/rounds-for-me", handlers.GetRoundsForMe)

		// Submissions
		authorized.POST("/submissions", handlers.SubmitFeedback)
		authorized.GET("/submissions/check/:roundId", handlers.CheckSubmissionStatus)
		authorized.GET("/submissions/:roundId", handlers.GetSubmissionDetails)

		// Consolidations (admin only for generation)
		authorized.POST("/rounds/:id/consolidate", middleware.AdminOnly(), handlers.ConsolidateFeedback)
		authorized.GET("/consolidations/:roundId", handlers.GetConsolidation)
		authorized.PUT("/consolidations/:id/notes", middleware.AdminOnly(), handlers.UpdateConsolidationNotes)
		authorized.POST("/consolidations/:id/share", middleware.AdminOnly(), handlers.ShareConsolidation)

		// Dashboard
		authorized.GET("/dashboard/stats", handlers.GetDashboardStats)
		authorized.GET("/dashboard/my-rounds", handlers.GetMyRounds)
		authorized.GET("/dashboard/rounds-for-me", handlers.GetRoundsForMe)
		authorized.GET("/dashboard/my-submissions", handlers.GetMySubmissions)
		authorized.GET("/dashboard/my-consolidations", handlers.GetMyConsolidations)
	}

	// Start server
	log.Println("Server starting on :8080")
	r.Run(":8080")
}
