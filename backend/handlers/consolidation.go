package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"smart360/database"
	"smart360/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// For development/demo purposes - in production, use actual OpenAI/Claude API
func generateMockConsolidation(submissions []models.Submission) map[string]interface{} {
	// Extract all responses
	allStrengths := []string{}
	allImprovements := []string{}
	allBehaviors := []string{}
	allAdvice := []string{}

	for _, sub := range submissions {
		var responses map[string]string
		json.Unmarshal([]byte(sub.Responses), &responses)
		if responses["a"] != "" {
			allStrengths = append(allStrengths, responses["a"])
		}
		if responses["b"] != "" {
			allImprovements = append(allImprovements, responses["b"])
		}
		if responses["c"] != "" {
			allBehaviors = append(allBehaviors, responses["c"])
		}
		if responses["d"] != "" {
			allAdvice = append(allAdvice, responses["d"])
		}
	}

	// Generate mock insights
	strengths := []string{
		"Strong communication skills and ability to collaborate effectively with team members",
		"Technical expertise and problem-solving abilities that drive project success",
		"Reliable and consistent delivery of high-quality work within deadlines",
	}

	improvements := []string{
		"Could benefit from more proactive communication about project status updates",
		"Opportunity to take on more leadership responsibilities in team settings",
		"Consider exploring additional technical skills to broaden expertise",
	}

	insights := []string{
		"Schedule regular check-ins with stakeholders to provide progress updates",
		"Volunteer to lead a small project or initiative to develop leadership skills",
		"Set aside time for learning and development in emerging technologies",
	}

	return map[string]interface{}{
		"executiveSummary":    fmt.Sprintf("Based on feedback from %d reviewers, the subject demonstrates strong technical abilities and collaborative skills. Key themes include effective communication, reliable delivery, and technical expertise. Development opportunities focus on proactive communication and leadership growth.", len(submissions)),
		"strengths":           strengths,
		"areasForImprovement": improvements,
		"actionableInsights":  insights,
		"questionSummaries": map[string]string{
			"a": fmt.Sprintf("Reviewers consistently highlighted strong technical skills and collaborative approach. Common themes: %s", strings.Join(allStrengths[:min(2, len(allStrengths))], "; ")),
			"b": fmt.Sprintf("Areas for growth identified include communication and leadership. Key points: %s", strings.Join(allImprovements[:min(2, len(allImprovements))], "; ")),
			"c": fmt.Sprintf("Specific behaviors noted include proactive problem-solving and team support. Examples: %s", strings.Join(allBehaviors[:min(2, len(allBehaviors))], "; ")),
			"d": fmt.Sprintf("Advice focuses on continued growth in communication and leadership. Suggestions: %s", strings.Join(allAdvice[:min(2, len(allAdvice))], "; ")),
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ConsolidateFeedback(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	roundID := c.Param("id")
	db := database.GetDB()

	// Verify round exists and is closed
	var round models.FeedbackRound
	if err := db.First(&round, roundID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	if round.Status != models.RoundClosed && round.Status != models.RoundActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Round must be active or closed to consolidate"})
		return
	}

	// Check if consolidation already exists
	var existing models.Consolidation
	if err := db.Where("round_id = ?", round.ID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Consolidation already exists for this round", "consolidationId": existing.ID})
		return
	}

	// Get all submissions for this round
	var submissions []models.Submission
	if err := db.Where("round_id = ?", round.ID).Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}

	if len(submissions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No submissions found for this round"})
		return
	}

	// Generate consolidation (mock for now)
	consolidationData := generateMockConsolidation(submissions)

	strengthsJSON, _ := json.Marshal(consolidationData["strengths"])
	improvementsJSON, _ := json.Marshal(consolidationData["areasForImprovement"])
	insightsJSON, _ := json.Marshal(consolidationData["actionableInsights"])
	questionSummariesJSON, _ := json.Marshal(consolidationData["questionSummaries"])

	consolidation := models.Consolidation{
		RoundID:             round.ID,
		GeneratedByID:       currentUser.ID,
		ExecutiveSummary:    consolidationData["executiveSummary"].(string),
		Strengths:           string(strengthsJSON),
		AreasForImprovement: string(improvementsJSON),
		ActionableInsights:  string(insightsJSON),
		QuestionSummaries:   string(questionSummariesJSON),
	}

	if err := db.Create(&consolidation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save consolidation"})
		return
	}

	// Load with relations
	db.Preload("Round.Subject").Preload("GeneratedBy").First(&consolidation, consolidation.ID)

	c.JSON(http.StatusCreated, consolidation)
}

func GetConsolidation(c *gin.Context) {
	roundID := c.Param("roundId")
	db := database.GetDB()

	var consolidation models.Consolidation
	if err := db.Where("round_id = ?", roundID).
		Preload("Round.Subject").
		Preload("Round.Reviewers.Reviewer").
		Preload("GeneratedBy").
		First(&consolidation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consolidation not found"})
		return
	}

	// Parse JSON fields
	var strengths []string
	var improvements []string
	var insights []string
	var questionSummaries map[string]string

	json.Unmarshal([]byte(consolidation.Strengths), &strengths)
	json.Unmarshal([]byte(consolidation.AreasForImprovement), &improvements)
	json.Unmarshal([]byte(consolidation.ActionableInsights), &insights)
	json.Unmarshal([]byte(consolidation.QuestionSummaries), &questionSummaries)

	c.JSON(http.StatusOK, gin.H{
		"id":                  consolidation.ID,
		"roundId":             consolidation.RoundID,
		"round":               consolidation.Round,
		"generatedBy":         consolidation.GeneratedBy,
		"executiveSummary":    consolidation.ExecutiveSummary,
		"strengths":           strengths,
		"areasForImprovement": improvements,
		"actionableInsights":  insights,
		"questionSummaries":   questionSummaries,
		"adminNotes":          consolidation.AdminNotes,
		"sharedAt":            consolidation.SharedAt,
		"createdAt":           consolidation.CreatedAt,
	})
}

func UpdateConsolidationNotes(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	id := c.Param("id")
	db := database.GetDB()

	var consolidation models.Consolidation
	if err := db.First(&consolidation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consolidation not found"})
		return
	}

	var req struct {
		AdminNotes string `json:"adminNotes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	consolidation.AdminNotes = req.AdminNotes
	db.Save(&consolidation)

	c.JSON(http.StatusOK, consolidation)
}

func ShareConsolidation(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	id := c.Param("id")
	db := database.GetDB()

	var consolidation models.Consolidation
	if err := db.Preload("Round").First(&consolidation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consolidation not found"})
		return
	}

	// Only admin or creator can share
	if consolidation.GeneratedByID != currentUser.ID && currentUser.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	now := time.Now()
	consolidation.SharedAt = &now
	db.Save(&consolidation)

	// Update round status to shared
	consolidation.Round.Status = models.RoundShared
	db.Save(&consolidation.Round)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Consolidation shared successfully",
		"sharedAt": consolidation.SharedAt,
	})
}

func GetMyConsolidatedFeedback(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	db := database.GetDB()

	// Find all consolidations for rounds where current user is the subject and has been shared
	var consolidations []models.Consolidation
	if err := db.Joins("JOIN feedback_rounds ON feedback_rounds.id = consolidations.round_id").
		Where("feedback_rounds.subject_id = ? AND consolidations.shared_at IS NOT NULL", currentUser.ID).
		Preload("Round.Subject").
		Preload("Round.Reviewers.Reviewer").
		Preload("GeneratedBy").
		Find(&consolidations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feedback"})
		return
	}

	c.JSON(http.StatusOK, consolidations)
}
