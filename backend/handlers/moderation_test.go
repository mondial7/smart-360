package handlers

import (
	"encoding/json"
	"smart360/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyModerationCleaned(t *testing.T) {
	originalResponses, _ := json.Marshal(map[string]string{
		"a": "She's just lazy.",
		"b": "Misses deadlines on the Acme project.",
	})
	original := models.Submission{
		Responses: string(originalResponses),
		Ratings: []models.CompetencyRating{
			{Key: "execution", Score: 3, Justification: "He's careless about details."},
			{Key: "collaboration", Score: 4, Justification: "Pairs generously."},
		},
		PrivateNotes: "I think she's too old for this team.",
	}

	cleaned := map[string]string{
		"response_a":                "Doesn't consistently follow through on commitments.",
		"response_b":                "Misses deadlines on the Acme project.", // unchanged
		"rating_0_justification":    "Sometimes ships before checking edge cases.",
		"rating_1_justification":    "Pairs generously.", // unchanged
		"private_notes":             "Stretch may exceed current scope; calibrate expectations.",
	}

	out, changed := applyModerationCleaned(&original, cleaned)

	// Original document untouched (we copied).
	assert.NotEqual(t, out.Responses, original.Responses)
	assert.Equal(t, original.PrivateNotes, "I think she's too old for this team.",
		"the input submission must not be mutated in place")

	// Responses updated for "a" only.
	var responsesOut map[string]string
	require.NoError(t, json.Unmarshal([]byte(out.Responses), &responsesOut))
	assert.Equal(t, "Doesn't consistently follow through on commitments.", responsesOut["a"])
	assert.Equal(t, "Misses deadlines on the Acme project.", responsesOut["b"])

	// Ratings: only the first justification changed.
	require.Len(t, out.Ratings, 2)
	assert.Equal(t, "Sometimes ships before checking edge cases.", out.Ratings[0].Justification)
	assert.Equal(t, "Pairs generously.", out.Ratings[1].Justification)
	// Original slice intact.
	assert.Equal(t, "He's careless about details.", original.Ratings[0].Justification,
		"original ratings slice must not be mutated")

	assert.Equal(t, "Stretch may exceed current scope; calibrate expectations.", out.PrivateNotes)

	assert.ElementsMatch(t,
		[]string{"response_a", "rating_0_justification", "private_notes"},
		changed,
		"changed list reflects exactly the fields whose content was rewritten")
}

func TestApplyModerationCleaned_NoChanges(t *testing.T) {
	responsesJSON, _ := json.Marshal(map[string]string{"a": "Solid."})
	submission := models.Submission{Responses: string(responsesJSON)}

	out, changed := applyModerationCleaned(&submission, map[string]string{
		"response_a": "Solid.", // identical
	})

	assert.Empty(t, changed, "matching content yields no diff")
	assert.Equal(t, submission.Responses, out.Responses)
}
