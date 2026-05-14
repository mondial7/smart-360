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
)

func newTemplate(slug, name string) *models.Template {
	return &models.Template{
		Slug:            slug,
		Name:            name,
		Description:     "test template",
		CoachingPersona: "a coach",
		Questions: []models.TemplateQuestion{
			{Key: "a", PeerText: "peer A", SelfText: "self A", CardTitle: "title A"},
			{Key: "b", PeerText: "peer B", SelfText: "self B", CardTitle: "title B"},
		},
	}
}

func TestMongoTemplateRepository_UpsertAndLookup(t *testing.T) {
	testDB := testutil.SetupTestMongoDB(t)
	repo := NewMongoTemplateRepository(testDB.DB)
	ctx := context.Background()

	t.Run("upsert_assigns_id_and_timestamps", func(t *testing.T) {
		template := newTemplate("eng", "Engineering")

		require.NoError(t, repo.Upsert(ctx, template))

		got, err := repo.FindBySlug(ctx, "eng")
		require.NoError(t, err)
		assert.Equal(t, "eng", got.Slug)
		assert.Equal(t, "Engineering", got.Name)
		assert.False(t, got.ID.IsZero(), "Upsert should populate an ID")
		assert.False(t, got.CreatedAt.IsZero())
		assert.False(t, got.UpdatedAt.IsZero())
	})

	t.Run("upsert_is_idempotent_by_slug", func(t *testing.T) {
		template := newTemplate("idem", "First name")
		require.NoError(t, repo.Upsert(ctx, template))

		updated := newTemplate("idem", "Renamed")
		require.NoError(t, repo.Upsert(ctx, updated))

		all, err := repo.FindAll(ctx)
		require.NoError(t, err)
		matched := 0
		for _, tmpl := range all {
			if tmpl.Slug == "idem" {
				matched++
				assert.Equal(t, "Renamed", tmpl.Name)
			}
		}
		assert.Equal(t, 1, matched, "upsert must not duplicate templates with the same slug")
	})

	t.Run("find_by_unknown_slug_returns_no_documents", func(t *testing.T) {
		_, err := repo.FindBySlug(ctx, "nonexistent-slug")
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})

	t.Run("find_by_unknown_id_returns_no_documents", func(t *testing.T) {
		_, err := repo.FindByID(ctx, primitive.NewObjectID())
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})
}
