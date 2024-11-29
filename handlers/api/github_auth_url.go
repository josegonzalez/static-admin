package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/middleware"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// GitHubAuthURLResponse represents the JSON response containing the GitHub auth URL
type GitHubAuthURLResponse struct {
	URL string `json:"url"`
}

// NewGitHubAuthURLHandler creates a new handler for the GitHub auth URL endpoint
func NewGitHubAuthURLHandler(config config.Config) (GitHubAuthURLHandler, error) {
	return GitHubAuthURLHandler{
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// GitHubAuthURLHandler handles the GitHub auth URL request
type GitHubAuthURLHandler struct {
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h GitHubAuthURLHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/github/auth-url", h.handler)
	r.OPTIONS("/github/auth-url", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the GET request for GitHub auth URL
func (h GitHubAuthURLHandler) handler(c *gin.Context) {
	// Extract bearer token
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing bearer token",
		})
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token",
		})
		return
	}

	// Get GitHub auth URL
	url := middleware.GetLoginURL(c, tokenString)
	c.JSON(http.StatusOK, GitHubAuthURLResponse{
		URL: url,
	})
}
