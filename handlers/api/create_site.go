package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// CreateSiteRequest represents the JSON data for creating a new site
type CreateSiteRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

// NewCreateSiteHandler creates a new handler for the site creation endpoint
func NewCreateSiteHandler(config config.Config) (CreateSiteHandler, error) {
	return CreateSiteHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// CreateSiteHandler handles the site creation request
type CreateSiteHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h CreateSiteHandler) GroupRegister(r *gin.RouterGroup) {
	r.PUT("/sites", h.handler)
}

// handler handles the PUT request for site creation
func (h CreateSiteHandler) handler(c *gin.Context) {
	var req CreateSiteRequest
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

	// Extract bearer token
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing bearer token",
		})
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	// Extract user ID from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process token",
		})
		return
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID in token",
		})
		return
	}

	// Check if site already exists
	var existingSite database.Site
	result := h.Database.Where("user_id = ? AND repository_url = ?", uint(userID), req.RepositoryURL).First(&existingSite)
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

	// Create new site
	site := database.Site{
		UserID:        uint(userID),
		RepositoryURL: req.RepositoryURL,
		Description:   req.Description,
		DefaultBranch: req.DefaultBranch,
		Private:       req.Private,
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
