package auth

import (
	"context"
	"fmt"
	"net/http"

	"static-admin/config"
	"static-admin/database"
	"static-admin/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"gorm.io/gorm"
)

// NewGithubCallbackHandler creates a new handler for the github callback page
func NewGithubCallbackHandler(config config.Config) (GithubCallbackHandler, error) {
	return GithubCallbackHandler{
		JWTSecret: []byte(config.JWTSecret),
		Database:  config.Database,
	}, nil
}

// GithubCallbackHandler handles the github callback request
type GithubCallbackHandler struct {
	JWTSecret []byte
	Database  *gorm.DB
}

// GroupRegister registers the handler with the given router
func (h GithubCallbackHandler) GroupRegister(auth *gin.RouterGroup) {
	auth.GET("/auth/github/callback", h.handler)
}

// handler handles the request for the page
func (h GithubCallbackHandler) handler(c *gin.Context) {
	// Check for JWT token in session
	tokenString := c.Query("state")
	if tokenString == "" {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Validate JWT token and extract user ID
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("invalid token claims"))
		return
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("invalid user_id in token"))
		return
	}

	stdctx := context.Background()
	tok, err := middleware.GithubConfig.Exchange(stdctx, c.Query("code"))
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to do exchange: %w", err))
		return
	}

	client := github.NewClient(middleware.GithubConfig.Client(stdctx, tok))
	user, _, err := client.Users.Get(stdctx, "")
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to get user: %w", err))
		return
	}

	// Update or create GitHub auth record in database
	githubAuth := database.GitHubAuth{
		UserID:      uint(userID),
		Login:       stringFromPointer(user.Login),
		Name:        stringFromPointer(user.Name),
		URL:         stringFromPointer(user.URL),
		AccessToken: tok.AccessToken,
	}

	result := h.Database.Where("user_id = ?", uint(userID)).FirstOrCreate(&githubAuth)
	if result.Error != nil {
		glog.Errorf("Failed to save GitHub auth: %v", result.Error)
		_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save auth data"))
		return
	}

	c.Redirect(http.StatusFound, "http://localhost:3000/dashboard?refetch-token=true")
}

func stringFromPointer(strPtr *string) (res string) {
	if strPtr == nil {
		res = ""
		return res
	}
	res = *strPtr
	return res
}
