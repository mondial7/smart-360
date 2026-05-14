package repositories

import (
	"context"
	"smart360/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoTemplateRepository implements TemplateRepository using MongoDB.
type MongoTemplateRepository struct {
	db *mongo.Database
}

func NewMongoTemplateRepository(db *mongo.Database) *MongoTemplateRepository {
	return &MongoTemplateRepository{db: db}
}

func (r *MongoTemplateRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error) {
	var t models.Template
	err := r.db.Collection("templates").FindOne(ctx, bson.M{"_id": id}).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *MongoTemplateRepository) FindBySlug(ctx context.Context, slug string) (*models.Template, error) {
	var t models.Template
	err := r.db.Collection("templates").FindOne(ctx, bson.M{"slug": slug}).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *MongoTemplateRepository) FindAll(ctx context.Context) ([]models.Template, error) {
	cursor, err := r.db.Collection("templates").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []models.Template
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

// Upsert creates or replaces a template by slug. Used both by admin-facing
// mutations and by the idempotent default-template seeder.
func (r *MongoTemplateRepository) Upsert(ctx context.Context, template *models.Template) error {
	if template.Slug == "" {
		return mongo.ErrEmptySlice // surface a clear failure rather than upserting an unslugged record
	}
	now := time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	template.UpdatedAt = now

	// We replace the whole document so the latest seed wins on every boot.
	opts := options.Replace().SetUpsert(true)
	_, err := r.db.Collection("templates").ReplaceOne(
		ctx,
		bson.M{"slug": template.Slug},
		template,
		opts,
	)
	return err
}
