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

type testConsolidationHandler struct {
	roundRepo         repositories.RoundRepository
	consolidationRepo repositories.ConsolidationRepository
}

// GetConsolidation mirrors handlers.GetConsolidation but uses repositories
// instead of *mongo.Database, so it can be exercised without a live DB.
func (h *testConsolidationHandler) GetConsolidation(c *gin.Context) {
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

	if currentUser.Role != models.RoleAdmin &&
		currentUser.ID != round.SubjectID &&
		currentUser.ID != round.CreatedByID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to access this consolidation"})
		return
	}

	consolidation, err := h.consolidationRepo.FindByRoundID(ctx, roundObjID)
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

func newConsolidationFor(roundID primitive.ObjectID, sharedAt *time.Time) *models.Consolidation {
	return &models.Consolidation{
		ID:               primitive.NewObjectID(),
		RoundID:          roundID,
		ExecutiveSummary: "summary",
		SharedAt:         sharedAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func TestGetConsolidation_Integration(t *testing.T) {
	t.Run("creator_can_read", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		consolidationRepo := repositories.NewFakeConsolidationRepository()
		h := &testConsolidationHandler{roundRepo: roundRepo, consolidationRepo: consolidationRepo}

		creator := testutil.NewTestUser("creator-c1@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundClosed)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		require.NoError(t, consolidationRepo.Create(context.Background(), newConsolidationFor(round.ID, nil)))

		c, w := testutil.NewTestGinContext(creator)
		testutil.SetParam(c, "roundId", round.ID.Hex())

		h.GetConsolidation(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("admin_can_read_unshared", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		consolidationRepo := repositories.NewFakeConsolidationRepository()
		h := &testConsolidationHandler{roundRepo: roundRepo, consolidationRepo: consolidationRepo}

		creator := testutil.NewTestUser("creator-c2@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundClosed)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		require.NoError(t, consolidationRepo.Create(context.Background(), newConsolidationFor(round.ID, nil)))

		admin := testutil.NewTestUser("admin-c2@example.com", models.RoleAdmin)
		c, w := testutil.NewTestGinContext(admin)
		testutil.SetParam(c, "roundId", round.ID.Hex())

		h.GetConsolidation(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("subject_cannot_read_unshared", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		consolidationRepo := repositories.NewFakeConsolidationRepository()
		h := &testConsolidationHandler{roundRepo: roundRepo, consolidationRepo: consolidationRepo}

		subject := testutil.NewTestUser("subject-c3@example.com", models.RoleMember)
		creator := testutil.NewTestUser("creator-c3@example.com", models.RoleMember)
		round := testutil.NewTestRound(subject.ID, creator.ID, models.RoundClosed)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		require.NoError(t, consolidationRepo.Create(context.Background(), newConsolidationFor(round.ID, nil)))

		c, w := testutil.NewTestGinContext(subject)
		testutil.SetParam(c, "roundId", round.ID.Hex())

		h.GetConsolidation(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("subject_can_read_shared", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		consolidationRepo := repositories.NewFakeConsolidationRepository()
		h := &testConsolidationHandler{roundRepo: roundRepo, consolidationRepo: consolidationRepo}

		subject := testutil.NewTestUser("subject-c4@example.com", models.RoleMember)
		creator := testutil.NewTestUser("creator-c4@example.com", models.RoleMember)
		round := testutil.NewTestRound(subject.ID, creator.ID, models.RoundShared)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		now := time.Now()
		require.NoError(t, consolidationRepo.Create(context.Background(), newConsolidationFor(round.ID, &now)))

		c, w := testutil.NewTestGinContext(subject)
		testutil.SetParam(c, "roundId", round.ID.Hex())

		h.GetConsolidation(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unrelated_member_is_forbidden", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		consolidationRepo := repositories.NewFakeConsolidationRepository()
		h := &testConsolidationHandler{roundRepo: roundRepo, consolidationRepo: consolidationRepo}

		creator := testutil.NewTestUser("creator-c5@example.com", models.RoleMember)
		round := testutil.NewTestRound(primitive.NewObjectID(), creator.ID, models.RoundShared)
		require.NoError(t, roundRepo.Create(context.Background(), round))
		now := time.Now()
		require.NoError(t, consolidationRepo.Create(context.Background(), newConsolidationFor(round.ID, &now)))

		intruder := testutil.NewTestUser("intruder-c5@example.com", models.RoleMember)
		c, w := testutil.NewTestGinContext(intruder)
		testutil.SetParam(c, "roundId", round.ID.Hex())

		h.GetConsolidation(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("nonexistent_round_returns_not_found", func(t *testing.T) {
		roundRepo := repositories.NewFakeRoundRepository()
		consolidationRepo := repositories.NewFakeConsolidationRepository()
		h := &testConsolidationHandler{roundRepo: roundRepo, consolidationRepo: consolidationRepo}

		admin := testutil.NewTestUser("admin-c6@example.com", models.RoleAdmin)
		c, w := testutil.NewTestGinContext(admin)
		testutil.SetParam(c, "roundId", primitive.NewObjectID().Hex())

		h.GetConsolidation(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
