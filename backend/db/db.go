package db

import (
	"fintech-labs/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("transaction.db"), &gorm.Config{})

	if err !=nil{
		log.Fatal("Failed to connect to database:",err)
	}

	err=DB.AutoMigrate(&models.Account{},&models.Transaction{})

	if err !=nil{
		log.Fatal("Migration failed")
	}
	log.Println("GORM database initialized.")
}
