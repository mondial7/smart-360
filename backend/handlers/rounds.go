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
	fmt.Printf("Adding %d reviewers to round %s\n", len(req.ReviewerIDs), roundObjID.Hex())
	for _, reviewerID := range req.ReviewerIDs {
		// Check if reviewer is the subject (not allowed)
		if reviewerID == round.SubjectID {
			fmt.Printf("  Skipping reviewer %s - cannot assign subject as reviewer\n", reviewerID.Hex())
			continue
		}

		fmt.Printf("  Adding reviewer: %s\n", reviewerID.Hex())
		reviewer := models.RoundReviewer{
			RoundID:    roundObjID,
			ReviewerID: reviewerID,
			CreatedAt:  time.Now(),
		}
		result, err := db.Collection("round_reviewers").InsertOne(ctx, reviewer)
		if err != nil {
			fmt.Printf("Failed to add reviewer %s: %v\n", reviewerID.Hex(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reviewer"})
			return
		}
		fmt.Printf("  Successfully added reviewer with ID: %s\n", result.InsertedID.(primitive.ObjectID).Hex())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reviewers added successfully"})
}

func RemoveReviewerFromRound(c *gin.Context) {
	roundID := c.Param("id")
	reviewerID := c.Param("reviewerId")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	db := database.GetDB()
	ctx := context.Background()

	// Convert IDs to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	reviewerObjID, err := primitive.ObjectIDFromHex(reviewerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reviewer ID"})
		return
	}

	// Verify round exists and user owns it (or is admin)
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	if currentUser.Role != models.RoleAdmin && round.CreatedByID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this round"})
		return
	}

	// Remove the reviewer
	result, err := db.Collection("round_reviewers").DeleteOne(ctx, bson.M{
		"round_id":    roundObjID,
		"reviewer_id": reviewerObjID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reviewer"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reviewer not found in this round"})
		return
	}

	fmt.Printf("Removed reviewer %s from round %s\n", reviewerObjID.Hex(), roundObjID.Hex())
	c.JSON(http.StatusOK, gin.H{"message": "Reviewer removed successfully"})
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
		"status":     models.RoundActive,
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

	// Get populated round data
	populatedRound, err := getPopulatedRound(ctx, db, roundObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	c.JSON(http.StatusOK, populatedRound)
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
		"status":     models.RoundClosed,
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("feedback_rounds").UpdateOne(ctx, bson.M{"_id": roundObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close round"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Round closed successfully"})
}

func UpdateFeedbackRound(c *gin.Context) {
	roundID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var req struct {
		SubjectID *primitive.ObjectID `json:"subjectId"`
		Deadline  *time.Time          `json:"deadline"`
		Status    *models.RoundStatus `json:"status"`
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

	// Verify round exists
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Allow admin or round owner to edit
	if currentUser.Role != models.RoleAdmin && round.CreatedByID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this round"})
		return
	}

	// Build update document
	update := bson.M{"$set": bson.M{
		"updated_at": time.Now(),
	}}

	if req.SubjectID != nil {
		// Check if user is trying to change subject to themselves
		if *req.SubjectID == currentUser.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot set yourself as subject"})
			return
		}

		// Check if subject exists
		var subject models.User
		err := db.Collection("users").FindOne(ctx, bson.M{"_id": *req.SubjectID}).Decode(&subject)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subject user not found"})
			return
		}

		update["$set"].(bson.M)["subject_id"] = *req.SubjectID
	}

	if req.Deadline != nil {
		update["$set"].(bson.M)["deadline"] = req.Deadline
	}

	if req.Status != nil {
		// Validate status transition
		if !isValidStatusTransition(round.Status, *req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
			return
		}
		update["$set"].(bson.M)["status"] = *req.Status
	}

	_, err = db.Collection("feedback_rounds").UpdateOne(ctx, bson.M{"_id": roundObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update round"})
		return
	}

	// Return updated round
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated round"})
		return
	}

	c.JSON(http.StatusOK, round)
}

func isValidStatusTransition(current, new models.RoundStatus) bool {
	// Define allowed transitions
	validTransitions := map[models.RoundStatus][]models.RoundStatus{
		models.RoundDraft:  {models.RoundDraft, models.RoundActive},
		models.RoundActive: {models.RoundActive, models.RoundClosed},
		models.RoundClosed: {models.RoundClosed, models.RoundShared},
		models.RoundShared: {models.RoundShared},
	}

	allowedStatuses, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == new {
			return true
		}
	}

	return false
}
