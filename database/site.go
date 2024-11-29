package database

import (
	"gorm.io/gorm"
)

// Site represents a configured GitHub repository
type Site struct {
	gorm.Model
	UserID        uint   `gorm:"not null;index:idx_user_repo,priority:1"`
	RepositoryURL string `gorm:"not null;index:idx_user_repo,priority:2"`
	Description   string
	DefaultBranch string `gorm:"not null"`
	Private       bool   `gorm:"not null"`
}
