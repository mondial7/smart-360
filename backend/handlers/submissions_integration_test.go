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

	round, err := h.roundRepo.FindByID(ctx, roundObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}
	if round.Status != models.RoundActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Round is not accepting submissions"})
		return
	}
	if round.SubjectID == currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Subjects cannot submit feedback on their own round"})
		return
	}
	isReviewer := false
	for _, r := range round.Reviewers {
		if r.ReviewerID == currentUser.ID {
			isReviewer = true
			break
		}
	}
	if !isReviewer {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a reviewer for this round"})
		return
	}

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

func (h *testSubmissionHandler) GetSubmissionDetails(c *gin.Context) {
	submissionID := c.Param("submissionId")
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	currentUser := user.(models.User)
	ctx := context.Background()

	submissionObjID, err := primitive.ObjectIDFromHex(submissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	submission, err := h.submissionRepo.FindByID(ctx, submissionObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	if currentUser.Role != models.RoleAdmin && submission.ReviewerID != currentUser.ID {
		round, err := h.roundRepo.FindByID(ctx, submission.RoundID)
		if err != nil || round.CreatedByID != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this submission"})
			return
		}
	}

	c.JSON(http.StatusOK, submission)
}

func (h *testSubmissionHandler) GetRoundSubmissions(c *gin.Context) {
	roundID := c.Param("roundId")
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	currentUser := user.(models.User)
	ctx := context.Background()

	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	round, err := h.roundRepo.FindByID(ctx, roundObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}
	if currentUser.Role != models.RoleAdmin && round.CreatedByID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view submissions for this round"})
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

	enlistReviewer := func(t *testing.T, round *models.FeedbackRound, reviewerID primitive.ObjectID) {
		t.Helper()
		require.NoError(t, roundRepo.AddReviewer(context.Background(), round.ID, models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			RoundID:    round.ID,
			ReviewerID: reviewerID,
			CreatedAt:  time.Now(),
		}))
	}

	t.Run("successful_submission", func(t *testing.T) {
		testUser := testutil.NewTestUser("reviewer@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)
		enlistReviewer(t, testRound, testUser.ID)

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

		handler.SubmitFeedback(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		submissions, err := submissionRepo.FindByRoundID(context.Background(), testRound.ID)
		require.NoError(t, err)
		assert.Len(t, submissions, 1)
		assert.Equal(t, testRound.ID, submissions[0].RoundID)
		assert.Equal(t, testUser.ID, submissions[0].ReviewerID)
		assert.Equal(t, string(responsesJSON), submissions[0].Responses)
	})

	t.Run("duplicate_submission_returns_conflict", func(t *testing.T) {
		testUser := testutil.NewTestUser("reviewer2@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)
		enlistReviewer(t, testRound, testUser.ID)

		responses := map[string]string{"a": "Good work"}
		responsesJSON, _ := json.Marshal(responses)
		firstSubmission := testutil.NewTestSubmission(testRound.ID, testUser.ID, string(responsesJSON))
		err = submissionRepo.Create(context.Background(), firstSubmission)
		require.NoError(t, err)

		c, w := testutil.NewTestGinContext(testUser)
		body := map[string]interface{}{
			"roundId":   testRound.ID.Hex(),
			"responses": string(responsesJSON),
		}
		err = testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.SubmitFeedback(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err = testutil.ParseJSONResponse(w, &response)
		require.NoError(t, err)
		assert.Equal(t, "Feedback already submitted", response["error"])
	})

	t.Run("non_reviewer_is_forbidden", func(t *testing.T) {
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		intruder := testutil.NewTestUser("phantom@example.com", models.RoleMember)
		c, w := testutil.NewTestGinContext(intruder)
		require.NoError(t, testutil.SetJSONBody(c, map[string]interface{}{
			"roundId":   testRound.ID.Hex(),
			"responses": "{}",
		}))

		handler.SubmitFeedback(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("subject_cannot_submit_on_own_round", func(t *testing.T) {
		subject := testutil.NewTestUser("subject@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(subject.ID, primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)
		enlistReviewer(t, testRound, subject.ID)

		c, w := testutil.NewTestGinContext(subject)
		require.NoError(t, testutil.SetJSONBody(c, map[string]interface{}{
			"roundId":   testRound.ID.Hex(),
			"responses": "{}",
		}))

		handler.SubmitFeedback(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("inactive_round_is_rejected", func(t *testing.T) {
		reviewer := testutil.NewTestUser("late-reviewer@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundClosed)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)
		enlistReviewer(t, testRound, reviewer.ID)

		c, w := testutil.NewTestGinContext(reviewer)
		require.NoError(t, testutil.SetJSONBody(c, map[string]interface{}{
			"roundId":   testRound.ID.Hex(),
			"responses": "{}",
		}))

		handler.SubmitFeedback(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
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

func TestGetSubmissionDetails_Integration(t *testing.T) {
	submissionRepo := repositories.NewFakeSubmissionRepository()
	roundRepo := repositories.NewFakeRoundRepository()

	handler := &testSubmissionHandler{
		submissionRepo: submissionRepo,
		roundRepo:      roundRepo,
	}

	t.Run("reviewer_can_read_own_submission", func(t *testing.T) {
		creator := testutil.NewTestUser("creator-d1@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		require.NoError(t, roundRepo.Create(context.Background(), round))

		reviewer := testutil.NewTestUser("reviewer-d1@example.com", models.RoleMember)
		submission := testutil.NewTestSubmission(round.ID, reviewer.ID, `{"a":"x"}`)
		require.NoError(t, submissionRepo.Create(context.Background(), submission))

		c, w := testutil.NewTestGinContext(reviewer)
		testutil.SetParam(c, "submissionId", submission.ID.Hex())

		handler.GetSubmissionDetails(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("round_creator_can_read_any_submission_for_their_round", func(t *testing.T) {
		creator := testutil.NewTestUser("creator-d2@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		require.NoError(t, roundRepo.Create(context.Background(), round))

		reviewer := testutil.NewTestUser("reviewer-d2@example.com", models.RoleMember)
		submission := testutil.NewTestSubmission(round.ID, reviewer.ID, `{"a":"x"}`)
		require.NoError(t, submissionRepo.Create(context.Background(), submission))

		c, w := testutil.NewTestGinContext(creator)
		testutil.SetParam(c, "submissionId", submission.ID.Hex())

		handler.GetSubmissionDetails(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("admin_can_read_any_submission", func(t *testing.T) {
		creator := testutil.NewTestUser("creator-d3@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		require.NoError(t, roundRepo.Create(context.Background(), round))

		reviewer := testutil.NewTestUser("reviewer-d3@example.com", models.RoleMember)
		submission := testutil.NewTestSubmission(round.ID, reviewer.ID, `{"a":"x"}`)
		require.NoError(t, submissionRepo.Create(context.Background(), submission))

		admin := testutil.NewTestUser("admin-d3@example.com", models.RoleAdmin)
		c, w := testutil.NewTestGinContext(admin)
		testutil.SetParam(c, "submissionId", submission.ID.Hex())

		handler.GetSubmissionDetails(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("non_owner_non_creator_non_admin_is_forbidden", func(t *testing.T) {
		creator := testutil.NewTestUser("creator-d4@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		require.NoError(t, roundRepo.Create(context.Background(), round))

		reviewer := testutil.NewTestUser("reviewer-d4@example.com", models.RoleMember)
		submission := testutil.NewTestSubmission(round.ID, reviewer.ID, `{"a":"x"}`)
		require.NoError(t, submissionRepo.Create(context.Background(), submission))

		intruder := testutil.NewTestUser("intruder-d4@example.com", models.RoleMember)
		c, w := testutil.NewTestGinContext(intruder)
		testutil.SetParam(c, "submissionId", submission.ID.Hex())

		handler.GetSubmissionDetails(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("nonexistent_submission_returns_not_found", func(t *testing.T) {
		admin := testutil.NewTestUser("admin-d5@example.com", models.RoleAdmin)
		c, w := testutil.NewTestGinContext(admin)
		testutil.SetParam(c, "submissionId", primitive.NewObjectID().Hex())

		handler.GetSubmissionDetails(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetRoundSubmissions_Integration(t *testing.T) {
	submissionRepo := repositories.NewFakeSubmissionRepository()
	roundRepo := repositories.NewFakeRoundRepository()

	handler := &testSubmissionHandler{
		submissionRepo: submissionRepo,
		roundRepo:      roundRepo,
	}

	t.Run("creator_can_read_all_submissions_for_round", func(t *testing.T) {
		creator := testutil.NewTestUser("creator@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			submission := testutil.NewTestSubmission(testRound.ID, primitive.NewObjectID(), "{}")
			err = submissionRepo.Create(context.Background(), submission)
			require.NoError(t, err)
		}

		c, w := testutil.NewTestGinContext(creator)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.GetRoundSubmissions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var submissions []models.Submission
		err = testutil.ParseJSONResponse(w, &submissions)
		require.NoError(t, err)
		assert.Len(t, submissions, 3)
	})

	t.Run("admin_can_read_all_submissions_for_round", func(t *testing.T) {
		creator := testutil.NewTestUser("creator2@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		admin := testutil.NewTestUser("admin@example.com", models.RoleAdmin)
		c, w := testutil.NewTestGinContext(admin)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.GetRoundSubmissions(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("non_creator_non_admin_is_forbidden", func(t *testing.T) {
		creator := testutil.NewTestUser("creator3@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		intruder := testutil.NewTestUser("intruder@example.com", models.RoleMember)
		c, w := testutil.NewTestGinContext(intruder)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.GetRoundSubmissions(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("team_admin_who_is_not_creator_is_forbidden", func(t *testing.T) {
		creator := testutil.NewTestUser("creator4@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		teamAdmin := testutil.NewTestUser("teamadmin@example.com", models.RoleTeamAdmin)
		c, w := testutil.NewTestGinContext(teamAdmin)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.GetRoundSubmissions(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("nonexistent_round_returns_not_found", func(t *testing.T) {
		admin := testutil.NewTestUser("admin2@example.com", models.RoleAdmin)
		c, w := testutil.NewTestGinContext(admin)
		testutil.SetParam(c, "roundId", primitive.NewObjectID().Hex())

		handler.GetRoundSubmissions(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns_empty_array_for_round_without_submissions", func(t *testing.T) {
		creator := testutil.NewTestUser("creator5@example.com", models.RoleMember)
		testRound := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		err := roundRepo.Create(context.Background(), testRound)
		require.NoError(t, err)

		c, w := testutil.NewTestGinContext(creator)
		testutil.SetParam(c, "roundId", testRound.ID.Hex())

		handler.GetRoundSubmissions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var submissions []models.Submission
		err = testutil.ParseJSONResponse(w, &submissions)
		require.NoError(t, err)
		assert.Len(t, submissions, 0)
	})
}
