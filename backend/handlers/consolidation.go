package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smart360/database"
	"smart360/models"
	"smart360/repositories"
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

	// Look up the round so we can resolve its template. The template carries
	// the persona and per-question labels we want the consolidation to use.
	var round models.FeedbackRound
	if err := db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundObjID}).Decode(&round); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}
	template, err := resolveTemplate(ctx, repositories.NewMongoTemplateRepository(db), round.TemplateID)
	if err != nil {
		fmt.Printf("Error resolving template for round: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load round template"})
		return
	}

	// Run the dedicated moderation pass before the synthesis prompt sees the
	// content. Each submission is scrubbed individually so the audit log can
	// point precisely at which fields were rewritten and why. If GEMINI_API_KEY
	// is unset, this is a no-op and the originals continue to the fallback.
	moderatedSubmissions, moderationLogs := moderateSubmissions(ctx, submissions, roundObjID, geminiKey)
	persistModerationLogs(ctx, db, moderationLogs)
	submissions = moderatedSubmissions

	var consolidation models.Consolidation

	if hasGemini {
		// Use Gemini for AI consolidation
		consolidation, err = generateGeminiConsolidation(submissions, roundObjID, currentUser.ID, geminiKey, template)
		if err != nil {
			// Sanitise before logging — the genai SDK embeds the API key in URLs
			// inside its error strings.
			fmt.Printf("Error generating Gemini consolidation: %s\n", sanitiseErr(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate AI consolidation"})
			return
		}
	} else {
		// Combine actual feedback submissions
		fmt.Printf("Using real feedback combination\n")
		consolidation = combineFeedbackSubmissions(submissions, roundObjID, currentUser.ID)
	}

	// Snapshot the per-question display labels from the template onto the
	// consolidation so the UI doesn't have to do a separate round-trip. We do
	// this regardless of which generator path was taken.
	if consolidation.QuestionLabels == nil {
		consolidation.QuestionLabels = snapshotQuestionLabels(template)
	}
	// Likert aggregates are deterministic — compute them here regardless of
	// which generator ran. The AI prompt also saw the raw per-reviewer scores
	// so its synthesis can reference them qualitatively.
	if consolidation.CompetencyRatings == nil {
		consolidation.CompetencyRatings = aggregateCompetencyRatings(submissions, template)
	}

	fmt.Printf("Created consolidation: %+v\n", consolidation)

	res, err := db.Collection("consolidations").InsertOne(ctx, consolidation)
	if err != nil {
		fmt.Printf("Error inserting consolidation: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create consolidation"})
		return
	}
	// Write the inserted ObjectID back onto the struct so the response body
	// carries a real id (without this the client sees "000…00" and has to
	// re-fetch by roundId to do anything useful with the consolidation).
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		consolidation.ID = oid
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

	// Private manager-only channel is never surfaced to the subject. Only the
	// round creator and global admins see it. Mutate the in-memory copy before
	// serialisation — the persisted doc is unchanged.
	if !canSeeManagerOnlyChannel(currentUser, round) {
		consolidation.ManagerOnlyChannel = nil
	}

	c.JSON(http.StatusOK, consolidation)
}

