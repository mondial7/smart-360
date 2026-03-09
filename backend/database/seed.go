package database

import (
	"context"
	"log"
	"smart360/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func SeedData() {
	db := GetDB()
	ctx := context.Background()

	// Check if we already have users
	count, err := db.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("Error counting users: %v", err)
		return
	}
	if count > 1 { // 1 because we might have the dev user already
		return
	}

	log.Println("Seeding development data...")

	now := time.Now()
	lastWeek := now.AddDate(0, 0, -7)
	twoDaysAgo := now.AddDate(0, 0, -2)

	users := []models.User{
		{
			Email:     "alice@example.com",
			Name:      "Alice Johnson",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: &lastWeek,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "bob@example.com",
			Name:      "Bob Smith",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: &twoDaysAgo,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "carol@example.com",
			Name:      "Carol Williams",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: &now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "david@example.com",
			Name:      "David Brown",
			PhotoURL:  "",
			Role:      models.RoleMember,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "emma@example.com",
			Name:      "Emma Davis",
			PhotoURL:  "",
			Role:      models.RoleAdmin,
			LastLogin: &now,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, user := range users {
		// Check if user already exists
		var existing models.User
		err := db.Collection("users").FindOne(ctx, bson.M{"email": user.Email}).Decode(&existing)
		if err != nil && err == mongo.ErrNoDocuments {
			// User doesn't exist, create it
			_, err := db.Collection("users").InsertOne(ctx, user)
			if err != nil {
				log.Printf("Failed to seed user %s: %v", user.Email, err)
			} else {
				log.Printf("Seeded user: %s", user.Name)
			}
		}
	}

	log.Println("Seeding complete!")
}
