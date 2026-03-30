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
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Convert id string to ObjectID
	consolidationObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consolidation ID"})
		return
	}

	// Get consolidation to find round ID
	var consolidation models.Consolidation
	err = db.Collection("consolidations").FindOne(ctx, bson.M{"_id": consolidationObjID}).Decode(&consolidation)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consolidation not found"})
		return
	}

	// Get the round to validate status
	var round models.FeedbackRound
	err = db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": consolidation.RoundID}).Decode(&round)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Validate round is in Closed status
	if round.Status != models.RoundClosed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Cannot share consolidation. Round status is '%s', must be 'closed'.", round.Status),
		})
		return
	}

	// Get subject for audit log
	var subject models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": round.SubjectID}).Decode(&subject)

	// Update consolidation with shared timestamp
	update := bson.M{"$set": bson.M{
		"shared_at":  time.Now(),
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("consolidations").UpdateOne(ctx, bson.M{"_id": consolidationObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share consolidation"})
		return
	}

	// Update round status to shared
	_, err = db.Collection("feedback_rounds").UpdateOne(
		ctx,
		bson.M{"_id": consolidation.RoundID},
		bson.M{"$set": bson.M{
			"status":     models.RoundShared,
			"updated_at": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update round status"})
		return
	}

	// Create audit log
	createAuditLog(ctx, AuditLogParams{
		Action:       models.AuditConsolidationShared,
		ActorID:      currentUser.ID,
		ActorName:    currentUser.Name,
		ActorEmail:   currentUser.Email,
		RoundID:      consolidation.RoundID,
		RoundSubject: subject.Name,
		Description:  fmt.Sprintf("Shared consolidation with subject (closed → shared)"),
		OldValue:     string(models.RoundClosed),
		NewValue:     string(models.RoundShared),
	})

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
	prompt := fmt.Sprintf(`You are an expert and caring Engineering Team Lead proficient in 360-degree feedback analysis. 
Please analyze the following feedback from multiple reviewers and provide a comprehensive consolidation.

Feedback data:
%s

Please provide the analysis in the following JSON format:
{
  "executive_summary": "A concise 2-3 sentence summary of the feedback from multiple reviewers and provide a comprehensive consolidation. Using an Engineering Manager tone.",
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

Focus on being constructive, specific, and actionable.

Return ONLY a single minified JSON object. Do not include any code fences, markdown, or explanatory text.`, strings.Join(feedbackTexts, "\n\n"))

	// Initialize Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return models.Consolidation{}, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-flash-latest")
	// Request strict JSON output
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"executive_summary": &genai.Schema{Type: genai.TypeString},
			"strengths": &genai.Schema{
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"areas_for_improvement": &genai.Schema{
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"actionable_insights": &genai.Schema{
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"question_summaries": &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"a": &genai.Schema{Type: genai.TypeString},
					"b": &genai.Schema{Type: genai.TypeString},
					"c": &genai.Schema{Type: genai.TypeString},
					"d": &genai.Schema{Type: genai.TypeString},
				},
				Required: []string{"a", "b", "c", "d"},
			},
		},
		Required: []string{"executive_summary", "strengths", "areas_for_improvement", "actionable_insights", "question_summaries"},
	}

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return models.Consolidation{}, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return models.Consolidation{}, fmt.Errorf("no response from Gemini")
	}

	// Extract text response safely
	var responseText string
	firstPart := resp.Candidates[0].Content.Parts[0]
	if t, ok := firstPart.(genai.Text); ok {
		responseText = string(t)
	} else {
		var b strings.Builder
		for _, p := range resp.Candidates[0].Content.Parts {
			if tt, ok := p.(genai.Text); ok {
				b.WriteString(string(tt))
			}
		}
		responseText = b.String()
	}

	// Sanitize and extract JSON from the model response
	clean := strings.TrimSpace(responseText)
	// Strip common code-fence wrappers
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```JSON")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	// Extract the JSON object slice in case there’s prose around it
	if i := strings.IndexRune(clean, '{'); i != -1 {
		if j := strings.LastIndex(clean, "}"); j != -1 && j > i {
			clean = clean[i : j+1]
		}
	}

	// Parse JSON response
	var aiResponse struct {
		ExecutiveSummary    string            `json:"executive_summary"`
		Strengths           []string          `json:"strengths"`
		AreasForImprovement []string          `json:"areas_for_improvement"`
		ActionableInsights  []string          `json:"actionable_insights"`
		QuestionSummaries   map[string]string `json:"question_summaries"`
	}

	// Validate and unmarshal JSON
	if !json.Valid([]byte(clean)) {
		fmt.Printf("invalid JSON from Gemini. RAW AS %q\n", clean)
		aiResponse = struct {
			ExecutiveSummary    string            `json:"executive_summary"`
			Strengths           []string          `json:"strengths"`
			AreasForImprovement []string          `json:"areas_for_improvement"`
			ActionableInsights  []string          `json:"actionable_insights"`
			QuestionSummaries   map[string]string `json:"question_summaries"`
		}{
			ExecutiveSummary:    "..",
			Strengths:           []string{".."},
			AreasForImprovement: []string{".."},
			ActionableInsights:  []string{".."},
			QuestionSummaries: map[string]string{
				"a": "..",
				"b": "..",
				"c": "..",
				"d": "..",
			},
		}
	} else if err := json.Unmarshal([]byte(clean), &aiResponse); err != nil {
		fmt.Printf("unmarshal error: %v\nRAW AS %q\n", err, clean)
		aiResponse = struct {
			ExecutiveSummary    string            `json:"executive_summary"`
			Strengths           []string          `json:"strengths"`
			AreasForImprovement []string          `json:"areas_for_improvement"`
			ActionableInsights  []string          `json:"actionable_insights"`
			QuestionSummaries   map[string]string `json:"question_summaries"`
		}{
			ExecutiveSummary:    "..",
			Strengths:           []string{".."},
			AreasForImprovement: []string{".."},
			ActionableInsights:  []string{".."},
			QuestionSummaries: map[string]string{
				"a": "..",
				"b": "..",
				"c": "..",
				"d": "..",
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
