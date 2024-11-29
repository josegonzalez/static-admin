package database

import (
	"gorm.io/gorm"
)

// GitHubAuth represents GitHub authentication data for a user
type GitHubAuth struct {
	gorm.Model
	UserID      uint   `gorm:"uniqueIndex"`
	Login       string `gorm:"not null"`
	Name        string `gorm:"not null"`
	URL         string `gorm:"not null"`
	AccessToken string `gorm:"not null"`
}
