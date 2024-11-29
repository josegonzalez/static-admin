package api

import (
	"net/http"
	"strings"
	"time"

	"static-admin/config"
	"static-admin/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// RevalidateResponse represents the JSON response for a successful token revalidation
type RevalidateResponse struct {
	Token string `json:"token"`
}

// NewRevalidateHandler creates a new handler for the token revalidation endpoint
func NewRevalidateHandler(config config.Config) (RevalidateHandler, error) {
	return RevalidateHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// RevalidateHandler handles the token revalidation request
type RevalidateHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h RevalidateHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/auth/revalidate", h.handler)
	r.OPTIONS("/auth/revalidate", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the GET request for token revalidation
func (h RevalidateHandler) handler(c *gin.Context) {
	// Extract bearer token
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing bearer token",
		})
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate existing token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	// Extract claims from existing token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process token",
		})
		return
	}

	// fetch associated GithubAuth record
	var githubAuth database.GitHubAuth
	if err := h.Database.Where("user_id = ?", claims["user_id"]).First(&githubAuth).Error; err != nil {
		claims["github_auth"] = false
	} else {
		claims["github_auth"] = githubAuth.AccessToken != ""
	}

	// Generate new token with updated claims
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     claims["user_id"],
		"email":       claims["email"],
		"github_auth": claims["github_auth"],
		"exp":         time.Now().Add(time.Hour * 72).Unix(), // 3 days expiration
	})

	// Sign the token
	signedToken, err := newToken.SignedString(h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, RevalidateResponse{
		Token: signedToken,
	})
}
