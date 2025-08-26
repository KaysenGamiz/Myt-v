package db

import (
	"log"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(path string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB could not be opened: %v", err)
	}

	// Migrate schema
	if err := DB.AutoMigrate(&Movie{}); err != nil {
		log.Fatalf("schema could not be migrated: %v", err)
	}
}
