package database

import (
	"errors"

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

// GetSite retrieves the site from the database
func GetSite(db *gorm.DB, siteID string, user User) (Site, error) {
	var site Site
	if err := db.Where("id = ? AND user_id = ?", siteID, user.ID).First(&site).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return site, errors.New("site not found")
		}

		return site, errors.New("failed to fetch site details")
	}

	return site, nil
}
