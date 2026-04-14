package repositories

import (
	"context"
	"encoding/json"
	"smart360/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMongoSubmissionRepository_Create(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoSubmissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("create_submission_successfully", func(t *testing.T) {
		responses := map[string]string{
			"a": "Strong technical skills",
			"b": "Could improve documentation",
		}
		responsesJSON, _ := json.Marshal(responses)

		submission := testutil.NewTestSubmission(
			primitive.NewObjectID(),
			primitive.NewObjectID(),
			string(responsesJSON),
		)

		err := repo.Create(ctx, submission)
		require.NoError(t, err)

		// Verify submission was created
		found, err := repo.FindByID(ctx, submission.ID)
		require.NoError(t, err)
		assert.Equal(t, submission.RoundID, found.RoundID)
		assert.Equal(t, submission.ReviewerID, found.ReviewerID)
		assert.Equal(t, submission.Responses, found.Responses)
	})
}

func TestMongoSubmissionRepository_FindByRoundID(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoSubmissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_submissions_by_round", func(t *testing.T) {
		roundID := primitive.NewObjectID()

		// Create multiple submissions for same round
		submission1 := testutil.NewTestSubmission(roundID, primitive.NewObjectID(), "{}")
		submission2 := testutil.NewTestSubmission(roundID, primitive.NewObjectID(), "{}")

		err := repo.Create(ctx, submission1)
		require.NoError(t, err)
		err = repo.Create(ctx, submission2)
		require.NoError(t, err)

		// Create submission for different round
		otherSubmission := testutil.NewTestSubmission(primitive.NewObjectID(), primitive.NewObjectID(), "{}")
		err = repo.Create(ctx, otherSubmission)
		require.NoError(t, err)

		// Find by round
		submissions, err := repo.FindByRoundID(ctx, roundID)
		require.NoError(t, err)
		assert.Len(t, submissions, 2)

		// Verify all submissions belong to the round
		for _, submission := range submissions {
			assert.Equal(t, roundID, submission.RoundID)
		}
	})

	t.Run("find_by_round_returns_empty_array_when_no_submissions", func(t *testing.T) {
		submissions, err := repo.FindByRoundID(ctx, primitive.NewObjectID())
		require.NoError(t, err)
		assert.Len(t, submissions, 0)
	})
}

func TestMongoSubmissionRepository_FindByRoundAndReviewer(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoSubmissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_existing_submission", func(t *testing.T) {
		roundID := primitive.NewObjectID()
		reviewerID := primitive.NewObjectID()

		submission := testutil.NewTestSubmission(roundID, reviewerID, "{\"a\":\"Good work\"}")
		err := repo.Create(ctx, submission)
		require.NoError(t, err)

		// Find by round and reviewer
		found, err := repo.FindByRoundAndReviewer(ctx, roundID, reviewerID)
		require.NoError(t, err)
		assert.Equal(t, submission.ID, found.ID)
		assert.Equal(t, roundID, found.RoundID)
		assert.Equal(t, reviewerID, found.ReviewerID)
	})

	t.Run("find_nonexistent_combination_returns_error", func(t *testing.T) {
		_, err := repo.FindByRoundAndReviewer(ctx, primitive.NewObjectID(), primitive.NewObjectID())
		require.Error(t, err)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})
}

func TestMongoSubmissionRepository_CountByRoundAndReviewer(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoSubmissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("count_returns_one_when_submission_exists", func(t *testing.T) {
		roundID := primitive.NewObjectID()
		reviewerID := primitive.NewObjectID()

		submission := testutil.NewTestSubmission(roundID, reviewerID, "{}")
		err := repo.Create(ctx, submission)
		require.NoError(t, err)

		count, err := repo.CountByRoundAndReviewer(ctx, roundID, reviewerID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("count_returns_zero_when_no_submission", func(t *testing.T) {
		count, err := repo.CountByRoundAndReviewer(ctx, primitive.NewObjectID(), primitive.NewObjectID())
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestMongoSubmissionRepository_Update(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoSubmissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("update_submission_successfully", func(t *testing.T) {
		submission := testutil.NewTestSubmission(
			primitive.NewObjectID(),
			primitive.NewObjectID(),
			"{\"a\":\"Original response\"}",
		)
		err := repo.Create(ctx, submission)
		require.NoError(t, err)

		// Update responses
		newResponses := map[string]string{
			"a": "Updated response",
			"b": "Additional feedback",
		}
		newResponsesJSON, _ := json.Marshal(newResponses)
		submission.Responses = string(newResponsesJSON)

		err = repo.Update(ctx, submission)
		require.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(ctx, submission.ID)
		require.NoError(t, err)
		assert.Equal(t, string(newResponsesJSON), found.Responses)
	})
}

func TestMongoSubmissionRepository_FindByID(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoSubmissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_existing_submission", func(t *testing.T) {
		submission := testutil.NewTestSubmission(
			primitive.NewObjectID(),
			primitive.NewObjectID(),
			"{}",
		)
		err := repo.Create(ctx, submission)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, submission.ID)
		require.NoError(t, err)
		assert.Equal(t, submission.ID, found.ID)
	})

	t.Run("find_nonexistent_submission_returns_error", func(t *testing.T) {
		_, err := repo.FindByID(ctx, primitive.NewObjectID())
		require.Error(t, err)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})
}

func TestMongoSubmissionRepository_DuplicatePrevention(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoSubmissionRepository(testDB.DB)
	ctx := context.Background()

	t.Run("allows_same_reviewer_different_rounds", func(t *testing.T) {
		reviewerID := primitive.NewObjectID()

		submission1 := testutil.NewTestSubmission(primitive.NewObjectID(), reviewerID, "{}")
		submission2 := testutil.NewTestSubmission(primitive.NewObjectID(), reviewerID, "{}")

		err := repo.Create(ctx, submission1)
		require.NoError(t, err)

		err = repo.Create(ctx, submission2)
		require.NoError(t, err) // Should succeed - different rounds
	})

	t.Run("allows_different_reviewers_same_round", func(t *testing.T) {
		roundID := primitive.NewObjectID()

		submission1 := testutil.NewTestSubmission(roundID, primitive.NewObjectID(), "{}")
		submission2 := testutil.NewTestSubmission(roundID, primitive.NewObjectID(), "{}")

		err := repo.Create(ctx, submission1)
		require.NoError(t, err)

		err = repo.Create(ctx, submission2)
		require.NoError(t, err) // Should succeed - different reviewers
	})
}
