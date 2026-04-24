package handlers

import (
	"encoding/json"
	"smart360/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCombineFeedbackSubmissions(t *testing.T) {
	roundID := primitive.NewObjectID()
	generatedByID := primitive.NewObjectID()

	t.Run("single_submission_with_all_responses", func(t *testing.T) {
		responses := map[string]string{
			"a": "Great communication skills",
			"b": "Needs to improve delegation",
			"c": "Takes initiative regularly",
			"d": "Set clearer priorities and deadlines",
		}
		responsesJSON, _ := json.Marshal(responses)

		submissions := []models.Submission{
			{
				ID:         primitive.NewObjectID(),
				RoundID:    roundID,
				ReviewerID: primitive.NewObjectID(),
				Responses:  string(responsesJSON),
			},
		}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		// Verify basic fields
		assert.Equal(t, roundID, result.RoundID)
		assert.Equal(t, generatedByID, result.GeneratedByID)
		assert.NotEmpty(t, result.ExecutiveSummary)
		assert.Contains(t, result.ExecutiveSummary, "1 reviewers")

		// Verify strengths were captured
		var strengths []string
		err := json.Unmarshal([]byte(result.Strengths), &strengths)
		require.NoError(t, err)
		assert.Len(t, strengths, 1)
		assert.Equal(t, "Great communication skills", strengths[0])

		// Verify improvements were captured
		var improvements []string
		err = json.Unmarshal([]byte(result.AreasForImprovement), &improvements)
		require.NoError(t, err)
		assert.Len(t, improvements, 1)
		assert.Equal(t, "Needs to improve delegation", improvements[0])

		// Verify actionable insights were captured
		var insights []string
		err = json.Unmarshal([]byte(result.ActionableInsights), &insights)
		require.NoError(t, err)
		assert.Len(t, insights, 1)
		assert.Equal(t, "Set clearer priorities and deadlines", insights[0])

		// Verify question summaries
		var questionSummaries map[string]string
		err = json.Unmarshal([]byte(result.QuestionSummaries), &questionSummaries)
		require.NoError(t, err)
		assert.Len(t, questionSummaries, 4)
		assert.Contains(t, questionSummaries["a"], "Great communication skills")
		assert.Contains(t, questionSummaries["b"], "Needs to improve delegation")
		assert.Contains(t, questionSummaries["c"], "Takes initiative regularly")
		assert.Contains(t, questionSummaries["d"], "Set clearer priorities")
	})

	t.Run("multiple_submissions", func(t *testing.T) {
		responses1 := map[string]string{
			"a": "Strong technical skills",
			"b": "Could improve documentation",
			"c": "Proactive problem solver",
			"d": "Focus on mentoring junior developers",
		}
		responses1JSON, _ := json.Marshal(responses1)

		responses2 := map[string]string{
			"a": "Excellent collaboration",
			"b": "Sometimes misses deadlines",
			"c": "Great team player",
			"d": "Improve time management skills",
		}
		responses2JSON, _ := json.Marshal(responses2)

		responses3 := map[string]string{
			"a": "Creative problem solving",
			"b": "Need better code reviews",
			"c": "Helps teammates voluntarily",
			"d": "Continue learning new technologies",
		}
		responses3JSON, _ := json.Marshal(responses3)

		submissions := []models.Submission{
			{Responses: string(responses1JSON)},
			{Responses: string(responses2JSON)},
			{Responses: string(responses3JSON)},
		}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		// Verify all submissions are counted
		assert.Contains(t, result.ExecutiveSummary, "3 reviewers")

		// Verify all strengths were captured
		var strengths []string
		err := json.Unmarshal([]byte(result.Strengths), &strengths)
		require.NoError(t, err)
		assert.Len(t, strengths, 3)
		assert.Contains(t, strengths, "Strong technical skills")
		assert.Contains(t, strengths, "Excellent collaboration")
		assert.Contains(t, strengths, "Creative problem solving")

		// Verify all improvements were captured
		var improvements []string
		err = json.Unmarshal([]byte(result.AreasForImprovement), &improvements)
		require.NoError(t, err)
		assert.Len(t, improvements, 3)

		// Verify actionable insights (all should be included as they're >10 chars)
		var insights []string
		err = json.Unmarshal([]byte(result.ActionableInsights), &insights)
		require.NoError(t, err)
		assert.Len(t, insights, 3)

		// Verify question summaries combine all responses
		var questionSummaries map[string]string
		err = json.Unmarshal([]byte(result.QuestionSummaries), &questionSummaries)
		require.NoError(t, err)
		assert.Contains(t, questionSummaries["a"], "Strong technical skills")
		assert.Contains(t, questionSummaries["a"], "Excellent collaboration")
		assert.Contains(t, questionSummaries["a"], "Creative problem solving")
	})

	t.Run("empty_submissions_array", func(t *testing.T) {
		submissions := []models.Submission{}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		// Should still return a valid consolidation
		assert.Equal(t, roundID, result.RoundID)
		assert.Equal(t, generatedByID, result.GeneratedByID)
		assert.Contains(t, result.ExecutiveSummary, "0 reviewers")

		// All arrays should be empty
		var strengths []string
		err := json.Unmarshal([]byte(result.Strengths), &strengths)
		require.NoError(t, err)
		assert.Len(t, strengths, 0)

		var questionSummaries map[string]string
		err = json.Unmarshal([]byte(result.QuestionSummaries), &questionSummaries)
		require.NoError(t, err)
		assert.Len(t, questionSummaries, 0)
	})

	t.Run("submission_with_missing_response_keys", func(t *testing.T) {
		// Only provide keys "a" and "d"
		responses := map[string]string{
			"a": "Good leadership",
			"d": "Improve conflict resolution",
		}
		responsesJSON, _ := json.Marshal(responses)

		submissions := []models.Submission{
			{Responses: string(responsesJSON)},
		}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		// Verify only provided keys are present
		var strengths []string
		err := json.Unmarshal([]byte(result.Strengths), &strengths)
		require.NoError(t, err)
		assert.Len(t, strengths, 1)

		var improvements []string
		err = json.Unmarshal([]byte(result.AreasForImprovement), &improvements)
		require.NoError(t, err)
		assert.Len(t, improvements, 0) // Key "b" was not provided

		var questionSummaries map[string]string
		err = json.Unmarshal([]byte(result.QuestionSummaries), &questionSummaries)
		require.NoError(t, err)
		assert.Len(t, questionSummaries, 2) // Only "a" and "d"
		assert.Contains(t, questionSummaries, "a")
		assert.Contains(t, questionSummaries, "d")
		assert.NotContains(t, questionSummaries, "b")
		assert.NotContains(t, questionSummaries, "c")
	})

	t.Run("submission_with_empty_string_responses", func(t *testing.T) {
		responses := map[string]string{
			"a": "Good teamwork",
			"b": "",  // Empty string should be ignored
			"c": "  ", // Whitespace-only should be included (not trimmed by function)
			"d": "",  // Empty string should be ignored
		}
		responsesJSON, _ := json.Marshal(responses)

		submissions := []models.Submission{
			{Responses: string(responsesJSON)},
		}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		// Empty strings should be filtered out
		var strengths []string
		err := json.Unmarshal([]byte(result.Strengths), &strengths)
		require.NoError(t, err)
		assert.Len(t, strengths, 1)

		var improvements []string
		err = json.Unmarshal([]byte(result.AreasForImprovement), &improvements)
		require.NoError(t, err)
		assert.Len(t, improvements, 0)

		// Whitespace-only strings are included (function doesn't trim)
		var questionSummaries map[string]string
		err = json.Unmarshal([]byte(result.QuestionSummaries), &questionSummaries)
		require.NoError(t, err)
		assert.Contains(t, questionSummaries, "a")
		assert.Contains(t, questionSummaries, "c") // Whitespace is not empty
		assert.NotContains(t, questionSummaries, "b")
		assert.NotContains(t, questionSummaries, "d")
	})

	t.Run("submission_with_invalid_json", func(t *testing.T) {
		submissions := []models.Submission{
			{Responses: "not valid json"},
			{Responses: ""},
		}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		// Should handle invalid JSON gracefully by skipping it
		assert.Equal(t, roundID, result.RoundID)
		assert.Contains(t, result.ExecutiveSummary, "2 reviewers")

		// All arrays should be empty since JSON parsing failed
		var strengths []string
		err := json.Unmarshal([]byte(result.Strengths), &strengths)
		require.NoError(t, err)
		assert.Len(t, strengths, 0)
	})

	t.Run("actionable_insights_filters_short_advice", func(t *testing.T) {
		responses := map[string]string{
			"a": "Good work",
			"d": "Keep it up", // Short advice (10 chars exactly - should be filtered)
		}
		responsesJSON, _ := json.Marshal(responses)

		submissions := []models.Submission{
			{Responses: string(responsesJSON)},
		}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		// Short advice should not be included in actionable insights
		var insights []string
		err := json.Unmarshal([]byte(result.ActionableInsights), &insights)
		require.NoError(t, err)
		assert.Len(t, insights, 0) // "Keep it up" is exactly 10 chars, filtered out
	})

	t.Run("verify_timestamps_are_set", func(t *testing.T) {
		responses := map[string]string{"a": "Test"}
		responsesJSON, _ := json.Marshal(responses)

		submissions := []models.Submission{
			{Responses: string(responsesJSON)},
		}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		// Verify timestamps are set
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())
	})
}
