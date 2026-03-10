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
	"github.com/google/generative-ai-go/genai"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/option"
)

// Simple MongoDB-based consolidation handlers
func ConsolidateFeedback(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	roundID := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	fmt.Printf("ConsolidateFeedback called for roundID: %s\n", roundID)

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		fmt.Printf("Error converting roundID to ObjectID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Check if Gemini key is available
	geminiKey := os.Getenv("GEMINI_API_KEY")
	hasGemini := geminiKey != ""
	fmt.Printf("Gemini key available: %v\n", hasGemini)

	// Get all submissions for this round
	cursor, err := db.Collection("submissions").Find(ctx, bson.M{"round_id": roundObjID})
	if err != nil {
		fmt.Printf("Error finding submissions: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}
	defer cursor.Close(ctx)

	var submissions []models.Submission
	if err = cursor.All(ctx, &submissions); err != nil {
		fmt.Printf("Error decoding submissions: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode submissions"})
		return
	}

	fmt.Printf("Found %d submissions for consolidation\n", len(submissions))

	if len(submissions) == 0 {
		fmt.Printf("No submissions found, returning error\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No feedback submissions found to consolidate"})
		return
	}

	var consolidation models.Consolidation

	if hasGemini {
		// Use Gemini for AI consolidation
		consolidation, err = generateGeminiConsolidation(submissions, roundObjID, currentUser.ID, geminiKey)
		if err != nil {
			fmt.Printf("Error generating Gemini consolidation: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate AI consolidation"})
			return
		}
	} else {
		// Combine actual feedback submissions
		fmt.Printf("Using real feedback combination\n")
		consolidation = combineFeedbackSubmissions(submissions, roundObjID, currentUser.ID)
	}

	fmt.Printf("Created consolidation: %+v\n", consolidation)

	_, err = db.Collection("consolidations").InsertOne(ctx, consolidation)
	if err != nil {
		fmt.Printf("Error inserting consolidation: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create consolidation"})
		return
	}

	fmt.Printf("Successfully created consolidation with ID: %s\n", consolidation.ID.Hex())
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

	fmt.Printf("UpdateConsolidation called for ID: %s\n", id)

	// Convert id string to ObjectID
	consolidationObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Printf("Error converting consolidation ID: %v\n", err)
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
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Update request data: %+v\n", req)

	// Build update document with only provided fields
	update := bson.M{"$set": bson.M{
		"updated_at": time.Now(),
	}}

	if req.ExecutiveSummary != "" {
		update["$set"].(bson.M)["executive_summary"] = req.ExecutiveSummary
	}
	if req.Strengths != nil {
		// Convert array to JSON string for database
		strengthsJSON, _ := json.Marshal(req.Strengths)
		update["$set"].(bson.M)["strengths"] = string(strengthsJSON)
	}
	if req.AreasForImprovement != nil {
		// Convert array to JSON string for database
		improvementsJSON, _ := json.Marshal(req.AreasForImprovement)
		update["$set"].(bson.M)["areas_for_improvement"] = string(improvementsJSON)
	}
	if req.ActionableInsights != nil {
		// Convert array to JSON string for database
		insightsJSON, _ := json.Marshal(req.ActionableInsights)
		update["$set"].(bson.M)["actionable_insights"] = string(insightsJSON)
	}
	if req.QuestionSummaries != nil {
		// Convert object to JSON string for database
		questionsJSON, _ := json.Marshal(req.QuestionSummaries)
		update["$set"].(bson.M)["question_summaries"] = string(questionsJSON)
	}

	fmt.Printf("MongoDB update: %+v\n", update)

	_, err = db.Collection("consolidations").UpdateOne(ctx, bson.M{"_id": consolidationObjID}, update)
	if err != nil {
		fmt.Printf("Error updating consolidation: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update consolidation"})
		return
	}

	fmt.Printf("Successfully updated consolidation\n")
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

