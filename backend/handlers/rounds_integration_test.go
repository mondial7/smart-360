package handlers

import (
	"context"
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
type testRoundHandler struct {
	roundRepo      repositories.RoundRepository
	userRepo       repositories.UserRepository
	submissionRepo repositories.SubmissionRepository
}

func (h *testRoundHandler) CreateFeedbackRound(c *gin.Context) {
	var req struct {
		SubjectID string     `json:"subjectId" binding:"required"`
		Deadline  *time.Time `json:"deadline"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	currentUser := user.(models.User)
	ctx := context.Background()

	// Convert subjectID to ObjectID
	subjectObjID, err := primitive.ObjectIDFromHex(req.SubjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subject ID"})
		return
	}

	// Verify subject exists
	_, err = h.userRepo.FindByID(ctx, subjectObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subject not found"})
		return
	}

	// Create round
	round := &models.FeedbackRound{
		ID:          primitive.NewObjectID(),
		SubjectID:   subjectObjID,
		CreatedByID: currentUser.ID,
		Status:      models.RoundDraft,
		Deadline:    req.Deadline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Reviewers:   []models.RoundReviewer{},
	}

	err = h.roundRepo.Create(ctx, round)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create round"})
		return
	}

	c.JSON(http.StatusCreated, round)
}

func (h *testRoundHandler) StartFeedbackRound(c *gin.Context) {
	roundID := c.Param("id")
	ctx := context.Background()

	// Convert roundID to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Get round
	round, err := h.roundRepo.FindByID(ctx, roundObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Validate status transition
	err = validateStatusTransition(round.Status, models.RoundActive)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update status
	err = h.roundRepo.UpdateStatus(ctx, roundObjID, models.RoundActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start round"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Round started successfully"})
}

func (h *testRoundHandler) CloseFeedbackRound(c *gin.Context) {
	roundID := c.Param("id")
	ctx := context.Background()

	// Convert roundID to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Get round
	round, err := h.roundRepo.FindByID(ctx, roundObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Validate status transition
	err = validateStatusTransition(round.Status, models.RoundClosed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update status
	err = h.roundRepo.UpdateStatus(ctx, roundObjID, models.RoundClosed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close round"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Round closed successfully"})
}

func (h *testRoundHandler) AddReviewersToRound(c *gin.Context) {
	roundID := c.Param("id")
	var req struct {
		ReviewerIDs []string `json:"reviewerIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Convert roundID to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Verify round exists
	_, err = h.roundRepo.FindByID(ctx, roundObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}

	// Add reviewers
	for _, reviewerIDStr := range req.ReviewerIDs {
		reviewerID, err := primitive.ObjectIDFromHex(reviewerIDStr)
		if err != nil {
			continue
		}

		reviewer := models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			RoundID:    roundObjID,
			ReviewerID: reviewerID,
			CreatedAt:  time.Now(),
		}

		err = h.roundRepo.AddReviewer(ctx, roundObjID, reviewer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reviewer"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reviewers added successfully"})
}

// GetRoundDetails mirrors handlers.GetRoundDetails authorization logic but
// uses the round repository directly (skipping the populated user fields).
// This is enough to exercise the auth branches and reviewer-stripping rule.
func (h *testRoundHandler) GetRoundDetails(c *gin.Context) {
	roundID := c.Param("id")
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

	populatedRound := &PopulatedRound{
		ID:          round.ID,
		SubjectID:   round.SubjectID,
		CreatedByID: round.CreatedByID,
		Status:      round.Status,
		Reviewers:   make([]PopulatedRoundReviewer, 0, len(round.Reviewers)),
	}
	for _, r := range round.Reviewers {
		populatedRound.Reviewers = append(populatedRound.Reviewers, PopulatedRoundReviewer{
			ID:         r.ID,
			RoundID:    r.RoundID,
			ReviewerID: r.ReviewerID,
			CreatedAt:  r.CreatedAt,
		})
	}

	isAdmin := currentUser.Role == models.RoleAdmin
	isCreator := currentUser.ID == populatedRound.CreatedByID
	isSubject := currentUser.ID == populatedRound.SubjectID
	isReviewer := false
	for _, r := range populatedRound.Reviewers {
		if r.ReviewerID == currentUser.ID {
			isReviewer = true
			break
		}
	}

	if !isAdmin && !isCreator && !isSubject && !isReviewer {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this round"})
		return
	}

	if !isAdmin && !isCreator {
		populatedRound.Reviewers = []PopulatedRoundReviewer{}
	}

	c.JSON(http.StatusOK, populatedRound)
}

// Integration Tests

func TestGetRoundDetails_Integration(t *testing.T) {
	t.Run("admin_sees_full_reviewer_roster", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		h := &testRoundHandler{roundRepo: roundRepo}

		creator := testutil.NewTestUser("creator-r1@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		require.NoError(t, roundRepo.AddReviewer(context.Background(), round.ID, models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			RoundID:    round.ID,
			ReviewerID: primitive.NewObjectID(),
			CreatedAt:  time.Now(),
		}))

		admin := testutil.NewTestUser("admin-r1@example.com", models.RoleAdmin)
		c, w := testutil.NewTestGinContext(admin)
		testutil.SetParam(c, "id", round.ID.Hex())

		h.GetRoundDetails(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var pr PopulatedRound
		require.NoError(t, testutil.ParseJSONResponse(w, &pr))
		assert.Len(t, pr.Reviewers, 1)
	})

	t.Run("creator_sees_full_reviewer_roster", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		h := &testRoundHandler{roundRepo: roundRepo}

		creator := testutil.NewTestUser("creator-r2@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		require.NoError(t, roundRepo.AddReviewer(context.Background(), round.ID, models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			RoundID:    round.ID,
			ReviewerID: primitive.NewObjectID(),
			CreatedAt:  time.Now(),
		}))

		c, w := testutil.NewTestGinContext(creator)
		testutil.SetParam(c, "id", round.ID.Hex())

		h.GetRoundDetails(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var pr PopulatedRound
		require.NoError(t, testutil.ParseJSONResponse(w, &pr))
		assert.Len(t, pr.Reviewers, 1)
	})

	t.Run("subject_sees_round_but_reviewer_roster_is_stripped", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		h := &testRoundHandler{roundRepo: roundRepo}

		subject := testutil.NewTestUser("subject-r3@example.com", models.RoleMember)
		creator := testutil.NewTestUser("creator-r3@example.com", models.RoleMember)
		round := testutil.NewTestRound(subject.ID, creator.ID, models.RoundShared)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		require.NoError(t, roundRepo.AddReviewer(context.Background(), round.ID, models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			RoundID:    round.ID,
			ReviewerID: primitive.NewObjectID(),
			CreatedAt:  time.Now(),
		}))

		c, w := testutil.NewTestGinContext(subject)
		testutil.SetParam(c, "id", round.ID.Hex())

		h.GetRoundDetails(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var pr PopulatedRound
		require.NoError(t, testutil.ParseJSONResponse(w, &pr))
		assert.Empty(t, pr.Reviewers, "subjects must not see who reviewed them")
	})

	t.Run("listed_reviewer_sees_round_but_no_other_reviewers", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		h := &testRoundHandler{roundRepo: roundRepo}

		creator := testutil.NewTestUser("creator-r4@example.com", models.RoleMember)
		reviewer := testutil.NewTestUser("reviewer-r4@example.com", models.RoleMember)
		other := testutil.NewTestUser("other-r4@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		require.NoError(t, roundRepo.AddReviewer(context.Background(), round.ID, models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			RoundID:    round.ID,
			ReviewerID: reviewer.ID,
			CreatedAt:  time.Now(),
		}))
		require.NoError(t, roundRepo.AddReviewer(context.Background(), round.ID, models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			RoundID:    round.ID,
			ReviewerID: other.ID,
			CreatedAt:  time.Now(),
		}))

		c, w := testutil.NewTestGinContext(reviewer)
		testutil.SetParam(c, "id", round.ID.Hex())

		h.GetRoundDetails(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var pr PopulatedRound
		require.NoError(t, testutil.ParseJSONResponse(w, &pr))
		assert.Empty(t, pr.Reviewers, "reviewers must not see other reviewers")
	})

	t.Run("unrelated_member_is_forbidden", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		h := &testRoundHandler{roundRepo: roundRepo}

		creator := testutil.NewTestUser("creator-r5@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundActive)
		require.NoError(t, roundRepo.Create(context.Background(), round))

		intruder := testutil.NewTestUser("intruder-r5@example.com", models.RoleMember)
		c, w := testutil.NewTestGinContext(intruder)
		testutil.SetParam(c, "id", round.ID.Hex())

		h.GetRoundDetails(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("nonexistent_round_returns_not_found", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		h := &testRoundHandler{roundRepo: roundRepo}

		admin := testutil.NewTestUser("admin-r6@example.com", models.RoleAdmin)
		c, w := testutil.NewTestGinContext(admin)
		testutil.SetParam(c, "id", primitive.NewObjectID().Hex())

		h.GetRoundDetails(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateFeedbackRound_Integration(t *testing.T) {
	roundRepo := repositories.NewFakeRoundRepository()
	userRepo := repositories.NewFakeUserRepository()
	submissionRepo := repositories.NewFakeSubmissionRepository()

	handler := &testRoundHandler{
		roundRepo:      roundRepo,
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
	}

	t.Run("successful_round_creation", func(t *testing.T) {
		// Create subject user
		subject := testutil.NewTestUser("subject@example.com", models.RoleMember)
		err := userRepo.Create(context.Background(), subject)
		require.NoError(t, err)

		// Create admin user
		admin := testutil.NewTestUser("admin@example.com", models.RoleAdmin)

		// Create test context
		c, w := testutil.NewTestGinContext(admin)

		deadline := time.Now().Add(7 * 24 * time.Hour)
		body := map[string]interface{}{
			"subjectId": subject.ID.Hex(),
			"deadline":  deadline,
		}
		err = testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		// Execute handler
		handler.CreateFeedbackRound(c)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.FeedbackRound
		err = testutil.ParseJSONResponse(w, &response)
		require.NoError(t, err)
		assert.Equal(t, subject.ID, response.SubjectID)
		assert.Equal(t, admin.ID, response.CreatedByID)
		assert.Equal(t, models.RoundDraft, response.Status)
	})

	t.Run("invalid_subject_id_returns_bad_request", func(t *testing.T) {
		admin := testutil.NewTestUser("admin2@example.com", models.RoleAdmin)

		c, w := testutil.NewTestGinContext(admin)
		body := map[string]interface{}{
			"subjectId": "invalid-id",
		}
		err := testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.CreateFeedbackRound(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("nonexistent_subject_returns_not_found", func(t *testing.T) {
		admin := testutil.NewTestUser("admin3@example.com", models.RoleAdmin)

		c, w := testutil.NewTestGinContext(admin)
		body := map[string]interface{}{
			"subjectId": primitive.NewObjectID().Hex(),
		}
		err := testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.CreateFeedbackRound(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestStartFeedbackRound_Integration(t *testing.T) {
	roundRepo := repositories.NewFakeRoundRepository()
	userRepo := repositories.NewFakeUserRepository()
	submissionRepo := repositories.NewFakeSubmissionRepository()

	handler := &testRoundHandler{
		roundRepo:      roundRepo,
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
	}

	t.Run("successfully_start_draft_round", func(t *testing.T) {
		// Create round in draft status
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		err := roundRepo.Create(context.Background(), round)
		require.NoError(t, err)

		// Start round
		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "id", round.ID.Hex())

		handler.StartFeedbackRound(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify status was updated
		updatedRound, err := roundRepo.FindByID(context.Background(), round.ID)
		require.NoError(t, err)
		assert.Equal(t, models.RoundActive, updatedRound.Status)
	})

	t.Run("cannot_start_active_round", func(t *testing.T) {
		// Create round in active status
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), round)
		require.NoError(t, err)

		// Try to start round again
		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "id", round.ID.Hex())

		handler.StartFeedbackRound(c)

		// Should return bad request due to invalid status transition
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("nonexistent_round_returns_not_found", func(t *testing.T) {
		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "id", primitive.NewObjectID().Hex())

		handler.StartFeedbackRound(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCloseFeedbackRound_Integration(t *testing.T) {
	roundRepo := repositories.NewFakeRoundRepository()
	userRepo := repositories.NewFakeUserRepository()
	submissionRepo := repositories.NewFakeSubmissionRepository()

	handler := &testRoundHandler{
		roundRepo:      roundRepo,
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
	}

	t.Run("successfully_close_active_round", func(t *testing.T) {
		// Create round in active status
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := roundRepo.Create(context.Background(), round)
		require.NoError(t, err)

		// Close round
		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "id", round.ID.Hex())

		handler.CloseFeedbackRound(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify status was updated
		updatedRound, err := roundRepo.FindByID(context.Background(), round.ID)
		require.NoError(t, err)
		assert.Equal(t, models.RoundClosed, updatedRound.Status)
	})

	t.Run("cannot_close_draft_round", func(t *testing.T) {
		// Create round in draft status
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		err := roundRepo.Create(context.Background(), round)
		require.NoError(t, err)

		// Try to close round
		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "id", round.ID.Hex())

		handler.CloseFeedbackRound(c)

		// Should return bad request due to invalid status transition
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAddReviewersToRound_Integration(t *testing.T) {
	roundRepo := repositories.NewFakeRoundRepository()
	userRepo := repositories.NewFakeUserRepository()
	submissionRepo := repositories.NewFakeSubmissionRepository()

	handler := &testRoundHandler{
		roundRepo:      roundRepo,
		userRepo:       userRepo,
		submissionRepo: submissionRepo,
	}

	t.Run("successfully_add_reviewers", func(t *testing.T) {
		// Create round
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		err := roundRepo.Create(context.Background(), round)
		require.NoError(t, err)

		// Create reviewers
		reviewer1 := testutil.NewTestUser("reviewer1@example.com", models.RoleMember)
		reviewer2 := testutil.NewTestUser("reviewer2@example.com", models.RoleMember)

		// Add reviewers
		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "id", round.ID.Hex())

		body := map[string]interface{}{
			"reviewerIds": []string{
				reviewer1.ID.Hex(),
				reviewer2.ID.Hex(),
			},
		}
		err = testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.AddReviewersToRound(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify reviewers were added
		reviewers, err := roundRepo.GetReviewers(context.Background(), round.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 2)
	})

	t.Run("nonexistent_round_returns_not_found", func(t *testing.T) {
		c, w := testutil.NewTestGinContext(nil)
		testutil.SetParam(c, "id", primitive.NewObjectID().Hex())

		body := map[string]interface{}{
			"reviewerIds": []string{primitive.NewObjectID().Hex()},
		}
		err := testutil.SetJSONBody(c, body)
		require.NoError(t, err)

		handler.AddReviewersToRound(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
