package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewTemplateDeleteHandler creates a new handler for the template deletion endpoint
func NewTemplateDeleteHandler(config config.Config) (TemplateDeleteHandler, error) {
	return TemplateDeleteHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// TemplateDeleteHandler handles the template deletion request
type TemplateDeleteHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h TemplateDeleteHandler) GroupRegister(r *gin.RouterGroup) {
	r.DELETE("/templates/:templateId", h.handler)
}

// handler handles the DELETE request for template deletion
func (h TemplateDeleteHandler) handler(c *gin.Context) {
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	templateID := c.Param("templateId")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Template ID is required",
		})
		return
	}

	// Start a transaction
	err := h.Database.Transaction(func(tx *gorm.DB) error {
		// Verify template exists and is owned by user
		var template database.Template
		if err := tx.Where("id = ? AND user_id = ?", templateID, user.ID).First(&template).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return err
			}
			return err
		}

		// Delete template fields first
		if err := tx.Where("template_id = ?", template.ID).Delete(&database.TemplateField{}).Error; err != nil {
			return err
		}

		// Delete template
		if err := tx.Delete(&template).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Template not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete template",
		})
		return
	}

	c.Status(http.StatusOK)
}
