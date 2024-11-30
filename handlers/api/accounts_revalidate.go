package api

import (
	"net/http"
	"time"

	"static-admin/config"
	"static-admin/middleware"

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
	user, ok := middleware.GetUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	githubAuth, exists := middleware.GetGitHubAuth(c)
	// Generate new token with updated claims
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID,
		"email":       user.Email,
		"github_auth": exists && githubAuth.AccessToken != "",
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
