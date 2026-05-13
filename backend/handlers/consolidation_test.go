package handlers

import (
	"encoding/json"
	"smart360/models"
	"strings"
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
		assert.Contains(t, result.ExecutiveSummary, "1 peer reviewers")

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
		assert.Contains(t, result.ExecutiveSummary, "3 peer reviewers")

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
		assert.Contains(t, result.ExecutiveSummary, "0 peer reviewers")

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

		// Should handle invalid JSON gracefully by skipping it: invalid submissions
		// do not contribute to the consolidation, so they should not be counted.
		assert.Equal(t, roundID, result.RoundID)
		assert.Contains(t, result.ExecutiveSummary, "0 peer reviewers")

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

	t.Run("self_submission_is_excluded_from_peer_aggregation_and_populates_delta", func(t *testing.T) {
		peerResponses, _ := json.Marshal(map[string]string{
			"a": "Strong on cross-team communication",
			"b": "Ships before testing edge cases",
			"c": "Mentorship",
			"d": "Pair with juniors weekly for the next two months",
		})
		selfResponses, _ := json.Marshal(map[string]string{
			"a": "I think I'm good at unblocking the team",
			"b": "I should probably document more",
			"c": "Technical depth",
			"d": "Read more system-design material",
		})

		submissions := []models.Submission{
			{Responses: string(peerResponses)},
			{Responses: string(selfResponses), IsSelf: true},
		}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		assert.Contains(t, result.ExecutiveSummary, "1 peer reviewers",
			"self submissions should not be counted as peer reviewers")
		assert.Contains(t, result.ExecutiveSummary, "self-assessment")

		// Peer aggregations should NOT include the self response.
		var strengths []string
		require.NoError(t, json.Unmarshal([]byte(result.Strengths), &strengths))
		assert.Len(t, strengths, 1)
		assert.Equal(t, "Strong on cross-team communication", strengths[0])

		// The delta should be populated and flagged as self-submitted.
		require.NotNil(t, result.SelfVsOthersDelta)
		assert.True(t, result.SelfVsOthersDelta.SelfSubmitted)
		assert.NotEmpty(t, result.SelfVsOthersDelta.Summary,
			"fallback should fill a coaching-flavoured summary when self is present")
		joined := strings.Join(append(append(append([]string{},
			result.SelfVsOthersDelta.BlindSpots...),
			result.SelfVsOthersDelta.HiddenStrengths...),
			result.SelfVsOthersDelta.Aligned...), " | ")
		assert.Contains(t, joined, "unblocking the team",
			"raw self answers should surface somewhere in the delta")
	})

	t.Run("no_self_submission_leaves_delta_flag_off", func(t *testing.T) {
		responses, _ := json.Marshal(map[string]string{"a": "Solid"})
		submissions := []models.Submission{{Responses: string(responses)}}

		result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

		require.NotNil(t, result.SelfVsOthersDelta)
		assert.False(t, result.SelfVsOthersDelta.SelfSubmitted)
		assert.Empty(t, result.SelfVsOthersDelta.Summary)
	})
}

func TestBuildFeedbackPrompts(t *testing.T) {
	t.Run("peer_blocks_include_relationship_and_frequency", func(t *testing.T) {
		peerResponses, _ := json.Marshal(map[string]string{"a": "Drives clarity in design reviews"})
		selfResponses, _ := json.Marshal(map[string]string{"a": "I try to facilitate"})

		submissions := []models.Submission{
			{
				Responses:            string(peerResponses),
				Relationship:         models.RelationshipManager,
				InteractionFrequency: models.InteractionDaily,
			},
			{Responses: string(selfResponses), IsSelf: true},
		}

		peerTexts, selfText, hasSelf := buildFeedbackPrompts(submissions)

		require.Len(t, peerTexts, 1)
		assert.Contains(t, peerTexts[0], "Relationship to subject: manager")
		assert.Contains(t, peerTexts[0], "Interaction frequency: daily")
		assert.Contains(t, peerTexts[0], "Drives clarity in design reviews")

		assert.True(t, hasSelf)
		assert.NotContains(t, selfText, "Relationship to subject",
			"self block should not include reviewer-relationship metadata")
		assert.Contains(t, selfText, "I try to facilitate")
	})

	t.Run("unspecified_relationship_is_labelled", func(t *testing.T) {
		peerResponses, _ := json.Marshal(map[string]string{"a": "x"})
		submissions := []models.Submission{
			{Responses: string(peerResponses)}, // no Relationship / InteractionFrequency
		}

		peerTexts, _, _ := buildFeedbackPrompts(submissions)

		require.Len(t, peerTexts, 1)
		assert.Contains(t, peerTexts[0], "Relationship to subject: unspecified")
		assert.Contains(t, peerTexts[0], "Interaction frequency: unspecified")
	})
}

