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

// SiteResponse represents a site in the JSON response
type SiteResponse struct {
	ID            uint   `json:"id"`
	RepositoryURL string `json:"url"`
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

	// Fetch sites for user
	var sites []database.Site
	if err := h.Database.Where("user_id = ?", uint(userID)).Find(&sites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sites",
		})
		return
	}

	// Convert to response format
	response := make([]SiteResponse, len(sites))
	for i, site := range sites {
		response[i] = SiteResponse{
			ID:            site.ID,
			RepositoryURL: site.RepositoryURL,
		}
	}

	c.JSON(http.StatusOK, response)
}
