package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TemplateUpdateRequest represents the JSON request for updating a template
type TemplateUpdateRequest struct {
	Name   string                `json:"name" binding:"required"`
	Fields []TemplateUpdateField `json:"fields" binding:"required"`
}

type TemplateUpdateField struct {
	ID               uint     `json:"id"`
	Name             string   `json:"name" binding:"required"`
	Type             string   `json:"type" binding:"required"`
	StringValue      string   `json:"stringValue"`
	BoolValue        bool     `json:"boolValue"`
	NumberValue      float64  `json:"numberValue"`
	DateTimeValue    string   `json:"dateTimeValue"`
	StringSliceValue []string `json:"stringSliceValue"`
}

// NewTemplateUpdateHandler creates a new handler for template updates
func NewTemplateUpdateHandler(config config.Config) (TemplateUpdateHandler, error) {
	return TemplateUpdateHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// TemplateUpdateHandler handles the template update request
type TemplateUpdateHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h TemplateUpdateHandler) GroupRegister(r *gin.RouterGroup) {
	r.POST("/templates/:templateId", h.handler)
}

// handler handles the POST request for template updates
func (h TemplateUpdateHandler) handler(c *gin.Context) {
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

	var req TemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Start a transaction
	err := h.Database.Transaction(func(tx *gorm.DB) error {
		// Verify template exists and is owned by user
		var template database.Template
		if err := tx.Where("id = ? AND user_id = ?", templateID, user.ID).First(&template).Error; err != nil {
			return err
		}

		// Update template name
		template.Name = req.Name
		if err := tx.Save(&template).Error; err != nil {
			return err
		}

		// Delete existing fields
		if err := tx.Where("template_id = ?", template.ID).Delete(&database.TemplateField{}).Error; err != nil {
			return err
		}

		// Create new fields
		for _, field := range req.Fields {
			templateField := database.TemplateField{
				TemplateID:       template.ID,
				Name:             field.Name,
				Type:             field.Type,
				StringValue:      field.StringValue,
				BoolValue:        field.BoolValue,
				NumberValue:      field.NumberValue,
				StringSliceValue: field.StringSliceValue,
			}
			if err := tx.Create(&templateField).Error; err != nil {
				return err
			}
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
			"error": "Failed to update template",
		})
		return
	}

	c.Status(http.StatusOK)
}