func TestCountByVoice(t *testing.T) {
	submissions := []models.Submission{
		{Relationship: models.RelationshipManager},
		{Relationship: models.RelationshipPeer},
		{Relationship: models.RelationshipPeer},
		{Relationship: models.RelationshipCrossFunctional}, // bucketed under peer
		{Relationship: models.RelationshipReport},
		{IsSelf: true}, // never counted
		{IsSelf: true, Relationship: models.RelationshipManager},
	}

	mgr, peer, report := countByVoice(submissions)
	assert.Equal(t, 1, mgr, "exactly one manager submission")
	assert.Equal(t, 3, peer, "peer voice covers peer + cross_functional")
	assert.Equal(t, 1, report, "exactly one report submission")
}

func TestBuildVoiceBreakdown(t *testing.T) {
	payload := aiVoiceBreakdownPayload{
		ManagerVoice: aiVoicePayload{Summary: "mgr summary", Themes: []string{"mgr theme"}},
		PeerVoice:    aiVoicePayload{Summary: "peer summary", Themes: []string{"peer theme"}},
		ReportVoice:  aiVoicePayload{Summary: "report summary", Themes: []string{"report theme"}},
	}

	t.Run("populates_voices_with_non_zero_counts", func(t *testing.T) {
		vb := buildVoiceBreakdown(payload, 1, 2, 0)
		require.NotNil(t, vb)
		require.NotNil(t, vb.ManagerVoice)
		assert.Equal(t, 1, vb.ManagerVoice.ReviewerCount)
		assert.Equal(t, "mgr summary", vb.ManagerVoice.Summary)
		require.NotNil(t, vb.PeerVoice)
		assert.Equal(t, 2, vb.PeerVoice.ReviewerCount)
		assert.Nil(t, vb.ReportVoice, "voice with zero reviewers must be dropped")
	})

	t.Run("returns_nil_when_no_peer_submissions", func(t *testing.T) {
		assert.Nil(t, buildVoiceBreakdown(payload, 0, 0, 0))
	})
}

func TestCombineFeedbackSubmissions_VoiceBreakdown(t *testing.T) {
	roundID := primitive.NewObjectID()
	generatedByID := primitive.NewObjectID()

	mgrResp, _ := json.Marshal(map[string]string{"a": "Sets clear scope"})
	peerResp, _ := json.Marshal(map[string]string{"a": "Pairs generously"})
	reportResp, _ := json.Marshal(map[string]string{"a": "Gives concrete feedback"})

	submissions := []models.Submission{
		{Responses: string(mgrResp), Relationship: models.RelationshipManager},
		{Responses: string(peerResp), Relationship: models.RelationshipPeer},
		{Responses: string(reportResp), Relationship: models.RelationshipReport},
	}

	result := combineFeedbackSubmissions(submissions, roundID, generatedByID)

	require.NotNil(t, result.VoiceBreakdown)
	require.NotNil(t, result.VoiceBreakdown.ManagerVoice)
	assert.Equal(t, 1, result.VoiceBreakdown.ManagerVoice.ReviewerCount)
	assert.Contains(t, result.VoiceBreakdown.ManagerVoice.Themes[0], "Sets clear scope")

	require.NotNil(t, result.VoiceBreakdown.PeerVoice)
	assert.Contains(t, result.VoiceBreakdown.PeerVoice.Themes[0], "Pairs generously")

	require.NotNil(t, result.VoiceBreakdown.ReportVoice)
	assert.Contains(t, result.VoiceBreakdown.ReportVoice.Themes[0], "Gives concrete feedback")
}