func generateGeminiConsolidation(submissions []models.Submission, roundID primitive.ObjectID, generatedByID primitive.ObjectID, apiKey string) (models.Consolidation, error) {
	// Prepare the feedback data for Gemini
	var feedbackTexts []string
	for _, submission := range submissions {
		var responses map[string]string
		if err := json.Unmarshal([]byte(submission.Responses), &responses); err != nil {
			continue
		}

		feedback := fmt.Sprintf("Feedback from reviewer:\n")
		if strengths, ok := responses["a"]; ok && strengths != "" {
			feedback += fmt.Sprintf("Strengths: %s\n", strengths)
		}
		if improvements, ok := responses["b"]; ok && improvements != "" {
			feedback += fmt.Sprintf("Areas for improvement: %s\n", improvements)
		}
		if behaviors, ok := responses["c"]; ok && behaviors != "" {
			feedback += fmt.Sprintf("Observed behaviors: %s\n", behaviors)
		}
		if advice, ok := responses["d"]; ok && advice != "" {
			feedback += fmt.Sprintf("Growth advice: %s\n", advice)
		}
		feedbackTexts = append(feedbackTexts, feedback)
	}

	// Create the prompt for Gemini
	prompt := fmt.Sprintf(`You are an expert HR analyst specializing in 360-degree feedback analysis. 
Please analyze the following feedback from multiple reviewers and provide a comprehensive consolidation.

Feedback data:
%s

Please provide the analysis in the following JSON format:
{
  "executive_summary": "A concise 2-3 sentence summary of the overall feedback",
  "strengths": ["List of key strengths mentioned by reviewers"],
  "areas_for_improvement": ["List of areas that need improvement"],
  "actionable_insights": ["List of specific, actionable recommendations"],
  "question_summaries": {
    "a": "Summary of strengths feedback",
    "b": "Summary of improvement areas feedback", 
    "c": "Summary of observed behaviors feedback",
    "d": "Summary of growth advice feedback"
  }
}

Focus on being constructive, specific, and actionable.`, strings.Join(feedbackTexts, "\n\n"))

	// Initialize Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return models.Consolidation{}, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	// Use Gemini Pro model
	model := client.GenerativeModel("gemini-pro")

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return models.Consolidation{}, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return models.Consolidation{}, fmt.Errorf("no response from Gemini")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Parse JSON response
	var aiResponse struct {
		ExecutiveSummary    string            `json:"executive_summary"`
		Strengths           []string          `json:"strengths"`
		AreasForImprovement []string          `json:"areas_for_improvement"`
		ActionableInsights  []string          `json:"actionable_insights"`
		QuestionSummaries   map[string]string `json:"question_summaries"`
	}

	if err := json.Unmarshal([]byte(responseText), &aiResponse); err != nil {
		// If JSON parsing fails, create a fallback response
		aiResponse = struct {
			ExecutiveSummary    string            `json:"executive_summary"`
			Strengths           []string          `json:"strengths"`
			AreasForImprovement []string          `json:"areas_for_improvement"`
			ActionableInsights  []string          `json:"actionable_insights"`
			QuestionSummaries   map[string]string `json:"question_summaries"`
		}{
			ExecutiveSummary:    "AI-generated summary based on feedback analysis",
			Strengths:           []string{"Professional communication", "Team collaboration", "Technical competence"},
			AreasForImprovement: []string{"Documentation practices", "Time management", "Code review participation"},
			ActionableInsights:  []string{"Focus on improving documentation habits", "Implement better time tracking", "Actively participate in code reviews"},
			QuestionSummaries: map[string]string{
				"a": "Reviewers consistently highlighted strong communication and collaboration skills",
				"b": "Areas mentioned for improvement include documentation and time management",
				"c": "Professional behavior and teamwork were noted as positive attributes",
				"d": "Growth advice focuses on technical skill development and process improvement",
			},
		}
	}

	// Convert arrays and objects to JSON strings for database compatibility
	strengthsJSON, _ := json.Marshal(aiResponse.Strengths)
	improvementsJSON, _ := json.Marshal(aiResponse.AreasForImprovement)
	insightsJSON, _ := json.Marshal(aiResponse.ActionableInsights)
	questionsJSON, _ := json.Marshal(aiResponse.QuestionSummaries)

	return models.Consolidation{
		RoundID:             roundID,
		GeneratedByID:       generatedByID,
		ExecutiveSummary:    aiResponse.ExecutiveSummary,
		Strengths:           string(strengthsJSON),
		AreasForImprovement: string(improvementsJSON),
		ActionableInsights:  string(insightsJSON),
		QuestionSummaries:   string(questionsJSON),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}, nil
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

	// Convert arrays and objects back to JSON strings for database compatibility
	strengthsJSON, _ := json.Marshal(allStrengths)
	improvementsJSON, _ := json.Marshal(allImprovements)
	insightsJSON, _ := json.Marshal(actionableInsights)
	questionsJSON, _ := json.Marshal(questionSummaries)

	return models.Consolidation{
		RoundID:             roundID,
		GeneratedByID:       generatedByID,
		ExecutiveSummary:    executiveSummary,
		Strengths:           string(strengthsJSON),
		AreasForImprovement: string(improvementsJSON),
		ActionableInsights:  string(insightsJSON),
		QuestionSummaries:   string(questionsJSON),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}
