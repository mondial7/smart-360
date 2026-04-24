package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestMongoDB holds the in-memory MongoDB server and client for testing
type TestMongoDB struct {
	Server *memongo.Server
	Client *mongo.Client
	DB     *mongo.Database
}

// SetupTestMongoDB creates an in-memory MongoDB instance for testing
// The MongoDB server is automatically stopped when the test completes via t.Cleanup()
func SetupTestMongoDB(t *testing.T) *TestMongoDB {
	// Start in-memory MongoDB server (use 6.0.0 for ARM64/Apple Silicon support)
	mongoServer, err := memongo.Start("6.0.0")
	require.NoError(t, err, "Failed to start in-memory MongoDB")

	// Connect to the in-memory MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoServer.URI()))
	require.NoError(t, err, "Failed to connect to in-memory MongoDB")

	// Ping to verify connection
	err = client.Ping(ctx, nil)
	require.NoError(t, err, "Failed to ping in-memory MongoDB")

	// Create test database
	db := client.Database("test_smart360")

	// Register cleanup to stop server and disconnect client when test completes
	t.Cleanup(func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Logf("Failed to disconnect from MongoDB: %v", err)
		}
		mongoServer.Stop()
	})

	return &TestMongoDB{
		Server: mongoServer,
		Client: client,
		DB:     db,
	}
}
