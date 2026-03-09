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
)

func CreateFeedbackRound(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var req struct {
		SubjectID primitive.ObjectID `json:"subjectId" binding:"required"`
		Deadline  *time.Time         `json:"deadline"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user is trying to create a round for themselves
	if req.SubjectID == currentUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create feedback round for yourself"})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Check if subject exists
	var subject models.User
	err := db.Collection("users").FindOne(ctx, bson.M{"_id": req.SubjectID}).Decode(&subject)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subject user not found"})
		return
	}

	round := models.FeedbackRound{
		SubjectID:   req.SubjectID,
		CreatedByID: currentUser.ID,
		Deadline:    req.Deadline,
		Status:      models.RoundDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = db.Collection("feedback_rounds").InsertOne(ctx, round)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feedback round"})
		return
	}

	c.JSON(http.StatusCreated, round)
}

func AddReviewersToRound(c *gin.Context) {
	roundID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var req struct {
		ReviewerIDs []primitive.ObjectID `json:"reviewerIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Verify round exists and user owns it
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	if round.CreatedByID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this round"})
		return
	}

	// Add reviewers
	for _, reviewerID := range req.ReviewerIDs {
		reviewer := models.RoundReviewer{
			RoundID:    roundObjID,
			ReviewerID: reviewerID,
			CreatedAt:  time.Now(),
		}
		_, err = db.Collection("round_reviewers").InsertOne(ctx, reviewer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reviewer"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reviewers added successfully"})
}

func StartFeedbackRound(c *gin.Context) {
	roundID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Verify round exists and user owns it
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	if round.CreatedByID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this round"})
		return
	}

	// Update round status
	update := bson.M{"$set": bson.M{
		"status": models.RoundActive,
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("feedback_rounds").UpdateOne(ctx, bson.M{"_id": roundObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start round"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Round started successfully"})
}

func GetRoundDetails(c *gin.Context) {
	roundID := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	c.JSON(http.StatusOK, round)
}

func CloseFeedbackRound(c *gin.Context) {
	roundID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Verify round exists and user owns it
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	if round.CreatedByID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this round"})
		return
	}

	// Update round status
	update := bson.M{"$set": bson.M{
		"status": models.RoundClosed,
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("feedback_rounds").UpdateOne(ctx, bson.M{"_id": roundObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close round"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Round closed successfully"})
}
