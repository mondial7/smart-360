package repositories

import (
	"context"
	"smart360/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoRoundRepository implements RoundRepository using MongoDB
type MongoRoundRepository struct {
	db *mongo.Database
}

// NewMongoRoundRepository creates a new MongoDB-backed round repository
func NewMongoRoundRepository(db *mongo.Database) *MongoRoundRepository {
	return &MongoRoundRepository{db: db}
}

// FindByID finds a round by ID
func (r *MongoRoundRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.FeedbackRound, error) {
	var round models.FeedbackRound
	err := r.db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": id}).Decode(&round)
	if err != nil {
		return nil, err
	}
	return &round, nil
}

// Create creates a new round
func (r *MongoRoundRepository) Create(ctx context.Context, round *models.FeedbackRound) error {
	// Initialize reviewers array if nil
	if round.Reviewers == nil {
		round.Reviewers = []models.RoundReviewer{}
	}

	_, err := r.db.Collection("feedback_rounds").InsertOne(ctx, round)
	return err
}

// Update updates an existing round
func (r *MongoRoundRepository) Update(ctx context.Context, round *models.FeedbackRound) error {
	_, err := r.db.Collection("feedback_rounds").ReplaceOne(
		ctx,
		bson.M{"_id": round.ID},
		round,
	)
	return err
}

// UpdateStatus updates a round's status
func (r *MongoRoundRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.RoundStatus) error {
	_, err := r.db.Collection("feedback_rounds").UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}

// FindBySubjectID finds all rounds for a given subject
func (r *MongoRoundRepository) FindBySubjectID(ctx context.Context, subjectID primitive.ObjectID) ([]models.FeedbackRound, error) {
	cursor, err := r.db.Collection("feedback_rounds").Find(ctx, bson.M{"subject_id": subjectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rounds []models.FeedbackRound
	if err = cursor.All(ctx, &rounds); err != nil {
		return nil, err
	}

	return rounds, nil
}

// FindAll returns all rounds
func (r *MongoRoundRepository) FindAll(ctx context.Context) ([]models.FeedbackRound, error) {
	cursor, err := r.db.Collection("feedback_rounds").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rounds []models.FeedbackRound
	if err = cursor.All(ctx, &rounds); err != nil {
		return nil, err
	}

	return rounds, nil
}

// AddReviewer adds a reviewer to a round
func (r *MongoRoundRepository) AddReviewer(ctx context.Context, roundID primitive.ObjectID, reviewer models.RoundReviewer) error {
	// Set the round ID
	reviewer.RoundID = roundID

	// Generate ID if not set
	if reviewer.ID.IsZero() {
		reviewer.ID = primitive.NewObjectID()
	}

	_, err := r.db.Collection("feedback_rounds").UpdateOne(
		ctx,
		bson.M{"_id": roundID},
		bson.M{"$addToSet": bson.M{"reviewers": reviewer}},
	)
	return err
}

// RemoveReviewer removes a reviewer from a round
func (r *MongoRoundRepository) RemoveReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) error {
	_, err := r.db.Collection("feedback_rounds").UpdateOne(
		ctx,
		bson.M{"_id": roundID},
		bson.M{"$pull": bson.M{"reviewers": bson.M{"reviewer_id": reviewerID}}},
	)
	return err
}

// GetReviewers returns all reviewers for a round
func (r *MongoRoundRepository) GetReviewers(ctx context.Context, roundID primitive.ObjectID) ([]models.RoundReviewer, error) {
	round, err := r.FindByID(ctx, roundID)
	if err != nil {
		return nil, err
	}

	if round.Reviewers == nil {
		return []models.RoundReviewer{}, nil
	}

	return round.Reviewers, nil
}
