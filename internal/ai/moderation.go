// Package ai holds the Gemini-backed moderation and consolidation-synthesis
// passes plus the deterministic competency aggregation. It is deliberately
// free of HTTP, database, and template concerns: callers pass in submissions
// and a template and get back a Consolidation and moderation audit logs.
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/mondial7/smart-360/internal/models"
)

// keyQueryParamRE matches any `key=<value>` query parameter. genai embeds the
// Gemini API key in request URLs, and its error strings include that URL
// verbatim — persisting it into an audit log would leak the live key. We
// rewrite the value to REDACTED before any error string is stored or logged.
var keyQueryParamRE = regexp.MustCompile(`(?i)([?&]key=)[^&"\s]+`)

// sanitiseErr renders err as a safe-to-log string, scrubbing the Gemini API key
// out of any URL the SDK threaded into the message.
func sanitiseErr(err error) string {
	if err == nil {
		return ""
	}
	return keyQueryParamRE.ReplaceAllString(err.Error(), "${1}REDACTED")
}

const (
	// moderationModelName is the Gemini model used for the scrub pass. Kept as a
	// const so the value lands in audit logs unambiguously.
	moderationModelName = "gemini-flash-latest"
	// moderationCallTimeout caps each individual moderation call. Gemini safety
	// filters occasionally don't return on borderline content; without a hard
	// cap a single hung call could block the whole round indefinitely.
	moderationCallTimeout = 25 * time.Second
	// moderationConcurrency caps in-flight moderation calls.
	moderationConcurrency = 5
)

// Progress is an optional callback the consolidation pipeline invokes so a
// caller (e.g. an SSE handler) can stream progress to the user. It must be safe
// to call from multiple goroutines.
type Progress func(ev ProgressEvent)

// ProgressEvent describes one step of the consolidation pipeline.
type ProgressEvent struct {
	Stage   string // "moderating" | "synthesizing" | "aggregating" | "done"
	Done    int
	Total   int
	Message string
}

func (p Progress) emit(ev ProgressEvent) {
	if p != nil {
		p(ev)
	}
}

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
  "cleaned": [
    {"key": "<one input field id>", "value": "<scrubbed text for that field>"}
  ],
  "flagged": <boolean — true if you removed or rewrote anything in any field>,
  "reasons": ["short bullets — one per rule applied, naming WHICH field was scrubbed and WHY"]
}

Include one "cleaned" entry per input field, preserving the original key. Return ONLY the minified JSON object. No code fences, no markdown, no prose.

Input:
%s`

type cleanedField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type moderationResult struct {
	Cleaned []cleanedField `json:"cleaned"`
	Flagged bool           `json:"flagged"`
	Reasons []string       `json:"reasons"`
}

func (r moderationResult) cleanedAsMap() map[string]string {
	m := make(map[string]string, len(r.Cleaned))
	for _, f := range r.Cleaned {
		m[f.Key] = f.Value
	}
	return m
}

// moderateSubmissions scrubs every submission in a round in parallel (capped
// concurrency) and returns the cleaned copies plus per-submission audit logs.
// When apiKey is empty the originals are returned untouched with no logs.
func moderateSubmissions(ctx context.Context, submissions []models.Submission, roundID, apiKey string, progress Progress) ([]models.Submission, []models.ModerationLog) {
	if apiKey == "" {
		return submissions, nil
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return submissions, []models.ModerationLog{{
			RoundID:     roundID,
			Model:       moderationModelName,
			Reasons:     []string{fmt.Sprintf("could not initialise moderation client: %s", sanitiseErr(err))},
			ModeratedAt: time.Now(),
		}}
	}
	defer client.Close()

	cleaned := make([]models.Submission, len(submissions))
	logs := make([]models.ModerationLog, len(submissions))
	sem := make(chan struct{}, moderationConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	done := 0
	total := len(submissions)

	for i, s := range submissions {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, s models.Submission) {
			defer wg.Done()
			defer func() { <-sem }()
			cleaned[i], logs[i] = moderateSubmission(ctx, client, s, roundID)
			mu.Lock()
			done++
			progress.emit(ProgressEvent{Stage: "moderating", Done: done, Total: total,
				Message: fmt.Sprintf("Screening submissions (%d/%d)", done, total)})
			mu.Unlock()
		}(i, s)
	}
	wg.Wait()
	return cleaned, logs
}

// moderateSubmission runs the scrub pass on one submission. On any failure the
// original is returned unchanged and the log records why, so a hiccup never
// blocks the round.
func moderateSubmission(ctx context.Context, client *genai.Client, submission models.Submission, roundID string) (models.Submission, models.ModerationLog) {
	log := models.ModerationLog{
		RoundID:      roundID,
		SubmissionID: submission.ID,
		Model:        moderationModelName,
		ModeratedAt:  time.Now(),
	}

	input := map[string]string{}
	for k, v := range submission.Responses {
		if strings.TrimSpace(v) != "" {
			input["response_"+k] = v
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
		return submission, log
	}

	inputJSON, _ := json.Marshal(input)
	prompt := fmt.Sprintf(moderationPrompt, string(inputJSON))

	model := client.GenerativeModel(moderationModelName)
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"cleaned": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"key":   {Type: genai.TypeString},
						"value": {Type: genai.TypeString},
					},
					Required: []string{"key", "value"},
				},
			},
			"flagged": {Type: genai.TypeBoolean},
			"reasons": {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
		},
		Required: []string{"cleaned", "flagged", "reasons"},
	}

	callCtx, cancel := context.WithTimeout(ctx, moderationCallTimeout)
	defer cancel()
	resp, err := model.GenerateContent(callCtx, genai.Text(prompt))
	if err != nil || len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Reasons = []string{fmt.Sprintf("moderation call failed: %s — original content was forwarded unchanged", sanitiseErr(err))}
		return submission, log
	}

	var raw string
	if t, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		raw = strings.TrimSpace(string(t))
	}

	result := moderationResult{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		log.Reasons = []string{fmt.Sprintf("moderation response was not valid JSON: %s — original content was forwarded unchanged", sanitiseErr(err))}
		return submission, log
	}

	scrubbed, fields := applyModerationCleaned(&submission, result.cleanedAsMap())
	log.Flagged = result.Flagged || len(fields) > 0
	log.Reasons = result.Reasons
	log.FieldsScrubbed = fields
	return scrubbed, log
}

// applyModerationCleaned writes the model's cleaned values back into a copy of
// the submission, returning the copy and the list of fields that actually
// changed. Responses is a typed map now, so no JSON string juggling is needed.
func applyModerationCleaned(submission *models.Submission, cleaned map[string]string) (models.Submission, []string) {
	out := *submission
	var changed []string

	if len(out.Responses) > 0 {
		responses := make(map[string]string, len(out.Responses))
		for k, v := range out.Responses {
			responses[k] = v
		}
		for key := range responses {
			if v, ok := cleaned["response_"+key]; ok && v != responses[key] {
				responses[key] = v
				changed = append(changed, "response_"+key)
			}
		}
		out.Responses = responses
	}

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
