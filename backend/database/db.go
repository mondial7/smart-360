package database

import (
	"log"
	"smart360/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("smart360.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.FeedbackRound{}, &models.RoundReviewer{}, &models.Submission{}, &models.Consolidation{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	DB = db
	return db
}

func GetDB() *gorm.DB {
	return DB
}
