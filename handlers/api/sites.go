package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/middleware"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SiteResponse represents a site in the JSON response
type SiteResponse struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	RepositoryURL string `json:"url"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

// NewSitesHandler creates a new handler for the sites endpoint
func NewSitesHandler(config config.Config) (SitesHandler, error) {
	return SitesHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// SitesHandler handles the sites request
type SitesHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h SitesHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/sites", h.handler)
	r.OPTIONS("/sites", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the GET request for sites
func (h SitesHandler) handler(c *gin.Context) {
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	// Fetch sites for user
	var sites []database.Site
	if err := h.Database.Where("user_id = ?", user.ID).Find(&sites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sites",
		})
		return
	}

	// Convert to response format
	response := make([]SiteResponse, len(sites))
	for i, site := range sites {
		parts := strings.Split(site.RepositoryURL, "/")
		repositoryName := parts[len(parts)-1]
		response[i] = SiteResponse{
			ID:            site.ID,
			UserID:        site.UserID,
			RepositoryURL: site.RepositoryURL,
			Name:          repositoryName,
			Description:   site.Description,
			DefaultBranch: site.DefaultBranch,
			Private:       site.Private,
		}
	}

	c.JSON(http.StatusOK, response)
}
