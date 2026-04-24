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

	return nil
}
