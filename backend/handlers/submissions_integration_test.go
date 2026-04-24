package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"smart360/models"
	"smart360/repositories"
	"smart360/testutil"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Test-specific handler that uses repositories for dependency injection
type testSubmissionHandler struct {
	submissionRepo repositories.SubmissionRepository
	roundRepo      repositories.RoundRepository
}

func (h *testSubmissionHandler) SubmitFeedback(c *gin.Context) {
	var req struct {
		RoundID   string `json:"roundId" binding:"required"`
		Responses string `json:"responses" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	currentUser := user.(models.User)
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(req.RoundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Check if round exists
	_, err = h.roundRepo.FindByID(ctx, roundObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Check if submission already exists
	count, err := h.submissionRepo.CountByRoundAndReviewer(ctx, roundObjID, currentUser.ID)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Feedback already submitted"})
		return
	}

	// Create submission
	submission := &models.Submission{
		ID:          primitive.NewObjectID(),
		RoundID:     roundObjID,
		ReviewerID:  currentUser.ID,
		Responses:   req.Responses,
		SubmittedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = h.submissionRepo.Create(ctx, submission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, submission)
}

func (h *testSubmissionHandler) CheckSubmissionStatus(c *gin.Context) {
	roundID := c.Param("roundId")
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	submission, err := h.submissionRepo.FindByRoundAndReviewer(ctx, roundObjID, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"submitted": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"submitted": true, "submittedAt": submission.SubmittedAt})
}

func (h *testSubmissionHandler) GetRoundSubmissions(c *gin.Context) {
	roundID := c.Param("roundId")
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	submissions, err := h.submissionRepo.FindByRoundID(ctx, roundObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

// Integration Tests

func TestSubmitFeedback_Integration(t *testing.T) {
	submissionRepo := repositories.NewFakeSubmissionRepository()
	roundRepo := repositories.NewFakeRoundRepository()

	handler := &testSubmissionHandler{
		submissionRepo: submissionRepo,
		roundRepo:      roundRepo,
	}

	t.Run("successful_submission", func(t *testing.T) {
		// Setup test data
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		testUser := testutil.NewTestUser("reviewer@example.com", models.RoleMember)

		// Create test context
		c, w := testutil.NewTestGinContext(testUser)

		responses := map[string]string{
			"a": "Strong leadership skills",
			"b": "Could improve delegation",
			"c": "Proactive problem solver",
			"d": "Focus on strategic planning",
		}
		responsesJSON, _ := json.Marshal(responses)

		body := map[string]interface{}{
			"roundId":   testRound.ID.Hex(),
			"responses": string(responsesJSON),
		}
		err = testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		// Execute handler
		handler.SubmitFeedback(c)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify submission was stored
		submissions, err := submissionRepo.FindByRoundID(context.Background(), testRound.ID)
		require.NoError(t, err)
		assert.Len(t, submissions, 1)
		assert.Equal(t, testRound.ID, submissions[0].RoundID)
		assert.Equal(t, testUser.ID, submissions[0].ReviewerID)
		assert.Equal(t, string(responsesJSON), submissions[0].Responses)
	})

	t.Run("duplicate_submission_returns_conflict", func(t *testing.T) {
		// Setup test data
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		testUser := testutil.NewTestUser("reviewer2@example.com", models.RoleMember)

		// Create first submission
		responses := map[string]string{"a": "Good work"}
		responsesJSON, _ := json.Marshal(responses)
		firstSubmission := testutil.NewTestSubmission(testRound.ID, testUser.ID, string(responsesJSON))
		err = submissionRepo.Create(context.Background(), firstSubmission)
		require.NoError(t, err)

		// Try to submit again
		c, w := testutil.NewTestGinContext(testUser)
		body := map[string]interface{}{
			"roundId":   testRound.ID.Hex(),
			"responses": string(responsesJSON),
		}
		err = testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.SubmitFeedback(c)

		// Should return 409 Conflict
		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err = testutil.ParseJSONResponse(w, &response)
		require.NoError(t, err)
		assert.Equal(t, "Feedback already submitted", response["error"])
	})

	t.Run("invalid_round_id_returns_bad_request", func(t *testing.T) {
		testUser := testutil.NewTestUser("reviewer3@example.com", models.RoleMember)

		c, w := testutil.NewTestGinContext(testUser)
		body := map[string]interface{}{
			"roundId":   "invalid-id",
			"responses": "{}",
		}
		err := testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.SubmitFeedback(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("nonexistent_round_returns_not_found", func(t *testing.T) {
		testUser := testutil.NewTestUser("reviewer4@example.com", models.RoleMember)

		c, w := testutil.NewTestGinContext(testUser)
		body := map[string]interface{}{
			"roundId":   primitive.NewObjectID().Hex(),
			"responses": "{}",
		}
		err := testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.SubmitFeedback(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("missing_required_fields_returns_bad_request", func(t *testing.T) {
		testUser := testutil.NewTestUser("reviewer5@example.com", models.RoleMember)

		c, w := testutil.NewTestGinContext(testUser)
		body := map[string]interface{}{
			"roundId": primitive.NewObjectID().Hex(),
			// Missing "responses" field
		}
		err := testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.SubmitFeedback(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCheckSubmissionStatus_Integration(t *testing.T) {
	submissionRepo := repositories.NewFakeSubmissionRepository()
	roundRepo := repositories.NewFakeRoundRepository()

	handler := &testSubmissionHandler{
		submissionRepo: submissionRepo,
		roundRepo:      roundRepo,
	}

	t.Run("submitted_returns_true", func(t *testing.T) {
		// Setup test data
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		testUser := testutil.NewTestUser("reviewer@example.com", models.RoleMember)

		// Create submission
		submission := testutil.NewTestSubmission(testRound.ID, testUser.ID, "{}")
		err = submissionRepo.Create(context.Background(), submission)
		require.NoError(t, err)

		// Check status
		c, w := testutil.NewTestGinContext(testUser)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.CheckSubmissionStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = testutil.ParseJSONResponse(w, &response)
		require.NoError(t, err)
		assert.Equal(t, true, response["submitted"])
		assert.NotNil(t, response["submittedAt"])
	})

	t.Run("not_submitted_returns_false", func(t *testing.T) {
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		testUser := testutil.NewTestUser("reviewer2@example.com", models.RoleMember)

		// Check status without submission
		c, w := testutil.NewTestGinContext(testUser)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.CheckSubmissionStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = testutil.ParseJSONResponse(w, &response)
		require.NoError(t, err)
		assert.Equal(t, false, response["submitted"])
	})
}

func TestGetRoundSubmissions_Integration(t *testing.T) {
	submissionRepo := repositories.NewFakeSubmissionRepository()
	roundRepo := repositories.NewFakeRoundRepository()

	handler := &testSubmissionHandler{
		submissionRepo: submissionRepo,
		roundRepo:      roundRepo,
	}

	t.Run("returns_all_submissions_for_round", func(t *testing.T) {
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		// Create multiple submissions
		for i := 0; i < 3; i++ {
			submission := testutil.NewTestSubmission(testRound.ID, primitive.NewObjectID(), "{}")
			err = submissionRepo.Create(context.Background(), submission)
			require.NoError(t, err)
		}

		// Get submissions
		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.GetRoundSubmissions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var submissions []models.Submission
		err = testutil.ParseJSONResponse(w, &submissions)
		require.NoError(t, err)
		assert.Len(t, submissions, 3)
	})

	t.Run("returns_empty_array_for_round_without_submissions", func(t *testing.T) {
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.GetRoundSubmissions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var submissions []models.Submission
		err = testutil.ParseJSONResponse(w, &submissions)
		require.NoError(t, err)
		assert.Len(t, submissions, 0)
	})
}
