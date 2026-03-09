package handlers

import (
	"net/http"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateRoundRequest struct {
	SubjectID   uint   `json:"subjectId" binding:"required"`
	ReviewerIDs []uint `json:"reviewerIds" binding:"required,min=1"`
	Deadline    string `json:"deadline" binding:"required"`
}

func CreateRound(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var req CreateRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse deadline
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format"})
		return
	}

	// Ensure deadline is in the future
	if deadline.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deadline must be in the future"})
		return
	}

	db := database.GetDB()

	// Verify subject exists and is not the creator
	var subject models.User
	if err := db.First(&subject, req.SubjectID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subject not found"})
		return
	}

	if subject.ID == currentUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create feedback round for yourself"})
		return
	}

	// Verify all reviewers exist and are not the subject
	var reviewers []models.User
	if err := db.Where("id IN ?", req.ReviewerIDs).Find(&reviewers).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reviewers"})
		return
	}

	if len(reviewers) != len(req.ReviewerIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Some reviewers not found"})
		return
	}

	// Filter out subject from reviewers if included
	validReviewerIDs := make([]uint, 0)
	for _, reviewer := range reviewers {
		if reviewer.ID != req.SubjectID {
			validReviewerIDs = append(validReviewerIDs, reviewer.ID)
		}
	}

	if len(validReviewerIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one reviewer required (subject cannot review themselves)"})
		return
	}

	// Create round
	round := models.FeedbackRound{
		SubjectID:   req.SubjectID,
		CreatedByID: currentUser.ID,
		Deadline:    &deadline,
		Status:      models.RoundActive,
	}

	if err := db.Create(&round).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create round"})
		return
	}

	// Create reviewer assignments
	for _, reviewerID := range validReviewerIDs {
		reviewer := models.RoundReviewer{
			RoundID:    round.ID,
			ReviewerID: reviewerID,
		}
		db.Create(&reviewer)
	}

	// Load round with relationships
	db.Preload("Subject").Preload("CreatedBy").Preload("Reviewers.Reviewer").First(&round, round.ID)

	c.JSON(http.StatusCreated, round)
}

func GetRounds(c *gin.Context) {
	db := database.GetDB()

	var rounds []models.FeedbackRound
	if err := db.Preload("Subject").Preload("CreatedBy").Preload("Reviewers.Reviewer").
		Order("created_at DESC").Find(&rounds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
		return
	}

	c.JSON(http.StatusOK, rounds)
}

func GetRound(c *gin.Context) {
	id := c.Param("id")

	db := database.GetDB()
	var round models.FeedbackRound

	if err := db.Preload("Subject").Preload("CreatedBy").Preload("Reviewers.Reviewer").
		First(&round, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	c.JSON(http.StatusOK, round)
}

func CloseRound(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	id := c.Param("id")
	db := database.GetDB()

	var round models.FeedbackRound
	if err := db.First(&round, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Only creator or admin can close
	if round.CreatedByID != currentUser.ID && currentUser.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	round.Status = models.RoundClosed
	db.Save(&round)

	c.JSON(http.StatusOK, round)
}

func GetMyPendingReviews(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	db := database.GetDB()

	// Find all rounds where user is a reviewer and hasn't submitted yet
	var rounds []models.FeedbackRound
	err := db.Joins("JOIN round_reviewers ON round_reviewers.round_id = feedback_rounds.id").
		Where("round_reviewers.reviewer_id = ?", currentUser.ID).
		Where("feedback_rounds.status = ?", models.RoundActive).
		Where("feedback_rounds.deadline > ?", time.Now()).
		Preload("Subject").
		Preload("CreatedBy").
		Preload("Reviewers.Reviewer").
		Find(&rounds).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	// Check submission status for each
	var pending []models.FeedbackRound
	for _, round := range rounds {
		var count int64
		db.Model(&models.Submission{}).Where("round_id = ? AND reviewer_id = ?", round.ID, currentUser.ID).Count(&count)
		if count == 0 {
			pending = append(pending, round)
		}
	}

	c.JSON(http.StatusOK, pending)
}