// canSeeManagerOnlyChannel returns true when the caller is allowed to see the
// private manager-only synthesis. The subject of the round must not see it,
// even after the consolidation is shared; the round creator and global admins
// can. Team admins outside the round's creation lineage can't (they have no
// reason to read another manager's private channel).
func canSeeManagerOnlyChannel(user models.User, round models.FeedbackRound) bool {
	if user.Role == models.RoleAdmin {
		return true
	}
	return user.ID == round.CreatedByID
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

func generateGeminiConsolidation(submissions []models.Submission, roundID primitive.ObjectID, generatedByID primitive.ObjectID, apiKey string, template *models.Template) (models.Consolidation, error) {
	// Segregate the subject's self-assessment from peer feedback so the model
	// can compute the self-vs-others delta — the highest-signal output of a 360.
	peerTexts, selfText, hasSelf := buildFeedbackPrompts(submissions, template)
	privateNotes := collectPrivateNotes(submissions)

	selfSection := "No self-assessment submitted — return self_vs_others_delta with self_submitted=false and empty arrays."
	if hasSelf {
		selfSection = "Self-assessment from the subject:\n" + selfText
	}

	managerOnlySection := "No private manager-only notes were submitted — return manager_only_channel with note_count=0, empty synthesis, and empty themes."
	if len(privateNotes) > 0 {
		managerOnlySection = "Private notes from reviewers, addressed to the manager only (the subject will NEVER see these):\n" + strings.Join(privateNotes, "\n")
	}

	// Reviewer counts per voice are deterministic, so compute them server-side
	// rather than asking the model. The model produces the synthesis text only.
	mgrCount, peerCount, reportCount := countByVoice(submissions)
	voiceContext := fmt.Sprintf(
		"Voice reviewer counts: manager=%d, peer=%d (includes cross-functional), report=%d. "+
			"Produce a summary and themes for every voice; if a voice has zero reviewers, "+
			"return an empty summary string and an empty themes array for that voice.",
		mgrCount, peerCount, reportCount,
	)

	persona := "a thoughtful coach helping someone grow over the next 6 months"
	if template != nil && strings.TrimSpace(template.CoachingPersona) != "" {
		persona = strings.TrimSpace(template.CoachingPersona)
	}

	prompt := fmt.Sprintf(`You are %s.
You are reviewing 360 feedback from multiple peers AND a self-assessment from the subject, and synthesising it for them.

Apply these guidelines strictly:
- Use behavioural, observable language. Avoid trait or personality labels (do not say "they ARE …"). Prefer "they often DO X, which leads to Y".
- Use growth-oriented framing. Replace deficit language ("weakness", "bad at", "lacks") with forward-looking framing ("opportunity to amplify", "would unlock impact by", "next-level habit to build").
- Never reproduce direct quotes that could identify a specific reviewer. Synthesise across reviewers.
- Be specific. Vague compliments ("good communicator", "team player") are useless — ground every point in observable behaviour or impact.
- Anchor your analysis on this person's last 3–6 months.
- Content has already been scrubbed by a separate moderation pass; you should not see identity-targeted or personality-attack language. If you do, drop the offending content silently and proceed.

Weight reviewer signals by relationship and interaction frequency:
- Daily peers and the subject's manager have the richest signal — give their themes more weight, especially for execution and collaboration behaviours.
- Direct reports (subjects who manage them) carry distinct, high-value signal for leadership and feedback behaviours — weight them heavily for those themes.
- Cross-functional collaborators with rare interaction provide thin signal — only surface their themes if they appear in at least one other reviewer's input, otherwise treat them as a hypothesis rather than a finding.
- If a theme appears in only one rarely-interacting reviewer's input, frame it cautiously ("one cross-functional partner observed …") instead of stating it as fact.

Use the Likert ratings (when present) as quantitative anchors for your synthesis:
- A wide spread (e.g., one reviewer at 2 and another at 5 on the same competency) is itself a finding — surface it as a calibration gap to investigate.
- Don't restate the average numbers in the executive summary — the UI shows them. Instead, name the underlying *behaviours* the scores point at.

For the self-vs-others delta:
- blind_spots: things peers consistently flagged that the self-assessment does not acknowledge. Frame as opportunities, not accusations.
- hidden_strengths: things peers value highly that the self-assessment underplays or omits.
- aligned: themes where the self-assessment and peer feedback clearly agree.
- summary: 1–2 sentences in a coaching tone naming the most important gap and why closing it matters.
- If no self-assessment was submitted, set self_submitted=false, return empty arrays, and summary="".

For the voice_breakdown — separate views by vantage so distinct signals do not get averaged into mush:
- manager_voice: synthesise only the feedback from reviewers tagged "manager (manages the subject)". Lean into themes a manager is uniquely placed to see (scope, growth trajectory, readiness, judgement).
- peer_voice: synthesise feedback from reviewers tagged as peers OR cross-functional collaborators. Lean into themes about day-to-day collaboration, execution, and how they affect the people around them.
- report_voice: synthesise feedback from reviewers tagged "direct report (the subject manages them)". Lean into themes a report is uniquely placed to see (clarity, support, feedback they give, psychological safety).
- For each voice, summary is 1–2 sentences in coaching tone, and themes is at most 5 behaviourally-anchored bullets. Do not duplicate the top-level executive summary verbatim — each voice should add what's distinctive about that vantage.
%s

For manager_only_channel — the private notes reviewers addressed to the manager and NOT to the subject:
- note_count is the number of private notes received (we will overwrite this server-side; you can pass back whatever).
- synthesis: 1–2 sentences naming the pattern across the private notes, written in a coaching tone for the manager. Empty string if there are no private notes.
- themes: at most 5 short bullets distilling what reviewers want the manager to know privately. Empty array if no notes.
- This block is for the manager only — it will never be shown to the subject. Speak frankly; you don't need the same diplomatic framing as the subject-facing output.
%s

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
%s
  },
  "self_vs_others_delta": {
    "self_submitted": true,
    "blind_spots": ["Themes peers see that the self-assessment misses (at most 5)"],
    "hidden_strengths": ["Themes peers value that the self-assessment underplays (at most 5)"],
    "aligned": ["Themes where self and peers clearly agree (at most 5)"],
    "summary": "1–2 sentence coaching framing of the most important gap"
  },
  "voice_breakdown": {
    "manager_voice": {"summary": "1–2 sentences from the manager's vantage, or empty string if no manager reviewer", "themes": ["At most 5 themes, or empty array"]},
    "peer_voice":    {"summary": "1–2 sentences from peer/cross-functional vantage, or empty string if no peer reviewer", "themes": ["At most 5 themes, or empty array"]},
    "report_voice":  {"summary": "1–2 sentences from a direct report's vantage, or empty string if no report reviewer", "themes": ["At most 5 themes, or empty array"]}
  },
  "manager_only_channel": {
    "synthesis": "1–2 sentences for the manager only; empty string if no private notes were submitted",
    "themes": ["Frank, manager-only bullets (at most 5); empty array if no private notes"]
  }
}

Return ONLY the minified JSON object. No code fences, no markdown, no prose.`, persona, voiceContext, managerOnlySection, strings.Join(peerTexts, "\n\n"), selfSection, questionSummariesBlock(template))

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
			"voice_breakdown": &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"manager_voice": voiceSchema(),
					"peer_voice":    voiceSchema(),
					"report_voice":  voiceSchema(),
				},
				Required: []string{"manager_voice", "peer_voice", "report_voice"},
			},
			"manager_only_channel": &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"synthesis": &genai.Schema{Type: genai.TypeString},
					"themes": &genai.Schema{
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
					},
				},
				Required: []string{"synthesis", "themes"},
			},
		},
		Required: []string{"executive_summary", "strengths", "areas_for_improvement", "actionable_insights", "question_summaries", "self_vs_others_delta", "voice_breakdown", "manager_only_channel"},
	}

	// Generate content. Cap the call at 60s — Gemini occasionally hangs on
	// borderline prompts; without a deadline a single consolidate request
	// could block the HTTP handler indefinitely.
	genCtx, genCancel := context.WithTimeout(ctx, 60*time.Second)
	defer genCancel()
	resp, err := model.GenerateContent(genCtx, genai.Text(prompt))
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
		QuestionLabels:      snapshotQuestionLabels(template),
		SelfVsOthersDelta: &models.SelfVsOthersDelta{
			SelfSubmitted:   aiResponse.SelfVsOthersDelta.SelfSubmitted,
			BlindSpots:      aiResponse.SelfVsOthersDelta.BlindSpots,
			HiddenStrengths: aiResponse.SelfVsOthersDelta.HiddenStrengths,
			Aligned:         aiResponse.SelfVsOthersDelta.Aligned,
			Summary:         aiResponse.SelfVsOthersDelta.Summary,
		},
		VoiceBreakdown:     buildVoiceBreakdown(aiResponse.VoiceBreakdown, mgrCount, peerCount, reportCount),
		ManagerOnlyChannel: buildManagerOnlyChannel(aiResponse.ManagerOnlyChannel, privateNotes),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}, nil
}

