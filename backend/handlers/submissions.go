package handlers

import (
	"encoding/json"
	"net/http"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
)

type SubmissionRequest struct {
	RoundID   uint              `json:"roundId" binding:"required"`
	Responses map[string]string `json:"responses" binding:"required"`
}

var StandardQuestions = []string{
	"What are this person's key strengths?",
	"What areas could this person improve?",
	"What specific behaviors or actions have you observed that stood out?",
	"What advice would you give to help this person grow?",
}

func SubmitFeedback(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var req SubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Verify round exists and is active
	var round models.FeedbackRound
	if err := db.First(&round, req.RoundID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	if round.Status != models.RoundActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Round is not active"})
		return
	}

	// Check deadline
	if round.Deadline != nil && round.Deadline.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Round deadline has passed"})
		return
	}

	// Verify user is assigned reviewer
	var reviewer models.RoundReviewer
	if err := db.Where("round_id = ? AND reviewer_id = ?", req.RoundID, currentUser.ID).First(&reviewer).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not assigned to review this round"})
		return
	}

	// Check if already submitted
	var existing models.Submission
	if err := db.Where("round_id = ? AND reviewer_id = ?", req.RoundID, currentUser.ID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already submitted feedback for this round"})
		return
	}

	// Validate all questions answered
	for i := range StandardQuestions {
		key := string(rune('a' + i)) // a, b, c, d
		if _, ok := req.Responses[key]; !ok || req.Responses[key] == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please answer all questions"})
			return
		}
	}

	// Convert responses to JSON
	responsesJSON, err := json.Marshal(req.Responses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process responses"})
		return
	}

	// Create submission
	submission := models.Submission{
		RoundID:     req.RoundID,
		ReviewerID:  currentUser.ID,
		Responses:   string(responsesJSON),
		SubmittedAt: time.Now(),
	}

	if err := db.Create(&submission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Feedback submitted successfully",
		"submissionId": submission.ID,
	})
}

func CheckSubmissionStatus(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	roundID := c.Param("roundId")

	db := database.GetDB()
	var count int64
	db.Model(&models.Submission{}).Where("round_id = ? AND reviewer_id = ?", roundID, currentUser.ID).Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"submitted": count > 0,
	})
}

func GetSubmissionDetails(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	roundID := c.Param("roundId")

	db := database.GetDB()
	var submission models.Submission
	if err := db.Where("round_id = ? AND reviewer_id = ?", roundID, currentUser.ID).First(&submission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	// Parse responses
	var responses map[string]string
	json.Unmarshal([]byte(submission.Responses), &responses)

	c.JSON(http.StatusOK, gin.H{
		"id":          submission.ID,
		"roundId":     submission.RoundID,
		"responses":   responses,
		"submittedAt": submission.SubmittedAt,
	})
}
