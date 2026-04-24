package repositories

import (
	"context"
	"smart360/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoSubmissionRepository implements SubmissionRepository using MongoDB
type MongoSubmissionRepository struct {
	db *mongo.Database
}

// NewMongoSubmissionRepository creates a new MongoDB-backed submission repository
func NewMongoSubmissionRepository(db *mongo.Database) *MongoSubmissionRepository {
	return &MongoSubmissionRepository{db: db}
}

// FindByRoundID finds all submissions for a given round
func (r *MongoSubmissionRepository) FindByRoundID(ctx context.Context, roundID primitive.ObjectID) ([]models.Submission, error) {
	cursor, err := r.db.Collection("submissions").Find(ctx, bson.M{"round_id": roundID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var submissions []models.Submission
	if err = cursor.All(ctx, &submissions); err != nil {
		return nil, err
	}

	return submissions, nil
}

// Create creates a new submission
func (r *MongoSubmissionRepository) Create(ctx context.Context, submission *models.Submission) error {
	_, err := r.db.Collection("submissions").InsertOne(ctx, submission)
	return err
}

// Update updates an existing submission
func (r *MongoSubmissionRepository) Update(ctx context.Context, submission *models.Submission) error {
	_, err := r.db.Collection("submissions").ReplaceOne(
		ctx,
		bson.M{"_id": submission.ID},
		submission,
	)
	return err
}

// FindByRoundAndReviewer finds a submission by round and reviewer
func (r *MongoSubmissionRepository) FindByRoundAndReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) (*models.Submission, error) {
	var submission models.Submission
	err := r.db.Collection("submissions").FindOne(ctx, bson.M{
		"round_id":    roundID,
		"reviewer_id": reviewerID,
	}).Decode(&submission)
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// CountByRoundAndReviewer counts submissions by round and reviewer
func (r *MongoSubmissionRepository) CountByRoundAndReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) (int64, error) {
	count, err := r.db.Collection("submissions").CountDocuments(ctx, bson.M{
		"round_id":    roundID,
		"reviewer_id": reviewerID,
	})
	return count, err
}

// FindByID finds a submission by ID
func (r *MongoSubmissionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Submission, error) {
	var submission models.Submission
	err := r.db.Collection("submissions").FindOne(ctx, bson.M{"_id": id}).Decode(&submission)
	if err != nil {
		return nil, err
	}
	return &submission, nil
}
