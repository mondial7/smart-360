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
		SubjectID  primitive.ObjectID `json:"subjectId" binding:"required"`
		TemplateID primitive.ObjectID `json:"templateId,omitempty"`
		Deadline   *time.Time         `json:"deadline"`
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

	// Team admins can only create rounds for their team members
	if currentUser.Role == models.RoleTeamAdmin {
		// Verify subject is in the same team
		if currentUser.TeamID == nil || subject.TeamID == nil || *currentUser.TeamID != *subject.TeamID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Team admins can only create rounds for their team members"})
			return
		}
	}

	templateID, err := resolveTemplateIDForCreate(ctx, req.TemplateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	round := models.FeedbackRound{
		SubjectID:   req.SubjectID,
		CreatedByID: currentUser.ID,
		TemplateID:  templateID,
		Deadline:    req.Deadline,
		Status:      models.RoundDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := db.Collection("feedback_rounds").InsertOne(ctx, round)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feedback round"})
		return
	}

	// Set the generated ID
	round.ID = result.InsertedID.(primitive.ObjectID)

	// Create audit log
	createAuditLog(ctx, AuditLogParams{
		Action:       models.AuditRoundCreated,
		ActorID:      currentUser.ID,
		ActorName:    currentUser.Name,
		ActorEmail:   currentUser.Email,
		RoundID:      round.ID,
		RoundSubject: subject.Name,
		Description:  fmt.Sprintf("Created new feedback round for %s", subject.Name),
	})

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

	// Get round subject for audit log
	var subject models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": round.SubjectID}).Decode(&subject)

	// Get current user for audit log
	var actor models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": currentUser.ID}).Decode(&actor)

	// Add reviewers
	fmt.Printf("Adding %d reviewers to round %s\n", len(req.ReviewerIDs), roundObjID.Hex())
	for _, reviewerID := range req.ReviewerIDs {
		// Check if reviewer is the subject (not allowed)
		if reviewerID == round.SubjectID {
			fmt.Printf("  Skipping reviewer %s - cannot assign subject as reviewer\n", reviewerID.Hex())
			continue
		}

		// Get reviewer user info
		var reviewerUser models.User
		err := db.Collection("users").FindOne(ctx, bson.M{"_id": reviewerID}).Decode(&reviewerUser)
		if err != nil {
			fmt.Printf("  Skipping reviewer %s - user not found\n", reviewerID.Hex())
			continue
		}

		// Team admins may only add reviewers from their own team.
		if currentUser.Role == models.RoleTeamAdmin {
			if currentUser.TeamID == nil || reviewerUser.TeamID == nil || *currentUser.TeamID != *reviewerUser.TeamID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Team admins can only add reviewers from their own team"})
				return
			}
		}

		fmt.Printf("  Adding reviewer: %s (%s)\n", reviewerID.Hex(), reviewerUser.Name)
		reviewer := models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			RoundID:    roundObjID,
			ReviewerID: reviewerID,
			CreatedAt:  time.Now(),
		}

		// Add reviewer to embedded array
		_, err = db.Collection("feedback_rounds").UpdateOne(
			ctx,
			bson.M{"_id": roundObjID},
			bson.M{"$push": bson.M{"reviewers": reviewer}},
		)
		if err != nil {
			fmt.Printf("Failed to add reviewer %s: %v\n", reviewerID.Hex(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reviewer"})
			return
		}

		// Create audit log
		createAuditLog(ctx, AuditLogParams{
			Action:       models.AuditReviewerAdded,
			ActorID:      currentUser.ID,
			ActorName:    actor.Name,
			ActorEmail:   actor.Email,
			RoundID:      roundObjID,
			RoundSubject: subject.Name,
			Description:  fmt.Sprintf("Added %s as reviewer", reviewerUser.Name),
			NewValue:     reviewerUser.Name,
		})

		fmt.Printf("  Successfully added reviewer %s\n", reviewerUser.Name)
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

	// Validate reviewer can be removed (hasn't submitted feedback)
	if err := validateReviewerRemoval(ctx, db, roundObjID, reviewerObjID); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Message})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate reviewer removal"})
		}
		return
	}

	// Get reviewer info for audit log
	var reviewerUser models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": reviewerObjID}).Decode(&reviewerUser)

	// Get round subject for audit log
	var subject models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": round.SubjectID}).Decode(&subject)

	// Get current user for audit log
	var actor models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": currentUser.ID}).Decode(&actor)

	// Remove the reviewer from embedded array
	result, err := db.Collection("feedback_rounds").UpdateOne(
		ctx,
		bson.M{"_id": roundObjID},
		bson.M{"$pull": bson.M{"reviewers": bson.M{"reviewer_id": reviewerObjID}}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reviewer"})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reviewer not found in this round"})
		return
	}

	// Create audit log
	createAuditLog(ctx, AuditLogParams{
		Action:       models.AuditReviewerRemoved,
		ActorID:      currentUser.ID,
		ActorName:    actor.Name,
		ActorEmail:   actor.Email,
		RoundID:      roundObjID,
		RoundSubject: subject.Name,
		Description:  fmt.Sprintf("Removed %s as reviewer", reviewerUser.Name),
		OldValue:     reviewerUser.Name,
	})

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

	// Validate status transition (must be Draft)
	if round.Status != models.RoundDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Cannot start round with status '%s'. Only draft rounds can be started.", round.Status),
		})
		return
	}

	oldStatus := round.Status

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

	// Get subject for audit log
	var subject models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": round.SubjectID}).Decode(&subject)

	// Create audit log
	createAuditLog(ctx, AuditLogParams{
		Action:       models.AuditRoundStatusChanged,
		ActorID:      currentUser.ID,
		ActorName:    currentUser.Name,
		ActorEmail:   currentUser.Email,
		RoundID:      roundObjID,
		RoundSubject: subject.Name,
		Description:  fmt.Sprintf("Started feedback round (draft → active)"),
		OldValue:     string(oldStatus),
		NewValue:     string(models.RoundActive),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Round started successfully"})
}

func GetRoundDetails(c *gin.Context) {
	roundID := c.Param("id")
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	populatedRound, err := getPopulatedRound(ctx, db, roundObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Authorization: admin / creator / subject / listed reviewer may view.
	// Reviewer roster is only revealed to admin and creator — exposing it
	// to subjects or peers would deanonymize the 360 feedback.
	isAdmin := currentUser.Role == models.RoleAdmin
	isCreator := currentUser.ID == populatedRound.CreatedByID
	isSubject := currentUser.ID == populatedRound.SubjectID
	isReviewer := false
	for _, r := range populatedRound.Reviewers {
		if r.ReviewerID == currentUser.ID {
			isReviewer = true
			break
		}
	}

	if !isAdmin && !isCreator && !isSubject && !isReviewer {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this round"})
		return
	}

	if !isAdmin && !isCreator {
		populatedRound.Reviewers = []PopulatedRoundReviewer{}
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

	// Validate status transition (must be Active)
	if round.Status != models.RoundActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Cannot close round with status '%s'. Only active rounds can be closed.", round.Status),
		})
		return
	}

	oldStatus := round.Status

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

	// Get subject for audit log
	var subject models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": round.SubjectID}).Decode(&subject)

	// Create audit log
	createAuditLog(ctx, AuditLogParams{
		Action:       models.AuditRoundStatusChanged,
		ActorID:      currentUser.ID,
		ActorName:    currentUser.Name,
		ActorEmail:   currentUser.Email,
		RoundID:      roundObjID,
		RoundSubject: subject.Name,
		Description:  fmt.Sprintf("Closed feedback round (active → closed)"),
		OldValue:     string(oldStatus),
		NewValue:     string(models.RoundClosed),
	})

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

	// Get current subject for audit log
	var oldSubject models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": round.SubjectID}).Decode(&oldSubject)

	// Build update document
	update := bson.M{"$set": bson.M{
		"updated_at": time.Now(),
	}}

	// Handle subject change
	if req.SubjectID != nil {
		// Validate subject change (only allowed in Draft)
		if err := validateSubjectChange(round.Status); err != nil {
			if validationErr, ok := err.(*ValidationError); ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Message})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate subject change"})
			}
			return
		}

		// Check if user is trying to change subject to themselves
		if *req.SubjectID == currentUser.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot set yourself as subject"})
			return
		}

		// Check if subject exists
		var newSubject models.User
		err := db.Collection("users").FindOne(ctx, bson.M{"_id": *req.SubjectID}).Decode(&newSubject)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subject user not found"})
			return
		}

		update["$set"].(bson.M)["subject_id"] = *req.SubjectID

		// Create audit log for subject change
		createAuditLog(ctx, AuditLogParams{
			Action:       models.AuditRoundSubjectChanged,
			ActorID:      currentUser.ID,
			ActorName:    currentUser.Name,
			ActorEmail:   currentUser.Email,
			RoundID:      roundObjID,
			RoundSubject: newSubject.Name,
			Description:  fmt.Sprintf("Changed subject from %s to %s", oldSubject.Name, newSubject.Name),
			OldValue:     oldSubject.Name,
			NewValue:     newSubject.Name,
		})
	}

	// Handle deadline change
	if req.Deadline != nil {
		oldDeadline := "none"
		if round.Deadline != nil {
			oldDeadline = round.Deadline.Format("2006-01-02")
		}
		newDeadline := req.Deadline.Format("2006-01-02")

		update["$set"].(bson.M)["deadline"] = req.Deadline

		// Create audit log for deadline change
		createAuditLog(ctx, AuditLogParams{
			Action:       models.AuditRoundDeadlineChanged,
			ActorID:      currentUser.ID,
			ActorName:    currentUser.Name,
			ActorEmail:   currentUser.Email,
			RoundID:      roundObjID,
			RoundSubject: oldSubject.Name,
			Description:  fmt.Sprintf("Changed deadline from %s to %s", oldDeadline, newDeadline),
			OldValue:     oldDeadline,
			NewValue:     newDeadline,
		})
	}

	// Handle status change
	if req.Status != nil {
		// Validate status transition using new validation function
		if err := validateStatusTransition(round.Status, *req.Status); err != nil {
			if validationErr, ok := err.(*ValidationError); ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Message})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
			}
			return
		}

		update["$set"].(bson.M)["status"] = *req.Status

		// Create audit log for status change
		createAuditLog(ctx, AuditLogParams{
			Action:       models.AuditRoundStatusChanged,
			ActorID:      currentUser.ID,
			ActorName:    currentUser.Name,
			ActorEmail:   currentUser.Email,
			RoundID:      roundObjID,
			RoundSubject: oldSubject.Name,
			Description:  fmt.Sprintf("Changed status from %s to %s", round.Status, *req.Status),
			OldValue:     string(round.Status),
			NewValue:     string(*req.Status),
		})
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
