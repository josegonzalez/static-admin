package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/github"
	"static-admin/middleware"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SiteCreateRequest represents the JSON data for creating a new site
type SiteCreateRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

// NewSiteCreateHandler creates a new handler for the site creation endpoint
func NewSiteCreateHandler(config config.Config) (SiteCreateHandler, error) {
	return SiteCreateHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// SiteCreateHandler handles the site creation request
type SiteCreateHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h SiteCreateHandler) GroupRegister(r *gin.RouterGroup) {
	r.PUT("/sites", h.handler)
}

// handler handles the PUT request for site creation
func (h SiteCreateHandler) handler(c *gin.Context) {
	user, ok := middleware.GetUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	githubAuth, exists := middleware.GetGitHubAuth(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	var req SiteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Validate URL format
	if !strings.HasPrefix(req.RepositoryURL, "https://github.com/") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid GitHub repository URL",
		})
		return
	}

	// Check if site already exists
	var existingSite database.Site
	result := h.Database.Where("user_id = ? AND repository_url = ?", user.ID, req.RepositoryURL).First(&existingSite)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Site already exists",
		})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check for existing site",
		})
		return
	}

	// fetch repository info
	repo, err := github.FetchRepository(github.FetchRepositoryInput{
		RepositoryURL: req.RepositoryURL,
		Token:         githubAuth.AccessToken,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch repository",
		})
		return
	}

	// Create new site
	site := database.Site{
		UserID:        user.ID,
		RepositoryURL: repo.HtmlURL,
		Description:   repo.Description,
		DefaultBranch: repo.DefaultBranch,
		Private:       repo.Private,
	}

	if err := h.Database.Create(&site).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create site",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":  site.ID,
		"url": site.RepositoryURL,
	})
}
