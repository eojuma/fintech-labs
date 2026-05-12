package db

import (
	"fintech-labs/backend/models"
	"log"
	"os" // Added this to read environment variables

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// 1. Get the path from the environment variable "DATABASE_PATH"
	dbPath := os.Getenv("DATABASE_PATH")

	// 2. If it's empty (like on your local laptop), use the default name
	if dbPath == "" {
		dbPath = "transaction.db"
	}

	var err error
	// 3. Use the dbPath variable here instead of the hardcoded string
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.Account{}, &models.Transaction{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
}