// collectPrivateNotes returns each non-empty private note from peer
// submissions, tagged with the relationship label so the manager has some
// context without identifying who wrote what. Self submissions never carry
// private notes.
func collectPrivateNotes(submissions []models.Submission) []string {
	var out []string
	for _, s := range submissions {
		if s.IsSelf {
			continue
		}
		note := strings.TrimSpace(s.PrivateNotes)
		if note == "" {
			continue
		}
		tag := relationshipLabel(s.Relationship)
		out = append(out, fmt.Sprintf("[%s] %s", tag, note))
	}
	return out
}

// buildManagerOnlyChannel promotes the AI payload into the typed model,
// attaching the raw notes the AI saw and the server-known note count. Returns
// nil when there were no private notes (so the API/UI can skip the section).
func buildManagerOnlyChannel(p aiManagerOnlyPayload, rawNotes []string) *models.ManagerOnlyChannel {
	if len(rawNotes) == 0 {
		return nil
	}
	return &models.ManagerOnlyChannel{
		NoteCount: len(rawNotes),
		Synthesis: p.Synthesis,
		Themes:    p.Themes,
		RawNotes:  rawNotes,
	}
}

// snapshotQuestionLabels copies the template's CardTitle per question key into
// a flat map, ready to persist on the Consolidation. Returns nil for a nil
// template so callers can rely on the omitempty BSON tag.
func snapshotQuestionLabels(template *models.Template) map[string]string {
	if template == nil || len(template.Questions) == 0 {
		return nil
	}
	labels := make(map[string]string, len(template.Questions))
	for _, q := range template.Questions {
		labels[q.Key] = q.CardTitle
	}
	return labels
}

