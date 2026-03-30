package handlers

import (
	"context"
	"net/http"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUsers(c *gin.Context) {
	db := database.GetDB()
	ctx := context.Background()

	cursor, err := db.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	// Convert id string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	err = db.Collection("users").FindOne(ctx, bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	var req struct {
		Role models.UserRole `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert id string to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	update := bson.M{"$set": bson.M{
		"role":       req.Role,
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("users").UpdateOne(ctx, bson.M{"_id": userObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// UserWithFeedbackStats enriches User with feedback metrics
type UserWithFeedbackStats struct {
	models.User
	LastFeedbackReceived  *time.Time `json:"lastFeedbackReceived"`
	ActiveRoundsAsSubject int64      `json:"activeRoundsAsSubject"`
	PendingReviews        int64      `json:"pendingReviews"`
	TotalFeedbackReceived int64      `json:"totalFeedbackReceived"`
}

// GetUsersWithFeedbackStats returns all users enriched with feedback statistics
func GetUsersWithFeedbackStats(c *gin.Context) {
	db := database.GetDB()
	ctx := context.Background()

	// Get all users
	cursor, err := db.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
		return
	}

	// Enrich each user with feedback stats
	var enrichedUsers []UserWithFeedbackStats
	for _, user := range users {
		stats := UserWithFeedbackStats{User: user}

		// Calculate stats for this user
		stats.LastFeedbackReceived = getLastFeedbackReceived(ctx, db, user.ID)
		stats.ActiveRoundsAsSubject = getActiveRoundsCount(ctx, db, user.ID)
		stats.PendingReviews = getPendingReviewsCount(ctx, db, user.ID)
		stats.TotalFeedbackReceived = getTotalFeedbackCount(ctx, db, user.ID)

		enrichedUsers = append(enrichedUsers, stats)
	}

	c.JSON(http.StatusOK, enrichedUsers)
}

// Helper: Get last feedback received date (most recent consolidation shared_at)
func getLastFeedbackReceived(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) *time.Time {
	// Find all rounds where user is the subject
	cursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{"subject_id": userID})
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	var rounds []models.FeedbackRound
	cursor.All(ctx, &rounds)

	var latestSharedAt *time.Time

	// For each round, check if consolidation exists and is shared
	for _, round := range rounds {
		var consolidation models.Consolidation
		err := db.Collection("consolidations").FindOne(ctx, bson.M{
			"round_id": round.ID,
		}).Decode(&consolidation)

		if err == nil && consolidation.SharedAt != nil {
			if latestSharedAt == nil || consolidation.SharedAt.After(*latestSharedAt) {
				latestSharedAt = consolidation.SharedAt
			}
		}
	}

	return latestSharedAt
}

// Helper: Get count of active rounds where user is the subject
func getActiveRoundsCount(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) int64 {
	count, _ := db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{
		"subject_id": userID,
		"status":     models.RoundActive,
	})
	return count
}

// Helper: Get count of pending reviews (active rounds where user is reviewer but hasn't submitted)
func getPendingReviewsCount(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) int64 {
	// Find all active rounds where user is a reviewer
	cursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{
		"reviewers.reviewer_id": userID,
		"status":                models.RoundActive,
	})
	if err != nil {
		return 0
	}
	defer cursor.Close(ctx)

	var rounds []models.FeedbackRound
	cursor.All(ctx, &rounds)

	var pendingCount int64
	for _, round := range rounds {
		// Check if user has submitted feedback for this round
		var submission models.Submission
		err := db.Collection("submissions").FindOne(ctx, bson.M{
			"round_id":    round.ID,
			"reviewer_id": userID,
		}).Decode(&submission)

		// If no submission found, it's pending
		if err != nil {
			pendingCount++
		}
	}

	return pendingCount
}

// Helper: Get total count of feedback received (shared consolidations)
func getTotalFeedbackCount(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) int64 {
	// Find all shared rounds where user is the subject
	cursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{
		"subject_id": userID,
		"status":     models.RoundShared,
	})
	if err != nil {
		return 0
	}
	defer cursor.Close(ctx)

	var rounds []models.FeedbackRound
	cursor.All(ctx, &rounds)

	var sharedCount int64
	for _, round := range rounds {
		// Verify consolidation exists and is shared
		var consolidation models.Consolidation
		err := db.Collection("consolidations").FindOne(ctx, bson.M{
			"round_id": round.ID,
		}).Decode(&consolidation)

		if err == nil && consolidation.SharedAt != nil {
			sharedCount++
		}
	}

	return sharedCount
}
