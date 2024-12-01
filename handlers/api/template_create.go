package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TemplateCreateRequest represents the JSON request for creating a template
type TemplateCreateRequest struct {
	Name   string                `json:"name" binding:"required"`
	Fields []TemplateCreateField `json:"fields" binding:"required"`
}

type TemplateCreateField struct {
	Name             string   `json:"name" binding:"required"`
	Type             string   `json:"type" binding:"required"`
	StringValue      string   `json:"stringValue"`
	BoolValue        bool     `json:"boolValue"`
	NumberValue      float64  `json:"numberValue"`
	DateTimeValue    string   `json:"dateTimeValue"`
	StringSliceValue []string `json:"stringSliceValue"`
}

// NewTemplateCreateHandler creates a new handler for template creation
func NewTemplateCreateHandler(config config.Config) (TemplateCreateHandler, error) {
	return TemplateCreateHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// TemplateCreateHandler handles the template creation request
type TemplateCreateHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h TemplateCreateHandler) GroupRegister(r *gin.RouterGroup) {
	r.PUT("/templates", h.handler)
}

// handler handles the PUT request for template creation
func (h TemplateCreateHandler) handler(c *gin.Context) {
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	var req TemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Start a transaction
	err := h.Database.Transaction(func(tx *gorm.DB) error {
		// Create template
		template := database.Template{
			UserID: user.ID,
			Name:   req.Name,
		}
		if err := tx.Create(&template).Error; err != nil {
			return err
		}

		// Create template fields
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create template",
		})
		return
	}

	c.Status(http.StatusCreated)
}
