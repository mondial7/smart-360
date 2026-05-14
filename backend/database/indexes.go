package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndexes creates necessary indexes for all collections
func CreateIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Audit logs indexes
	auditLogIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "round_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "action", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "round_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err := db.Collection("audit_logs").Indexes().CreateMany(ctx, auditLogIndexes)
	if err != nil {
		log.Printf("Failed to create audit_logs indexes: %v", err)
		return err
	}
	log.Println("Created audit_logs indexes")

	// Submissions indexes - compound unique index
	submissionIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "round_id", Value: 1},
				{Key: "reviewer_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err = db.Collection("submissions").Indexes().CreateMany(ctx, submissionIndexes)
	if err != nil {
		log.Printf("Failed to create submissions indexes: %v", err)
		return err
	}
	log.Println("Created submissions indexes")

	// Templates indexes - slug is the public, stable identifier and must be unique.
	templateIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err = db.Collection("templates").Indexes().CreateMany(ctx, templateIndexes); err != nil {
		log.Printf("Failed to create templates indexes: %v", err)
		return err
	}
	log.Println("Created templates indexes")

	// Moderation logs indexes - we read by round to surface the audit trail.
	moderationIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "round_id", Value: 1}, {Key: "moderated_at", Value: -1}}},
		{Keys: bson.D{{Key: "submission_id", Value: 1}}},
	}
	if _, err = db.Collection("moderation_logs").Indexes().CreateMany(ctx, moderationIndexes); err != nil {
		log.Printf("Failed to create moderation_logs indexes: %v", err)
		return err
	}
	log.Println("Created moderation_logs indexes")

	return nil
}
