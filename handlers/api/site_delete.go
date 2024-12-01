package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewSiteDeleteHandler creates a new handler for the site deletion endpoint
func NewSiteDeleteHandler(config config.Config) (SiteDeleteHandler, error) {
	return SiteDeleteHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// SiteDeleteHandler handles the site deletion request
type SiteDeleteHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h SiteDeleteHandler) GroupRegister(r *gin.RouterGroup) {
	r.DELETE("/sites/:siteId", h.handler)
	r.OPTIONS("/sites/:siteId", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the DELETE request for site deletion
func (h SiteDeleteHandler) handler(c *gin.Context) {
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	siteID := c.Param("siteId")
	_, err := database.GetSite(h.Database, siteID, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Delete site (only if it belongs to the user)
	result := h.Database.Where("id = ? AND user_id = ?", c.Param("siteId"), user.ID).Delete(&database.Site{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete site",
		})
		return
	}

	c.Status(http.StatusOK)
}
