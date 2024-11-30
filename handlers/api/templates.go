package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TemplateResponse represents a template in the JSON response
type TemplateResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// NewTemplatesHandler creates a new handler for the templates endpoint
func NewTemplatesHandler(config config.Config) (TemplatesHandler, error) {
	return TemplatesHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// TemplatesHandler handles the templates request
type TemplatesHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h TemplatesHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/templates", h.handler)
	r.OPTIONS("/templates", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the GET request for templates
func (h TemplatesHandler) handler(c *gin.Context) {
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	// Fetch templates for user
	var templates []database.Template
	if err := h.Database.Where("user_id = ?", user.ID).Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch templates",
		})
		return
	}

	// Convert to response format
	response := make([]TemplateResponse, len(templates))
	for i, template := range templates {
		response[i] = TemplateResponse{
			ID:   template.ID,
			Name: template.Name,
		}
	}

	c.JSON(http.StatusOK, response)
}
