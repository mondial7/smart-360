package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"smart360/database"
	"smart360/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/option"
)

// moderationModelName is the Gemini model used for the scrub pass. Kept as a
// const so the value lands in audit logs unambiguously.
const moderationModelName = "gemini-flash-latest"

// GetModerationLogsForRound returns the audit trail of moderation passes for
// a round. Admin / round creator only — the subject never sees these.
func GetModerationLogsForRound(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	// Same pattern as other round-scoped handlers in this package: the path
	// uses :id, the handler looks it up.
	roundID := c.Param("id")
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
	if currentUser.Role != models.RoleAdmin && currentUser.ID != round.CreatedByID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view moderation logs for this round"})
		return
	}

	cur, err := db.Collection("moderation_logs").Find(ctx, bson.M{"round_id": roundObjID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch moderation logs"})
		return
	}
	defer cur.Close(ctx)

	var logs []models.ModerationLog
	if err := cur.All(ctx, &logs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode moderation logs"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// moderationPrompt is the strict, audit-friendly prompt for the scrub pass.
// It's intentionally narrow: do not synthesise, do not add, only scrub.
const moderationPrompt = `You are a strict moderation assistant for 360-degree feedback. Your only job is to scrub one reviewer's content before it is forwarded to the consolidation step.

Apply these rules to every input field:
- Remove personality or trait labels ("they are toxic", "she's lazy", "he's just difficult"). Keep observable behaviour described neutrally.
- Remove identity-targeted comments (about race, gender, religion, age, nationality, sexuality, disability). If an identity-tinged comment is actually describing an observable workplace pattern, rewrite it in behavioural terms instead.
- Remove direct quotes or specific incidents that would identify the reviewer (private 1:1 conversations, dated incidents only the reviewer would know).
- Remove off-topic content (politics, personal gossip unrelated to professional behaviour).
- Do NOT add new content, opinions, or speculation. Only scrub.
- Preserve the original tone and structure as much as possible — short, sharp feedback stays short and sharp after scrubbing.
- If a field is entirely fine, return it unchanged.

Input is a JSON object whose keys are field identifiers and values are the reviewer's text. Output a single minified JSON object with exactly this shape:
{
  "cleaned": {"<same keys as input>": "<scrubbed text for that key>"},
  "flagged": <boolean — true if you removed or rewrote anything in any field>,
  "reasons": ["short bullets — one per rule applied, naming WHICH field was scrubbed and WHY"]
}

Return ONLY the minified JSON object. No code fences, no markdown, no prose.

Input:
%s`

// moderationResult is what the model returns. The reasons array is what we
// persist in the audit log; the cleaned map is what the consolidation prompt
// will then see.
type moderationResult struct {
	Cleaned map[string]string `json:"cleaned"`
	Flagged bool              `json:"flagged"`
	Reasons []string          `json:"reasons"`
}

// moderateSubmission runs the scrub pass on a single submission and returns
// (cleaned submission, audit log). If the model call fails, the original
// submission is returned unchanged and the log records the failure so the
// admin can see something went wrong without blocking the round entirely.
func moderateSubmission(ctx context.Context, client *genai.Client, submission models.Submission, roundID primitive.ObjectID) (models.Submission, models.ModerationLog) {
	log := models.ModerationLog{
		RoundID:      roundID,
		SubmissionID: submission.ID,
		Model:        moderationModelName,
		ModeratedAt:  time.Now(),
	}

	// Build the JSON-input the model sees. Keys are stable so we can map the
	// scrubbed values back. We include free-text question responses, rating
	// justifications, and private notes — anything a reviewer can type.
	input := map[string]string{}
	var responses map[string]string
	if err := json.Unmarshal([]byte(submission.Responses), &responses); err == nil {
		for k, v := range responses {
			if strings.TrimSpace(v) != "" {
				input["response_"+k] = v
			}
		}
	}
	for i, r := range submission.Ratings {
		if strings.TrimSpace(r.Justification) != "" {
			input[fmt.Sprintf("rating_%d_justification", i)] = r.Justification
		}
	}
	if strings.TrimSpace(submission.PrivateNotes) != "" {
		input["private_notes"] = submission.PrivateNotes
	}

	if len(input) == 0 {
		// Nothing for the model to look at — no point spending a call.
		return submission, log
	}

	inputJSON, _ := json.Marshal(input)
	prompt := fmt.Sprintf(moderationPrompt, string(inputJSON))

	model := client.GenerativeModel(moderationModelName)
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"cleaned": &genai.Schema{Type: genai.TypeObject},
			"flagged": &genai.Schema{Type: genai.TypeBoolean},
			"reasons": &genai.Schema{
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
		},
		Required: []string{"cleaned", "flagged", "reasons"},
	}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil || len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Reasons = []string{fmt.Sprintf("moderation call failed: %v — original content was forwarded unchanged", err)}
		return submission, log
	}

	var raw string
	if t, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		raw = strings.TrimSpace(string(t))
	}

	result := moderationResult{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		log.Reasons = []string{fmt.Sprintf("moderation response was not valid JSON: %v — original content was forwarded unchanged", err)}
		return submission, log
	}

	scrubbed, fields := applyModerationCleaned(&submission, result.Cleaned)
	log.Flagged = result.Flagged || len(fields) > 0
	log.Reasons = result.Reasons
	log.FieldsScrubbed = fields
	return scrubbed, log
}

