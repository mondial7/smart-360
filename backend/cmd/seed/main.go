package main

import (
	"log"
	"smart360/database"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🌱 Smart 360 Development Database Seeder")
	log.Println("=========================================")
	log.Println("")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// Initialize database
	database.InitDB()

	// Run comprehensive dev seed
	database.SeedDevData()

	log.Println("")
	log.Println("✅ Database seeding complete!")
	log.Println("")
	log.Println("🎯 Next steps:")
	log.Println("   1. Start backend: go run main.go")
	log.Println("   2. Login as admin: curl 'http://localhost:8080/api/auth/dev-login?email=admin@example.com'")
	log.Println("   3. Open the application and test the AI consolidation feature!")
	log.Println("")
}
