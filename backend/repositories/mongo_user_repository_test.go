package repositories

import (
	"context"
	"smart360/models"
	"smart360/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoUserRepository_Create(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("create_user_successfully", func(t *testing.T) {
		user := testutil.NewTestUser("test@example.com", models.RoleMember)

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Verify user was created
		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.Name, found.Name)
		assert.Equal(t, user.Role, found.Role)
	})

	t.Run("duplicate_email_returns_error", func(t *testing.T) {
		// Create index on email field to enforce uniqueness
		unique := true
		_, err := testDB.DB.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: map[string]interface{}{"email": 1},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		})
		require.NoError(t, err)

		user1 := testutil.NewTestUser("duplicate@example.com", models.RoleMember)
		err = repo.Create(ctx, user1)
		require.NoError(t, err)

		// Try to create another user with same email
		user2 := testutil.NewTestUser("duplicate@example.com", models.RoleAdmin)
		err = repo.Create(ctx, user2)

		// Should return duplicate key error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
	})
}

func TestMongoUserRepository_FindByID(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_existing_user", func(t *testing.T) {
		user := testutil.NewTestUser("find@example.com", models.RoleMember)
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("find_nonexistent_user_returns_error", func(t *testing.T) {
		nonexistentID := primitive.NewObjectID()

		_, err := repo.FindByID(ctx, nonexistentID)
		require.Error(t, err)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})
}

func TestMongoUserRepository_FindByEmail(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_existing_user_by_email", func(t *testing.T) {
		user := testutil.NewTestUser("email@example.com", models.RoleAdmin)
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByEmail(ctx, "email@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, models.RoleAdmin, found.Role)
	})

	t.Run("find_nonexistent_email_returns_error", func(t *testing.T) {
		_, err := repo.FindByEmail(ctx, "nonexistent@example.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})
}

func TestMongoUserRepository_UpdateRole(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("update_user_role_successfully", func(t *testing.T) {
		user := testutil.NewTestUser("role@example.com", models.RoleMember)
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Update role to admin
		err = repo.UpdateRole(ctx, user.ID, models.RoleAdmin)
		require.NoError(t, err)

		// Verify role was updated
		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, models.RoleAdmin, found.Role)
	})

	t.Run("update_nonexistent_user_succeeds_without_error", func(t *testing.T) {
		nonexistentID := primitive.NewObjectID()

		// MongoDB UpdateOne doesn't return error if document not found
		err := repo.UpdateRole(ctx, nonexistentID, models.RoleAdmin)
		assert.NoError(t, err)
	})
}

func TestMongoUserRepository_FindAll(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("find_all_users", func(t *testing.T) {
		// Create multiple users
		user1 := testutil.NewTestUser("user1@example.com", models.RoleMember)
		user2 := testutil.NewTestUser("user2@example.com", models.RoleAdmin)
		user3 := testutil.NewTestUser("user3@example.com", models.RoleTeamAdmin)

		err := repo.Create(ctx, user1)
		require.NoError(t, err)
		err = repo.Create(ctx, user2)
		require.NoError(t, err)
		err = repo.Create(ctx, user3)
		require.NoError(t, err)

		// Find all users
		users, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.Len(t, users, 3)

		// Verify all users are present
		emails := make(map[string]bool)
		for _, user := range users {
			emails[user.Email] = true
		}
		assert.True(t, emails["user1@example.com"])
		assert.True(t, emails["user2@example.com"])
		assert.True(t, emails["user3@example.com"])
	})

	t.Run("find_all_returns_empty_array_when_no_users", func(t *testing.T) {
		// Use a fresh database
		freshDB := testutil.SetupTestMongoDB(t)
		freshRepo := NewMongoUserRepository(freshDB.DB)

		users, err := freshRepo.FindAll(ctx)
		require.NoError(t, err)
		assert.Len(t, users, 0)
	})
}