// countByVoice counts peer submissions per vantage. The "peer" voice covers
// both `peer` and `cross_functional` relationships — the distinction is for
// signal-weighting, not for voice separation.
func countByVoice(submissions []models.Submission) (manager, peer, report int) {
	for _, s := range submissions {
		if s.IsSelf {
			continue
		}
		switch s.Relationship {
		case models.RelationshipManager:
			manager++
		case models.RelationshipReport:
			report++
		case models.RelationshipPeer, models.RelationshipCrossFunctional:
			peer++
		}
	}
	return manager, peer, report
}

// buildVoiceBreakdown promotes the AI voice payload into the typed model
// struct, attaching the reviewer count we computed server-side and dropping
// any voice that had zero reviewers (preventing the model from making up text
// for a vantage that was never represented).
func buildVoiceBreakdown(p aiVoiceBreakdownPayload, mgrCount, peerCount, reportCount int) *models.VoiceBreakdown {
	if mgrCount == 0 && peerCount == 0 && reportCount == 0 {
		return nil
	}
	vb := &models.VoiceBreakdown{}
	if mgrCount > 0 {
		vb.ManagerVoice = &models.Voice{
			ReviewerCount: mgrCount,
			Summary:       p.ManagerVoice.Summary,
			Themes:        p.ManagerVoice.Themes,
		}
	}
	if peerCount > 0 {
		vb.PeerVoice = &models.Voice{
			ReviewerCount: peerCount,
			Summary:       p.PeerVoice.Summary,
			Themes:        p.PeerVoice.Themes,
		}
	}
	if reportCount > 0 {
		vb.ReportVoice = &models.Voice{
			ReviewerCount: reportCount,
			Summary:       p.ReportVoice.Summary,
			Themes:        p.ReportVoice.Themes,
		}
	}
	return vb
}

func voiceSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"summary": &genai.Schema{Type: genai.TypeString},
			"themes": &genai.Schema{
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
		},
		Required: []string{"summary", "themes"},
	}
}

// aiDeltaPayload mirrors the self_vs_others_delta block from the Gemini schema.
type aiDeltaPayload struct {
	SelfSubmitted   bool     `json:"self_submitted"`
	BlindSpots      []string `json:"blind_spots"`
	HiddenStrengths []string `json:"hidden_strengths"`
	Aligned         []string `json:"aligned"`
	Summary         string   `json:"summary"`
}

// aiVoicePayload mirrors a single voice block from the Gemini schema.
type aiVoicePayload struct {
	Summary string   `json:"summary"`
	Themes  []string `json:"themes"`
}

// aiVoiceBreakdownPayload mirrors the voice_breakdown block.
type aiVoiceBreakdownPayload struct {
	ManagerVoice aiVoicePayload `json:"manager_voice"`
	PeerVoice    aiVoicePayload `json:"peer_voice"`
	ReportVoice  aiVoicePayload `json:"report_voice"`
}

