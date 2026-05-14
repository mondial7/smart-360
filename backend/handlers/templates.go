package handlers

import (
	"context"
	"errors"
	"net/http"
	"smart360/database"
	"smart360/models"
	"smart360/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ListTemplates returns every available round template. Any authenticated user
// can read templates — the question wording is what reviewers see, so there's
// nothing sensitive to gate.
func ListTemplates(c *gin.Context) {
	repo := repositories.NewMongoTemplateRepository(database.GetDB())
	templates, err := repo.FindAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}
	c.JSON(http.StatusOK, templates)
}

// GetTemplate resolves a template by either ObjectID hex OR slug — the submit
// form fetches by ID embedded in the round, while wizards/clients can use the
// stable slug. One endpoint, two valid keys, keeps the API surface small.
func GetTemplate(c *gin.Context) {
	idOrSlug := c.Param("idOrSlug")
	repo := repositories.NewMongoTemplateRepository(database.GetDB())
	ctx := context.Background()

	if oid, err := primitive.ObjectIDFromHex(idOrSlug); err == nil {
		t, err := repo.FindByID(ctx, oid)
		if err == nil {
			c.JSON(http.StatusOK, t)
			return
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch template"})
			return
		}
	}

	t, err := repo.FindBySlug(ctx, idOrSlug)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch template"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// resolveTemplate returns the template referenced by templateID, falling back
// to the default-slug template when the ID is zero (e.g. legacy rounds from
// before configurable templates). Returns nil and an error only if even the
// default can't be loaded.
func resolveTemplate(ctx context.Context, repo repositories.TemplateRepository, templateID primitive.ObjectID) (*models.Template, error) {
	if !templateID.IsZero() {
		t, err := repo.FindByID(ctx, templateID)
		if err == nil {
			return t, nil
		}
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		// fall through to default if the referenced template was deleted
	}
	return repo.FindBySlug(ctx, models.DefaultTemplateSlug)
}

// resolveTemplateIDForCreate validates the incoming templateID against the
// templates collection. Zero means "use the bundled default" — we resolve it
// to the default's actual ID so every round in the DB carries an explicit
// template reference. Returns a user-friendly error if the ID is unknown.
func resolveTemplateIDForCreate(ctx context.Context, templateID primitive.ObjectID) (primitive.ObjectID, error) {
	repo := repositories.NewMongoTemplateRepository(database.GetDB())
	if templateID.IsZero() {
		t, err := repo.FindBySlug(ctx, models.DefaultTemplateSlug)
		if err != nil {
			return primitive.NilObjectID, errors.New("default template is not seeded — re-run startup")
		}
		return t.ID, nil
	}
	t, err := repo.FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return primitive.NilObjectID, errors.New("unknown templateId")
		}
		return primitive.NilObjectID, err
	}
	return t.ID, nil
}
