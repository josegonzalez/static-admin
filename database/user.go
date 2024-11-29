package database

import (
	"gorm.io/gorm"
)

// User model
type User struct {
	gorm.Model
	Name     string `gorm:"size:100;not null"`
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}
