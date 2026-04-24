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
	// Step 1: Find all rounds where user is the subject AND status is "shared"
	roundCursor, _ := db.Collection("feedback_rounds").Find(ctx, bson.M{
		"subject_id": currentUser.ID,
		"status":     models.RoundShared,
	})
	var roundIDs []interface{}
	if roundCursor != nil {
		defer roundCursor.Close(ctx)
		for roundCursor.Next(ctx) {
			var round models.FeedbackRound
			if err := roundCursor.Decode(&round); err == nil {
				roundIDs = append(roundIDs, round.ID)
			}
		}
	}

	// Step 2: Count consolidations for those rounds where shared_at is not null
	var myFeedbackCount int64
	if len(roundIDs) > 0 {
		myFeedbackCount, _ = db.Collection("consolidations").CountDocuments(ctx, bson.M{
			"round_id":  bson.M{"$in": roundIDs},
			"shared_at": bson.M{"$ne": nil},
		})
	}

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

	// Step 1: Find all rounds where user is the subject AND status is "shared"
	roundCursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{
		"subject_id": currentUser.ID,
		"status":     models.RoundShared,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
		return
	}
	defer roundCursor.Close(ctx)

	// Step 2: Extract round IDs
	var roundIDs []interface{}
	for roundCursor.Next(ctx) {
		var round models.FeedbackRound
		if err := roundCursor.Decode(&round); err != nil {
			continue
		}
		roundIDs = append(roundIDs, round.ID)
	}

	// If no rounds found, return empty array
	if len(roundIDs) == 0 {
		c.JSON(http.StatusOK, []models.Consolidation{})
		return
	}

	// Step 3: Find consolidations for those rounds where shared_at is not null
	cursor, err := db.Collection("consolidations").Find(ctx, bson.M{
		"round_id":  bson.M{"$in": roundIDs},
		"shared_at": bson.M{"$ne": nil},
	})
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
