package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TemplateFieldResponse represents a template field in the JSON response
type TemplateFieldResponse struct {
	ID               uint     `json:"id"`
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	StringValue      string   `json:"stringValue"`
	BoolValue        bool     `json:"boolValue"`
	NumberValue      float64  `json:"numberValue"`
	DateTimeValue    string   `json:"dateTimeValue"`
	StringSliceValue []string `json:"stringSliceValue"`
}

// SingleTemplateResponse represents a single template with its fields
type SingleTemplateResponse struct {
	ID     uint                    `json:"id"`
	Name   string                  `json:"name"`
	Fields []TemplateFieldResponse `json:"fields"`
}

// NewTemplateHandler creates a new handler for the single template endpoint
func NewTemplateHandler(config config.Config) (TemplateHandler, error) {
	return TemplateHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// TemplateHandler handles the single template request
type TemplateHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h TemplateHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/templates/:templateId", h.handler)
	r.OPTIONS("/templates/:templateId", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the GET request for a single template
func (h TemplateHandler) handler(c *gin.Context) {
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

	// Fetch template and verify ownership
	var template database.Template
	if err := h.Database.Where("id = ? AND user_id = ?", templateID, user.ID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Template not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch template",
		})
		return
	}

	// Fetch template fields
	var fields []database.TemplateField
	if err := h.Database.Where("template_id = ?", template.ID).Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch template fields",
		})
		return
	}

	// Convert fields to response format
	fieldResponses := make([]TemplateFieldResponse, len(fields))
	for i, field := range fields {
		fieldResponses[i] = TemplateFieldResponse{
			ID:               field.ID,
			Name:             field.Name,
			Type:             field.Type,
			StringValue:      field.StringValue,
			BoolValue:        field.BoolValue,
			NumberValue:      field.NumberValue,
			DateTimeValue:    field.DateTimeValue.Format("2006-01-02T15:04:05Z07:00"),
			StringSliceValue: field.StringSliceValue,
		}
	}

	// Return combined response
	c.JSON(http.StatusOK, SingleTemplateResponse{
		ID:     template.ID,
		Name:   template.Name,
		Fields: fieldResponses,
	})
}
