package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func InitDB() *mongo.Database {
	// Get MongoDB configuration
	uri := getEnv("MONGODB_URI", "mongodb://admin:password123@localhost:27017")
	dbName := getEnv("MONGODB_DB", "smart360")

	log.Printf("Connecting to MongoDB at %s, database: %s", uri, dbName)

	// Connect to MongoDB
	client, err := mongo.Connect(nil, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	DB = client.Database(dbName)
	log.Printf("Successfully connected to MongoDB!")

	// Create indexes
	if err := CreateIndexes(DB); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	return DB
}

func GetDB() *mongo.Database {
	return DB
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
