package handlers

import (
	"context"
	"fmt"
	"net/http"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Simple MongoDB-based submission handlers
func GetRoundSubmissions(c *gin.Context) {
	roundID := c.Param("roundId")
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Authorization: only the round creator or a global admin may read raw
	// submissions, since the (reviewer_id, responses) pairing breaks the
	// 360-feedback anonymity guarantee.
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}
	if currentUser.Role != models.RoleAdmin && round.CreatedByID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view submissions for this round"})
		return
	}

	cursor, err := db.Collection("submissions").Find(ctx, bson.M{"round_id": roundObjID})
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

func GetSubmissionDetails(c *gin.Context) {
	submissionID := c.Param("submissionId")
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

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

	// Authorization: only the reviewer who wrote it, the round creator, or
	// a global admin may read the (reviewer_id, responses) pairing.
	if currentUser.Role != models.RoleAdmin && submission.ReviewerID != currentUser.ID {
		var round models.FeedbackRound
		err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": submission.RoundID}).Decode(&round)
		if err != nil || round.CreatedByID != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this submission"})
			return
		}
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
		"round_id":    roundObjID,
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
		RoundID              string                      `json:"roundId" binding:"required"`
		Responses            string                      `json:"responses" binding:"required"`
		Relationship         models.ReviewerRelationship `json:"relationship,omitempty"`
		InteractionFrequency models.InteractionFrequency `json:"interactionFrequency,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	roundObjID, err := primitive.ObjectIDFromHex(req.RoundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Validate round exists, is active, caller is an enlisted reviewer, and
	// caller is not the round subject. Without this check any authenticated
	// user can inject phantom submissions into any round in any status.
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}
	if round.Status != models.RoundActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Round is not accepting submissions"})
		return
	}
	isSelf := round.SubjectID == currentUser.ID
	if !isSelf {
		isReviewer := false
		for _, r := range round.Reviewers {
			if r.ReviewerID == currentUser.ID {
				isReviewer = true
				break
			}
		}
		if !isReviewer {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not a reviewer for this round"})
			return
		}
		// Peer feedback must declare the reviewer's relationship and how often
		// they interact with the subject. Without this metadata the AI cannot
		// down-weight thin signals and the resulting consolidation drifts toward
		// false equivalence between a daily collaborator and a one-off contact.
		if !req.Relationship.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing relationship (manager, report, peer, cross_functional)"})
			return
		}
		if !req.InteractionFrequency.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing interactionFrequency (daily, weekly, monthly, rarely)"})
			return
		}
	}

	var existing models.Submission
	err = db.Collection("submissions").FindOne(ctx, bson.M{
		"round_id":    roundObjID,
		"reviewer_id": currentUser.ID,
	}).Decode(&existing)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Feedback already submitted"})
		return
	}

	submission := models.Submission{
		RoundID:     roundObjID,
		ReviewerID:  currentUser.ID,
		Responses:   req.Responses,
		IsSelf:      isSelf,
		SubmittedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}
	if !isSelf {
		submission.Relationship = req.Relationship
		submission.InteractionFrequency = req.InteractionFrequency
	}

	_, err = db.Collection("submissions").InsertOne(ctx, submission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feedback submitted successfully"})
}

func UpdateSubmission(c *gin.Context) {
	submissionID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Convert submissionID string to ObjectID
	submissionObjID, err := primitive.ObjectIDFromHex(submissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	var req struct {
		Responses string `json:"responses" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing submission to verify ownership
	var existingSubmission models.Submission
	err = db.Collection("submissions").FindOne(ctx, bson.M{"_id": submissionObjID}).Decode(&existingSubmission)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	// Verify user owns this submission
	if existingSubmission.ReviewerID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to edit this submission"})
		return
	}

	// Check if round is still active
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": existingSubmission.RoundID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	if round.Status != models.RoundActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot edit feedback: round is no longer active"})
		return
	}

	// Update submission
	update := bson.M{"$set": bson.M{
		"responses":  req.Responses,
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("submissions").UpdateOne(ctx, bson.M{"_id": submissionObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update submission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback updated successfully"})
}

func DebugSubmissions(c *gin.Context) {
	db := database.GetDB()
	ctx := context.Background()

	fmt.Printf("DebugSubmissions called\n")

	// Get all submissions
	cursor, err := db.Collection("submissions").Find(ctx, bson.M{})
	if err != nil {
		fmt.Printf("Error finding all submissions: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}
	defer cursor.Close(ctx)

	var submissions []models.Submission
	if err = cursor.All(ctx, &submissions); err != nil {
		fmt.Printf("Error decoding all submissions: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode submissions"})
		return
	}

	fmt.Printf("Total submissions in database: %d\n", len(submissions))

	// Log each submission details
	for i, sub := range submissions {
		fmt.Printf("Submission %d: ID=%s, RoundID=%s, ReviewerID=%s, SubmittedAt=%v\n",
			i+1, sub.ID.Hex(), sub.RoundID.Hex(), sub.ReviewerID.Hex(), sub.SubmittedAt)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":       len(submissions),
		"submissions": submissions,
	})
}
