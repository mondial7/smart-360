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

// Simple MongoDB-based submission handlers
func GetSubmissionDetails(c *gin.Context) {
	submissionID := c.Param("submissionId")
	db := database.GetDB()
	ctx := context.Background()

	// Convert submissionID string to ObjectID
	submissionObjID, err := primitive.ObjectIDFromHex(submissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	var submission models.Submission
	err = db.Collection("submissions").FindOne(ctx, bson.M{"_id": submissionObjID}).Decode(&submission)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	c.JSON(http.StatusOK, submission)
}

func CheckSubmissionStatus(c *gin.Context) {
	roundID := c.Param("roundId")
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

	var submission models.Submission
	err = db.Collection("submissions").FindOne(ctx, bson.M{
		"round_id": roundObjID,
		"reviewer_id": currentUser.ID,
	}).Decode(&submission)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"submitted": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"submitted": true, "submittedAt": submission.SubmittedAt})
}

func SubmitFeedback(c *gin.Context) {
	var req struct {
		RoundID   string `json:"roundId" binding:"required"`
		Responses string `json:"responses" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(req.RoundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Check if submission already exists
	var existing models.Submission
	err = db.Collection("submissions").FindOne(ctx, bson.M{
		"round_id": roundObjID,
		"reviewer_id": currentUser.ID,
	}).Decode(&existing)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Feedback already submitted"})
		return
	}

	// Create submission
	submission := models.Submission{
		RoundID:     roundObjID,
		ReviewerID:  currentUser.ID,
		Responses:   req.Responses,
		SubmittedAt: time.Now(),
	}

	_, err = db.Collection("submissions").InsertOne(ctx, submission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feedback submitted successfully"})
}