// aiManagerOnlyPayload mirrors the manager_only_channel block.
type aiManagerOnlyPayload struct {
	Synthesis string   `json:"synthesis"`
	Themes    []string `json:"themes"`
}

// aiPayload is the full JSON shape we expect from Gemini.
type aiPayload struct {
	ExecutiveSummary    string                  `json:"executive_summary"`
	Strengths           []string                `json:"strengths"`
	AreasForImprovement []string                `json:"areas_for_improvement"`
	ActionableInsights  []string                `json:"actionable_insights"`
	QuestionSummaries   map[string]string       `json:"question_summaries"`
	SelfVsOthersDelta   aiDeltaPayload          `json:"self_vs_others_delta"`
	VoiceBreakdown      aiVoiceBreakdownPayload `json:"voice_breakdown"`
	ManagerOnlyChannel  aiManagerOnlyPayload    `json:"manager_only_channel"`
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
	p.VoiceBreakdown.ManagerVoice.Themes = []string{}
	p.VoiceBreakdown.PeerVoice.Themes = []string{}
	p.VoiceBreakdown.ReportVoice.Themes = []string{}
	p.ManagerOnlyChannel.Themes = []string{}
	return p
}

// buildFeedbackPrompts splits submissions into peer feedback blocks and the
// subject's self-assessment block. Peer blocks are tagged with the reviewer's
// relationship and interaction frequency so the model can down-weight thin
// signals (a rare cross-functional contact) versus rich ones (a daily peer).
// The template (if non-nil) drives the human-readable label for each question
// key so the model knows what each line is about.
func buildFeedbackPrompts(submissions []models.Submission, template *models.Template) (peerTexts []string, selfText string, hasSelf bool) {
	labels := questionLabelsFromTemplate(template)
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
		if !submission.IsSelf {
			block += fmt.Sprintf("Relationship to subject: %s\n", relationshipLabel(submission.Relationship))
			block += fmt.Sprintf("Interaction frequency: %s\n", frequencyLabel(submission.InteractionFrequency))
		}
		// Iterate template questions in order so the prompt block reads
		// consistently across rounds even if a template defines a custom shape.
		if template != nil && len(template.Questions) > 0 {
			for _, q := range template.Questions {
				if ans, ok := responses[q.Key]; ok && ans != "" {
					block += fmt.Sprintf("%s: %s\n", labels[q.Key], ans)
				}
			}
		} else {
			for _, key := range []string{"a", "b", "c", "d"} {
				if ans, ok := responses[key]; ok && ans != "" {
					block += fmt.Sprintf("%s: %s\n", labels[key], ans)
				}
			}
		}

		if len(submission.Ratings) > 0 {
			block += "Ratings (1–5):\n"
			names := competencyNamesByKey(template)
			for _, r := range submission.Ratings {
				name := names[r.Key]
				if name == "" {
					name = r.Key
				}
				block += fmt.Sprintf("  • %s — %d. %s\n", name, r.Score, strings.TrimSpace(r.Justification))
			}
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

func competencyNamesByKey(template *models.Template) map[string]string {
	if template == nil {
		return nil
	}
	out := make(map[string]string, len(template.Competencies))
	for _, c := range template.Competencies {
		out[c.Key] = c.Name
	}
	return out
}

// questionLabelsFromTemplate maps each question key to a short label used to
// tag the line in the AI prompt. Falls back to the legacy a/b/c/d wording so
// rounds without a template still produce readable prompts in tests.
func questionLabelsFromTemplate(template *models.Template) map[string]string {
	labels := map[string]string{
		"a": "What to continue (biggest positive impact, with example)",
		"b": "What's blocking growth (last 3–6 months)",
		"c": "Where to double down (one strength to amplify)",
		"d": "Suggested experiment (next 30–60 days)",
	}
	if template == nil {
		return labels
	}
	for _, q := range template.Questions {
		if q.CardTitle != "" {
			labels[q.Key] = q.CardTitle
		}
	}
	return labels
}

// questionSummariesBlock emits the per-question instruction lines for the
// "question_summaries" object in the prompt's expected JSON shape. Each line
// reads `    "a": "Synthesis across reviewers about: <CardTitle>",`.
func questionSummariesBlock(template *models.Template) string {
	keys := []string{"a", "b", "c", "d"}
	if template != nil && len(template.Questions) > 0 {
		keys = nil
		for _, q := range template.Questions {
			keys = append(keys, q.Key)
		}
	}
	labels := questionLabelsFromTemplate(template)

	var lines []string
	for i, key := range keys {
		comma := ","
		if i == len(keys)-1 {
			comma = ""
		}
		lines = append(lines, fmt.Sprintf(`    "%s": "Synthesis across reviewers about: %s"%s`, key, labels[key], comma))
	}
	return strings.Join(lines, "\n")
}

func relationshipLabel(r models.ReviewerRelationship) string {
	switch r {
	case models.RelationshipManager:
		return "manager (manages the subject)"
	case models.RelationshipReport:
		return "direct report (the subject manages them)"
	case models.RelationshipPeer:
		return "peer (direct teammate)"
	case models.RelationshipCrossFunctional:
		return "cross-functional collaborator (different team)"
	}
	return "unspecified"
}

func frequencyLabel(f models.InteractionFrequency) string {
	switch f {
	case models.InteractionDaily:
		return "daily — works together most days"
	case models.InteractionWeekly:
		return "weekly — syncs at least once a week"
	case models.InteractionMonthly:
		return "monthly — connects occasionally"
	case models.InteractionRarely:
		return "rarely — limited direct interaction"
	}
	return "unspecified"
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

	// Per-voice aggregates feed the fallback VoiceBreakdown when no LLM key is
	// configured. Keeping it lightweight: we just stitch the raw "continue"
	// answers as themes so the manager has something to look at.
	type voiceAcc struct {
		count    int
		themes   []string
		blockers []string
	}
	mgrAcc := voiceAcc{}
	peerAcc := voiceAcc{}
	reportAcc := voiceAcc{}

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

		var acc *voiceAcc
		switch submission.Relationship {
		case models.RelationshipManager:
			acc = &mgrAcc
		case models.RelationshipReport:
			acc = &reportAcc
		case models.RelationshipPeer, models.RelationshipCrossFunctional:
			acc = &peerAcc
		}
		if acc != nil {
			acc.count++
			if t := responses["a"]; t != "" {
				acc.themes = append(acc.themes, t)
			}
			if t := responses["b"]; t != "" {
				acc.blockers = append(acc.blockers, t)
			}
		}

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

	toVoice := func(label string, acc voiceAcc) *models.Voice {
		if acc.count == 0 {
			return nil
		}
		v := &models.Voice{
			ReviewerCount: acc.count,
			Summary:       fmt.Sprintf("%s feedback captured from %d reviewer(s). AI synthesis unavailable without GEMINI_API_KEY — themes below are raw answers, not summaries.", label, acc.count),
			Themes:        []string{},
		}
		if len(acc.themes) > 0 {
			v.Themes = append(v.Themes, "What to continue: "+strings.Join(acc.themes, "; "))
		}
		if len(acc.blockers) > 0 {
			v.Themes = append(v.Themes, "What's blocking growth: "+strings.Join(acc.blockers, "; "))
		}
		return v
	}

	var voiceBreakdown *models.VoiceBreakdown
	if mgrAcc.count+peerAcc.count+reportAcc.count > 0 {
		voiceBreakdown = &models.VoiceBreakdown{
			ManagerVoice: toVoice("Manager", mgrAcc),
			PeerVoice:    toVoice("Peer", peerAcc),
			ReportVoice:  toVoice("Direct report", reportAcc),
		}
	}

	// Without an LLM we can't synthesise the private notes; surface the raw,
	// relationship-tagged texts so the manager can read them directly.
	var managerOnly *models.ManagerOnlyChannel
	if rawNotes := collectPrivateNotes(submissions); len(rawNotes) > 0 {
		managerOnly = &models.ManagerOnlyChannel{
			NoteCount: len(rawNotes),
			Synthesis: "AI synthesis unavailable without GEMINI_API_KEY. Read the raw private notes below.",
			Themes:    []string{},
			RawNotes:  rawNotes,
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
		VoiceBreakdown:      voiceBreakdown,
		ManagerOnlyChannel:  managerOnly,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}
