package database

import (
	"time"

	"gorm.io/gorm"
)

// User model
type User struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"size:100;not null"`
	Email     string         `gorm:"unique;not null"`
	Password  string         `gorm:"not null"`
	CreatedAt time.Time      // Tracks when the record was created
	UpdatedAt time.Time      // Tracks when the record was last modified
	DeletedAt gorm.DeletedAt `gorm:"index"` // Tracks soft deletion timestamp
}
