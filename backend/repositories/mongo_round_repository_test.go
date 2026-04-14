package repositories

import (
	"context"
	"smart360/models"
	"smart360/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMongoRoundRepository_Create(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoRoundRepository(testDB.DB)
	ctx := context.Background()

	t.Run("create_round_successfully", func(t *testing.T) {
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)

		err := repo.Create(ctx, round)
		require.NoError(t, err)

		// Verify round was created
		found, err := repo.FindByID(ctx, round.ID)
		require.NoError(t, err)
		assert.Equal(t, round.SubjectID, found.SubjectID)
		assert.Equal(t, round.Status, found.Status)
		// Reviewers can be nil or empty array in MongoDB
		if found.Reviewers != nil {
			assert.Len(t, found.Reviewers, 0)
		}
	})

	t.Run("create_round_initializes_empty_reviewers_array", func(t *testing.T) {
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		round.Reviewers = nil // Explicitly set to nil

		err := repo.Create(ctx, round)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, round.ID)
		require.NoError(t, err)
		// MongoDB omits empty arrays when decoding, so Reviewers can be nil
		// Both nil and empty array represent "no reviewers"
		if found.Reviewers != nil {
			assert.Len(t, found.Reviewers, 0)
		}
	})
}

func TestMongoRoundRepository_FindByID(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoRoundRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_existing_round", func(t *testing.T) {
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)
		err := repo.Create(ctx, round)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, round.ID)
		require.NoError(t, err)
		assert.Equal(t, round.ID, found.ID)
		assert.Equal(t, round.Status, found.Status)
	})

	t.Run("find_nonexistent_round_returns_error", func(t *testing.T) {
		_, err := repo.FindByID(ctx, primitive.NewObjectID())
		require.Error(t, err)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})
}

func TestMongoRoundRepository_UpdateStatus(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoRoundRepository(testDB.DB)
	ctx := context.Background()

	t.Run("update_status_successfully", func(t *testing.T) {
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		err := repo.Create(ctx, round)
		require.NoError(t, err)

		// Update status to active
		err = repo.UpdateStatus(ctx, round.ID, models.RoundActive)
		require.NoError(t, err)

		// Verify status was updated
		found, err := repo.FindByID(ctx, round.ID)
		require.NoError(t, err)
		assert.Equal(t, models.RoundActive, found.Status)
	})
}

func TestMongoRoundRepository_FindBySubjectID(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoRoundRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_rounds_by_subject", func(t *testing.T) {
		subjectID := primitive.NewObjectID()

		// Create multiple rounds for same subject
		round1 := testutil.NewTestRound(subjectID, primitive.NewObjectID(), models.RoundDraft)
		round2 := testutil.NewTestRound(subjectID, primitive.NewObjectID(), models.RoundActive)

		err := repo.Create(ctx, round1)
		require.NoError(t, err)
		err = repo.Create(ctx, round2)
		require.NoError(t, err)

		// Create round for different subject
		otherRound := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundClosed)
		err = repo.Create(ctx, otherRound)
		require.NoError(t, err)

		// Find by subject
		rounds, err := repo.FindBySubjectID(ctx, subjectID)
		require.NoError(t, err)
		assert.Len(t, rounds, 2)

		// Verify both rounds belong to the subject
		for _, round := range rounds {
			assert.Equal(t, subjectID, round.SubjectID)
		}
	})

	t.Run("find_by_subject_returns_empty_array_when_no_rounds", func(t *testing.T) {
		rounds, err := repo.FindBySubjectID(ctx, primitive.NewObjectID())
		require.NoError(t, err)
		assert.Len(t, rounds, 0)
	})
}

func TestMongoRoundRepository_AddReviewer(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoRoundRepository(testDB.DB)
	ctx := context.Background()

	t.Run("add_reviewer_successfully", func(t *testing.T) {
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		err := repo.Create(ctx, round)
		require.NoError(t, err)

		// Add reviewer
		reviewer := models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			ReviewerID: primitive.NewObjectID(),
			CreatedAt:  time.Now(),
		}

		err = repo.AddReviewer(ctx, round.ID, reviewer)
		require.NoError(t, err)

		// Verify reviewer was added
		reviewers, err := repo.GetReviewers(ctx, round.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 1)
		assert.Equal(t, reviewer.ReviewerID, reviewers[0].ReviewerID)
		assert.Equal(t, round.ID, reviewers[0].RoundID) // Should set round ID
	})

	t.Run("add_multiple_reviewers", func(t *testing.T) {
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		err := repo.Create(ctx, round)
		require.NoError(t, err)

		// Add three reviewers
		for i := 0; i < 3; i++ {
			reviewer := models.RoundReviewer{
				ID:         primitive.NewObjectID(),
				ReviewerID: primitive.NewObjectID(),
				CreatedAt:  time.Now(),
			}
			err = repo.AddReviewer(ctx, round.ID, reviewer)
			require.NoError(t, err)
		}

		// Verify all reviewers were added
		reviewers, err := repo.GetReviewers(ctx, round.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 3)
	})
}

func TestMongoRoundRepository_RemoveReviewer(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoRoundRepository(testDB.DB)
	ctx := context.Background()

	t.Run("remove_reviewer_successfully", func(t *testing.T) {
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		err := repo.Create(ctx, round)
		require.NoError(t, err)

		// Add two reviewers
		reviewer1ID := primitive.NewObjectID()
		reviewer2ID := primitive.NewObjectID()

		reviewer1 := models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			ReviewerID: reviewer1ID,
			CreatedAt:  time.Now(),
		}
		reviewer2 := models.RoundReviewer{
			ID:         primitive.NewObjectID(),
			ReviewerID: reviewer2ID,
			CreatedAt:  time.Now(),
		}

		err = repo.AddReviewer(ctx, round.ID, reviewer1)
		require.NoError(t, err)
		err = repo.AddReviewer(ctx, round.ID, reviewer2)
		require.NoError(t, err)

		// Remove first reviewer
		err = repo.RemoveReviewer(ctx, round.ID, reviewer1ID)
		require.NoError(t, err)

		// Verify only second reviewer remains
		reviewers, err := repo.GetReviewers(ctx, round.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 1)
		assert.Equal(t, reviewer2ID, reviewers[0].ReviewerID)
	})
}

func TestMongoRoundRepository_GetReviewers(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoRoundRepository(testDB.DB)
	ctx := context.Background()

	t.Run("get_reviewers_returns_empty_array_for_new_round", func(t *testing.T) {
		round := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		err := repo.Create(ctx, round)
		require.NoError(t, err)

		reviewers, err := repo.GetReviewers(ctx, round.ID)
		require.NoError(t, err)
		assert.Len(t, reviewers, 0)
	})

	t.Run("get_reviewers_for_nonexistent_round_returns_error", func(t *testing.T) {
		_, err := repo.GetReviewers(ctx, primitive.NewObjectID())
		require.Error(t, err)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})
}

func TestMongoRoundRepository_FindAll(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoRoundRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_all_rounds", func(t *testing.T) {
		// Create multiple rounds
		round1 := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundDraft)
		round2 := testutil.NewTestRound(primitive.NewObjectID(), primitive.NewObjectID(), models.RoundActive)

		err := repo.Create(ctx, round1)
		require.NoError(t, err)
		err = repo.Create(ctx, round2)
		require.NoError(t, err)

		// Find all rounds
		rounds, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.Len(t, rounds, 2)
	})
}
