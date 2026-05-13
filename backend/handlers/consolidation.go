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
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	var round models.FeedbackRound
	if err := db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Authorization mirrors DownloadConsolidationPDF: admin, round creator, or
	// the round subject (only after the consolidation has been shared).
	if currentUser.Role != models.RoleAdmin &&
		currentUser.ID != round.SubjectID &&
		currentUser.ID != round.CreatedByID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to access this consolidation"})
		return
	}

	var consolidation models.Consolidation
	err = db.Collection("consolidations").FindOne(ctx, bson.M{"round_id": roundObjID}).Decode(&consolidation)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consolidation not found"})
		return
	}

	if currentUser.Role != models.RoleAdmin && currentUser.ID == round.SubjectID && consolidation.SharedAt == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Consolidation has not been shared yet"})
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
	// Segregate the subject's self-assessment from peer feedback so the model
	// can compute the self-vs-others delta — the highest-signal output of a 360.
	peerTexts, selfText, hasSelf := buildFeedbackPrompts(submissions)

	selfSection := "No self-assessment submitted — return self_vs_others_delta with self_submitted=false and empty arrays."
	if hasSelf {
		selfSection = "Self-assessment from the subject:\n" + selfText
	}

	prompt := fmt.Sprintf(`You are a thoughtful coach helping someone grow over the next 6 months.
You are reviewing 360 feedback from multiple peers AND a self-assessment from the subject, and synthesising it for them.

Apply these guidelines strictly:
- Use behavioural, observable language. Avoid trait or personality labels (do not say "they ARE …"). Prefer "they often DO X, which leads to Y".
- Use growth-oriented framing. Replace deficit language ("weakness", "bad at", "lacks") with forward-looking framing ("opportunity to amplify", "would unlock impact by", "next-level habit to build").
- Never reproduce direct quotes that could identify a specific reviewer. Synthesise across reviewers.
- If any reviewer input contains personal attacks, identity-targeted comments, or content unrelated to professional behaviour, omit it entirely from the consolidation. Do not surface it to the subject and do not reference that it was filtered.
- Be specific. Vague compliments ("good communicator", "team player") are useless — ground every point in observable behaviour or impact.
- Anchor your analysis on this person's last 3–6 months.

For the self-vs-others delta:
- blind_spots: things peers consistently flagged that the self-assessment does not acknowledge. Frame as opportunities, not accusations.
- hidden_strengths: things peers value highly that the self-assessment underplays or omits.
- aligned: themes where the self-assessment and peer feedback clearly agree.
- summary: 1–2 sentences in a coaching tone naming the most important gap and why closing it matters.
- If no self-assessment was submitted, set self_submitted=false, return empty arrays, and summary="".

Feedback data from peer reviewers:
%s

%s

Return a single minified JSON object with this exact shape:
{
  "executive_summary": "2–3 sentences in a coaching tone. Forward-looking. No verdicts.",
  "strengths": ["Behaviourally anchored strengths (at most 5)"],
  "areas_for_improvement": ["Growth-oriented opportunities, NOT deficits (at most 5)"],
  "actionable_insights": ["The TOP 1–3 focus areas the subject should act on first. Never more than 3. Each should be a concrete next step, not advice in the abstract."],
  "question_summaries": {
    "a": "Synthesis across reviewers of what this person should continue doing",
    "b": "Synthesis across reviewers of what's blocking their next level of impact",
    "c": "Synthesis across reviewers of the one strength to double down on",
    "d": "Synthesis across reviewers of concrete experiments for the next 30–60 days"
  },
  "self_vs_others_delta": {
    "self_submitted": true,
    "blind_spots": ["Themes peers see that the self-assessment misses (at most 5)"],
    "hidden_strengths": ["Themes peers value that the self-assessment underplays (at most 5)"],
    "aligned": ["Themes where self and peers clearly agree (at most 5)"],
    "summary": "1–2 sentence coaching framing of the most important gap"
  }
}

Return ONLY the minified JSON object. No code fences, no markdown, no prose.`, strings.Join(peerTexts, "\n\n"), selfSection)

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
			"self_vs_others_delta": &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"self_submitted": &genai.Schema{Type: genai.TypeBoolean},
					"blind_spots": &genai.Schema{
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
					},
					"hidden_strengths": &genai.Schema{
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
					},
					"aligned": &genai.Schema{
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
					},
					"summary": &genai.Schema{Type: genai.TypeString},
				},
				Required: []string{"self_submitted", "blind_spots", "hidden_strengths", "aligned", "summary"},
			},
		},
		Required: []string{"executive_summary", "strengths", "areas_for_improvement", "actionable_insights", "question_summaries", "self_vs_others_delta"},
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

	aiResponse := aiPayload{}
	if !json.Valid([]byte(clean)) {
		fmt.Printf("invalid JSON from Gemini. RAW AS %q\n", clean)
		aiResponse = fallbackAIPayload(hasSelf)
	} else if err := json.Unmarshal([]byte(clean), &aiResponse); err != nil {
		fmt.Printf("unmarshal error: %v\nRAW AS %q\n", err, clean)
		aiResponse = fallbackAIPayload(hasSelf)
	}

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
		SelfVsOthersDelta: &models.SelfVsOthersDelta{
			SelfSubmitted:   aiResponse.SelfVsOthersDelta.SelfSubmitted,
			BlindSpots:      aiResponse.SelfVsOthersDelta.BlindSpots,
			HiddenStrengths: aiResponse.SelfVsOthersDelta.HiddenStrengths,
			Aligned:         aiResponse.SelfVsOthersDelta.Aligned,
			Summary:         aiResponse.SelfVsOthersDelta.Summary,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// aiDeltaPayload mirrors the self_vs_others_delta block from the Gemini schema.
type aiDeltaPayload struct {
	SelfSubmitted   bool     `json:"self_submitted"`
	BlindSpots      []string `json:"blind_spots"`
	HiddenStrengths []string `json:"hidden_strengths"`
	Aligned         []string `json:"aligned"`
	Summary         string   `json:"summary"`
}

// aiPayload is the full JSON shape we expect from Gemini.
type aiPayload struct {
	ExecutiveSummary    string            `json:"executive_summary"`
	Strengths           []string          `json:"strengths"`
	AreasForImprovement []string          `json:"areas_for_improvement"`
	ActionableInsights  []string          `json:"actionable_insights"`
	QuestionSummaries   map[string]string `json:"question_summaries"`
	SelfVsOthersDelta   aiDeltaPayload    `json:"self_vs_others_delta"`
}

func fallbackAIPayload(hasSelf bool) aiPayload {
	p := aiPayload{
		ExecutiveSummary:    "Consolidation could not be generated automatically. Please review the raw submissions.",
		Strengths:           []string{},
		AreasForImprovement: []string{},
		ActionableInsights:  []string{},
		QuestionSummaries: map[string]string{
			"a": "",
			"b": "",
			"c": "",
			"d": "",
		},
	}
	p.SelfVsOthersDelta.SelfSubmitted = hasSelf
	p.SelfVsOthersDelta.BlindSpots = []string{}
	p.SelfVsOthersDelta.HiddenStrengths = []string{}
	p.SelfVsOthersDelta.Aligned = []string{}
	return p
}

// buildFeedbackPrompts splits submissions into peer feedback blocks and the
// subject's self-assessment block.
func buildFeedbackPrompts(submissions []models.Submission) (peerTexts []string, selfText string, hasSelf bool) {
	for _, submission := range submissions {
		var responses map[string]string
		if err := json.Unmarshal([]byte(submission.Responses), &responses); err != nil {
			continue
		}

		header := "Feedback from peer reviewer:"
		if submission.IsSelf {
			header = "Self-assessment from the subject:"
		}
		block := header + "\n"
		if continueDoing, ok := responses["a"]; ok && continueDoing != "" {
			block += fmt.Sprintf("What to continue (biggest positive impact, with example): %s\n", continueDoing)
		}
		if blockers, ok := responses["b"]; ok && blockers != "" {
			block += fmt.Sprintf("What's blocking growth (last 3–6 months): %s\n", blockers)
		}
		if amplify, ok := responses["c"]; ok && amplify != "" {
			block += fmt.Sprintf("Where to double down (one strength to amplify): %s\n", amplify)
		}
		if experiment, ok := responses["d"]; ok && experiment != "" {
			block += fmt.Sprintf("Suggested experiment (next 30–60 days): %s\n", experiment)
		}

		if submission.IsSelf {
			selfText = block
			hasSelf = true
		} else {
			peerTexts = append(peerTexts, block)
		}
	}
	return peerTexts, selfText, hasSelf
}

func combineFeedbackSubmissions(submissions []models.Submission, roundID primitive.ObjectID, generatedByID primitive.ObjectID) models.Consolidation {
	var allContinue []string
	var allBlockers []string
	var allAmplify []string
	var allExperiments []string
	var selfResponses map[string]string
	hasSelf := false
	peerCount := 0
	questionSummaries := make(map[string]string)

	for _, submission := range submissions {
		var responses map[string]string
		if err := json.Unmarshal([]byte(submission.Responses), &responses); err != nil {
			continue
		}

		if submission.IsSelf {
			selfResponses = responses
			hasSelf = true
			continue
		}
		peerCount++

		if continueDoing, ok := responses["a"]; ok && continueDoing != "" {
			allContinue = append(allContinue, continueDoing)
		}
		if blockers, ok := responses["b"]; ok && blockers != "" {
			allBlockers = append(allBlockers, blockers)
		}
		if amplify, ok := responses["c"]; ok && amplify != "" {
			allAmplify = append(allAmplify, amplify)
		}
		if experiment, ok := responses["d"]; ok && experiment != "" {
			allExperiments = append(allExperiments, experiment)
		}
	}

	executiveSummary := fmt.Sprintf("Consolidated feedback from %d peer reviewers. ", peerCount)
	if hasSelf {
		executiveSummary += "Includes a self-assessment from the subject for delta analysis. "
	}
	if len(allContinue) > 0 {
		executiveSummary += "Reviewers called out concrete behaviours worth continuing. "
	}
	if len(allBlockers) > 0 || len(allExperiments) > 0 {
		executiveSummary += "There are clear opportunities to unlock the next level of impact over the coming months."
	}

	// Create question summaries
	if len(allContinue) > 0 {
		questionSummaries["a"] = "What to continue: " + strings.Join(allContinue, "; ")
	}
	if len(allBlockers) > 0 {
		questionSummaries["b"] = "What's blocking growth: " + strings.Join(allBlockers, "; ")
	}
	if len(allAmplify) > 0 {
		questionSummaries["c"] = "Where to double down: " + strings.Join(allAmplify, "; ")
	}
	if len(allExperiments) > 0 {
		questionSummaries["d"] = "Suggested experiments (next 30–60 days): " + strings.Join(allExperiments, "; ")
	}

	// Top focus areas (cap at 3) drawn from the suggested experiments.
	var actionableInsights []string
	for _, experiment := range allExperiments {
		if len(actionableInsights) >= 3 {
			break
		}
		if len(experiment) > 10 {
			actionableInsights = append(actionableInsights, experiment)
		}
	}

	// Convert arrays and objects back to JSON strings for database compatibility
	strengthsJSON, _ := json.Marshal(allContinue)
	improvementsJSON, _ := json.Marshal(allBlockers)
	insightsJSON, _ := json.Marshal(actionableInsights)
	questionsJSON, _ := json.Marshal(questionSummaries)

	// Without an LLM we can't infer semantic delta; surface the raw self answers
	// alongside a flag so the UI can prompt the manager to interpret them.
	delta := &models.SelfVsOthersDelta{
		SelfSubmitted:   hasSelf,
		BlindSpots:      []string{},
		HiddenStrengths: []string{},
		Aligned:         []string{},
	}
	if hasSelf {
		delta.Summary = "Self-assessment captured — AI-assisted delta unavailable without GEMINI_API_KEY. Compare manually."
		if selfContinue := selfResponses["a"]; selfContinue != "" {
			delta.Aligned = append(delta.Aligned, "Self — what to continue: "+selfContinue)
		}
		if selfBlockers := selfResponses["b"]; selfBlockers != "" {
			delta.BlindSpots = append(delta.BlindSpots, "Self — what's blocking growth: "+selfBlockers)
		}
		if selfAmplify := selfResponses["c"]; selfAmplify != "" {
			delta.HiddenStrengths = append(delta.HiddenStrengths, "Self — where to double down: "+selfAmplify)
		}
	}

	return models.Consolidation{
		RoundID:             roundID,
		GeneratedByID:       generatedByID,
		ExecutiveSummary:    executiveSummary,
		Strengths:           string(strengthsJSON),
		AreasForImprovement: string(improvementsJSON),
		ActionableInsights:  string(insightsJSON),
		QuestionSummaries:   string(questionsJSON),
		SelfVsOthersDelta:   delta,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}