// applyModerationCleaned writes the model's cleaned values back into a copy of
// the submission. Returns the modified submission and the list of fields that
// actually changed (so the audit log records which fields were scrubbed
// independently of the model's own self-reported `flagged` flag).
func applyModerationCleaned(submission *models.Submission, cleaned map[string]string) (models.Submission, []string) {
	out := *submission
	var changed []string

	// Responses are stored as a JSON-encoded string; mutate the parsed map.
	var responses map[string]string
	if err := json.Unmarshal([]byte(out.Responses), &responses); err == nil {
		for key := range responses {
			if v, ok := cleaned["response_"+key]; ok && v != responses[key] {
				responses[key] = v
				changed = append(changed, "response_"+key)
			}
		}
		if b, err := json.Marshal(responses); err == nil {
			out.Responses = string(b)
		}
	}

	// Ratings is a slice — mutate a copy.
	if len(out.Ratings) > 0 {
		newRatings := make([]models.CompetencyRating, len(out.Ratings))
		copy(newRatings, out.Ratings)
		for i := range newRatings {
			key := fmt.Sprintf("rating_%d_justification", i)
			if v, ok := cleaned[key]; ok && v != newRatings[i].Justification {
				newRatings[i].Justification = v
				changed = append(changed, key)
			}
		}
		out.Ratings = newRatings
	}

	if v, ok := cleaned["private_notes"]; ok && v != out.PrivateNotes {
		out.PrivateNotes = v
		changed = append(changed, "private_notes")
	}

	return out, changed
}

// persistModerationLogs writes the audit entries to the moderation_logs
// collection. Failures are logged to stdout but don't block the consolidation
// — the audit record is best-effort and never worth failing a round over.
func persistModerationLogs(ctx context.Context, db *mongo.Database, logs []models.ModerationLog) {
	if len(logs) == 0 {
		return
	}
	docs := make([]interface{}, 0, len(logs))
	for _, l := range logs {
		docs = append(docs, l)
	}
	if _, err := db.Collection("moderation_logs").InsertMany(ctx, docs); err != nil {
		fmt.Printf("warning: failed to persist %d moderation log(s): %v\n", len(logs), err)
	}
}

// moderateSubmissions iterates over every submission in a round and returns
// the cleaned copies plus the corresponding audit logs. If the Gemini client
// can't be initialised, returns the originals untouched with a single log
// entry recording the skip.
func moderateSubmissions(ctx context.Context, submissions []models.Submission, roundID primitive.ObjectID, apiKey string) ([]models.Submission, []models.ModerationLog) {
	if apiKey == "" {
		return submissions, nil
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return submissions, []models.ModerationLog{{
			RoundID:     roundID,
			Model:       moderationModelName,
			Reasons:     []string{fmt.Sprintf("could not initialise moderation client: %v", err)},
			ModeratedAt: time.Now(),
		}}
	}
	defer client.Close()

	cleaned := make([]models.Submission, 0, len(submissions))
	logs := make([]models.ModerationLog, 0, len(submissions))
	for _, s := range submissions {
		out, log := moderateSubmission(ctx, client, s, roundID)
		cleaned = append(cleaned, out)
		logs = append(logs, log)
	}
	return cleaned, logs
}
