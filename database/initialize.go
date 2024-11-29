package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Initialize creates a connection to the database and creates the users table if it doesn't exist
func Initialize() (*gorm.DB, error) {
	// Connect to SQLite database
	db, err := gorm.Open(sqlite.Open("gorm_example.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	models := []interface{}{
		&User{},
		&GitHubAuth{},
		&Site{},
	}

	// AutoMigrate the schema
	if err := db.AutoMigrate(models...); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
