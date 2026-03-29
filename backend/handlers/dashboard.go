package handlers

import (
	"context"
	"net/http"
	"smart360/database"
	"smart360/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetDashboardStats(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Get total users
	totalUsers, _ := db.Collection("users").CountDocuments(ctx, bson.M{})

	// Get admin and member counts
	adminCount, _ := db.Collection("users").CountDocuments(ctx, bson.M{"role": models.RoleAdmin})
	memberCount, _ := db.Collection("users").CountDocuments(ctx, bson.M{"role": models.RoleMember})

	// Get total rounds
	totalRounds, _ := db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{})

	// Get active rounds
	activeRounds, _ := db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{"status": models.RoundActive})

	// Get rounds created by current user
	myRounds, _ := db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{"created_by_id": currentUser.ID})

	// Get rounds where user is subject
	subjectRounds, _ := db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{"subject_id": currentUser.ID})

	// Get submissions by current user
	mySubmissions, _ := db.Collection("submissions").CountDocuments(ctx, bson.M{"reviewer_id": currentUser.ID})

	// Get consolidations for user (where they are the subject and it's shared)
	myFeedbackCount, _ := db.Collection("consolidations").CountDocuments(ctx, bson.M{
		"generated_by_id": currentUser.ID,
	})

	stats := gin.H{
		"totalUsers":      totalUsers,
		"adminCount":      adminCount,
		"memberCount":     memberCount,
		"totalRounds":     totalRounds,
		"activeRounds":    activeRounds,
		"myRounds":        myRounds,
		"subjectRounds":   subjectRounds,
		"mySubmissions":   mySubmissions,
		"myFeedbackCount": myFeedbackCount,
	}

	c.JSON(http.StatusOK, stats)
}

func GetAllRounds(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Only admins can get all rounds
	if currentUser.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Get all rounds
	cursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
		return
	}
	defer cursor.Close(ctx)

	var rounds []models.FeedbackRound
	if err = cursor.All(ctx, &rounds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode rounds"})
		return
	}

	// Populate each round
	var populatedRounds []PopulatedRound
	for _, round := range rounds {
		populatedRound, err := getPopulatedRound(ctx, db, round.ID)
		if err != nil {
			continue // Skip rounds that can't be populated
		}
		populatedRounds = append(populatedRounds, *populatedRound)
	}

	c.JSON(http.StatusOK, populatedRounds)
}

func GetMyRounds(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Get rounds created by current user
	cursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{"created_by_id": currentUser.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
		return
	}
	defer cursor.Close(ctx)

	var rounds []models.FeedbackRound
	if err = cursor.All(ctx, &rounds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode rounds"})
		return
	}

	// Populate each round
	var populatedRounds []PopulatedRound
	for _, round := range rounds {
		populatedRound, err := getPopulatedRound(ctx, db, round.ID)
		if err != nil {
			continue // Skip rounds that can't be populated
		}
		populatedRounds = append(populatedRounds, *populatedRound)
	}

	c.JSON(http.StatusOK, populatedRounds)
}

func GetRoundsForMe(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Find all rounds where the current user is in the reviewers array
	cursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{
		"reviewers.reviewer_id": currentUser.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
		return
	}
	defer cursor.Close(ctx)

	var rounds []models.FeedbackRound
	if err = cursor.All(ctx, &rounds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode rounds"})
		return
	}

	// Populate each round
	var populatedRounds []PopulatedRound
	for _, round := range rounds {
		populatedRound, err := getPopulatedRound(ctx, db, round.ID)
		if err != nil {
			continue // Skip rounds that can't be populated
		}
		populatedRounds = append(populatedRounds, *populatedRound)
	}

	c.JSON(http.StatusOK, populatedRounds)
}

func GetMySubmissions(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Get submissions by current user
	cursor, err := db.Collection("submissions").Find(ctx, bson.M{"reviewer_id": currentUser.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}
	defer cursor.Close(ctx)

	var submissions []models.Submission
	if err = cursor.All(ctx, &submissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode submissions"})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

func GetMyConsolidations(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Get consolidations generated by current user
	cursor, err := db.Collection("consolidations").Find(ctx, bson.M{"generated_by_id": currentUser.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consolidations"})
		return
	}
	defer cursor.Close(ctx)

	var consolidations []models.Consolidation
	if err = cursor.All(ctx, &consolidations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode consolidations"})
		return
	}

	c.JSON(http.StatusOK, consolidations)
}
