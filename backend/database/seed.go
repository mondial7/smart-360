package database

import (
	"log"
	"smart360/models"
	"time"
)

func SeedData() {
	db := GetDB()

	// Check if we already have users
	var count int64
	db.Model(&models.User{}).Count(&count)
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
		},
		{
			Email:     "bob@example.com",
			Name:      "Bob Smith",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: &twoDaysAgo,
		},
		{
			Email:     "carol@example.com",
			Name:      "Carol Williams",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: &now,
		},
		{
			Email:     "david@example.com",
			Name:      "David Brown",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: nil,
		},
		{
			Email:     "emma@example.com",
			Name:      "Emma Davis",
			PhotoURL:  "",
			Role:      models.RoleAdmin,
			LastLogin: &now,
		},
	}

	for _, user := range users {
		var existing models.User
		if err := db.Where("email = ?", user.Email).First(&existing).Error; err != nil {
			// User doesn't exist, create it
			if err := db.Create(&user).Error; err != nil {
				log.Printf("Failed to seed user %s: %v", user.Email, err)
			} else {
				log.Printf("Seeded user: %s", user.Name)
			}
		}
	}

	log.Println("Seeding complete!")
}
