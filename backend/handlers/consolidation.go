package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smart360/database"
	"smart360/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Simple MongoDB-based consolidation handlers
func ConsolidateFeedback(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	roundID := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Check if OpenAI key is available
	openAIKey := os.Getenv("OPENAI_API_KEY")
	hasOpenAI := openAIKey != ""

	// Get all submissions for this round
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

	if len(submissions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No feedback submissions found to consolidate"})
		return
	}

	var consolidation models.Consolidation

	if hasOpenAI {
		// Use OpenAI for AI consolidation (mock for now)
		consolidation = models.Consolidation{
			RoundID:             roundObjID,
			GeneratedByID:       currentUser.ID,
			ExecutiveSummary:    "This is a mock executive summary for development purposes.",
			Strengths:           []string{"Good communication", "Team collaboration", "Technical skills"},
			AreasForImprovement: []string{"Time management", "Documentation", "Code reviews"},
			ActionableInsights:  []string{"Focus on prioritization", "Improve documentation practices", "Implement regular code reviews"},
			QuestionSummaries:   map[string]string{"a": "Summary of responses for question 1", "b": "Summary of responses for question 2", "c": "Summary of responses for question 3", "d": "Summary of responses for question 4"},
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
	} else {
		// Combine actual feedback submissions
		consolidation = combineFeedbackSubmissions(submissions, roundObjID, currentUser.ID)
	}

	_, err = db.Collection("consolidations").InsertOne(ctx, consolidation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create consolidation"})
		return
	}

	c.JSON(http.StatusCreated, consolidation)
}

func GetConsolidation(c *gin.Context) {
	roundID := c.Param("roundId")
	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	var consolidation models.Consolidation
	err = db.Collection("consolidations").FindOne(ctx, bson.M{"round_id": roundObjID}).Decode(&consolidation)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consolidation not found"})
		return
	}

	c.JSON(http.StatusOK, consolidation)
}

func UpdateConsolidationNotes(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	var req struct {
		AdminNotes string `json:"adminNotes" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert id string to ObjectID
	consolidationObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consolidation ID"})
		return
	}

	update := bson.M{"$set": bson.M{
		"admin_notes": req.AdminNotes,
		"updated_at":  time.Now(),
	}}

	_, err = db.Collection("consolidations").UpdateOne(ctx, bson.M{"_id": consolidationObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update consolidation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consolidation notes updated successfully"})
}

func ShareConsolidation(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	// Convert id string to ObjectID
	consolidationObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consolidation ID"})
		return
	}

	update := bson.M{"$set": bson.M{
		"shared_at":  time.Now(),
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("consolidations").UpdateOne(ctx, bson.M{"_id": consolidationObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share consolidation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consolidation shared successfully"})
}

func UpdateConsolidation(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	// Convert id string to ObjectID
	consolidationObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consolidation ID"})
		return
	}

	var req struct {
		ExecutiveSummary    string            `json:"executiveSummary,omitempty"`
		Strengths           []string          `json:"strengths,omitempty"`
		AreasForImprovement []string          `json:"areasForImprovement,omitempty"`
		ActionableInsights  []string          `json:"actionableInsights,omitempty"`
		QuestionSummaries   map[string]string `json:"questionSummaries,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update document with only provided fields
	update := bson.M{"$set": bson.M{
		"updated_at": time.Now(),
	}}

	if req.ExecutiveSummary != "" {
		update["$set"].(bson.M)["executive_summary"] = req.ExecutiveSummary
	}
	if req.Strengths != nil {
		update["$set"].(bson.M)["strengths"] = req.Strengths
	}
	if req.AreasForImprovement != nil {
		update["$set"].(bson.M)["areas_for_improvement"] = req.AreasForImprovement
	}
	if req.ActionableInsights != nil {
		update["$set"].(bson.M)["actionable_insights"] = req.ActionableInsights
	}
	if req.QuestionSummaries != nil {
		update["$set"].(bson.M)["question_summaries"] = req.QuestionSummaries
	}

	_, err = db.Collection("consolidations").UpdateOne(ctx, bson.M{"_id": consolidationObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update consolidation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consolidation updated successfully"})
}

func GetMyConsolidatedFeedback(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Find consolidations where user is the subject
	cursor, err := db.Collection("consolidations").Find(ctx, bson.M{
		"generated_by_id": currentUser.ID,
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

func combineFeedbackSubmissions(submissions []models.Submission, roundID primitive.ObjectID, generatedByID primitive.ObjectID) models.Consolidation {
	var allStrengths []string
	var allImprovements []string
	var allBehaviors []string
	var allAdvice []string
	questionSummaries := make(map[string]string)

	for _, submission := range submissions {
		// Parse the JSON responses
		var responses map[string]string
		if err := json.Unmarshal([]byte(submission.Responses), &responses); err != nil {
			continue
		}

		// Collect responses for each question
		if strengths, ok := responses["a"]; ok && strengths != "" {
			allStrengths = append(allStrengths, strengths)
		}
		if improvements, ok := responses["b"]; ok && improvements != "" {
			allImprovements = append(allImprovements, improvements)
		}
		if behaviors, ok := responses["c"]; ok && behaviors != "" {
			allBehaviors = append(allBehaviors, behaviors)
		}
		if advice, ok := responses["d"]; ok && advice != "" {
			allAdvice = append(allAdvice, advice)
		}
	}

	// Create executive summary from all responses
	executiveSummary := "Consolidated feedback from " + fmt.Sprintf("%d", len(submissions)) + " reviewers. "
	if len(allStrengths) > 0 {
		executiveSummary += "Key strengths identified include communication and collaboration. "
	}
	if len(allImprovements) > 0 {
		executiveSummary += "Areas for improvement focus on documentation and process. "
	}

	// Create question summaries
	if len(allStrengths) > 0 {
		questionSummaries["a"] = "Reviewers highlighted: " + strings.Join(allStrengths, "; ")
	}
	if len(allImprovements) > 0 {
		questionSummaries["b"] = "Suggested improvements: " + strings.Join(allImprovements, "; ")
	}
	if len(allBehaviors) > 0 {
		questionSummaries["c"] = "Observed behaviors: " + strings.Join(allBehaviors, "; ")
	}
	if len(allAdvice) > 0 {
		questionSummaries["d"] = "Growth advice: " + strings.Join(allAdvice, "; ")
	}

	// Create actionable insights from advice
	var actionableInsights []string
	for _, advice := range allAdvice {
		if len(advice) > 10 {
			actionableInsights = append(actionableInsights, advice)
		}
	}

	return models.Consolidation{
		RoundID:             roundID,
		GeneratedByID:       generatedByID,
		ExecutiveSummary:    executiveSummary,
		Strengths:           allStrengths,
		AreasForImprovement: allImprovements,
		ActionableInsights:  actionableInsights,
		QuestionSummaries:   questionSummaries,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}
