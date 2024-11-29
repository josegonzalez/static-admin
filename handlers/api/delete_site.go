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

// NewDeleteSiteHandler creates a new handler for the site deletion endpoint
func NewDeleteSiteHandler(config config.Config) (DeleteSiteHandler, error) {
	return DeleteSiteHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// DeleteSiteHandler handles the site deletion request
type DeleteSiteHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h DeleteSiteHandler) GroupRegister(r *gin.RouterGroup) {
	r.DELETE("/sites/:id", h.handler)
	r.OPTIONS("/sites/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the DELETE request for site deletion
func (h DeleteSiteHandler) handler(c *gin.Context) {
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

	// Delete site (only if it belongs to the user)
	result := h.Database.Where("id = ? AND user_id = ?", c.Param("id"), uint(userID)).Delete(&database.Site{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete site",
		})
		return
	}

	c.Status(http.StatusOK)
}
