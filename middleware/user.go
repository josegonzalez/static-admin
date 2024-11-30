package middleware

import (
	"net/http"
	"slices"
	"static-admin/database"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// UserContext is the context for the user
type UserContext struct {
	// GitHubAuth is the GitHub authentication details for the user
	GitHubAuth *database.GitHubAuth

	// JWTToken is the JWT token for the user
	JWTToken string

	// User is the user that is currently logged in
	User database.User
}

type UserMiddleware struct {
	Database     *gorm.DB
	JWTSecret    []byte
	IgnoredPaths []string
}

// User is a middleware that authenticates the user and sets the user context
func User(input UserMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		if slices.Contains(input.IgnoredPaths, c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract bearer token
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing bearer token",
			})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return input.JWTSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Extract user ID from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process token claims",
			})
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user ID in token",
			})
			return
		}

		// Fetch user from database
		var user database.User
		if err := input.Database.First(&user, uint(userID)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "User not found",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user details",
			})
			return
		}

		// Fetch GitHub auth data
		var githubAuth database.GitHubAuth
		err = input.Database.Where("user_id = ?", user.ID).First(&githubAuth).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch GitHub authentication details",
			})
			return
		}

		// Store user context
		ctx := UserContext{
			GitHubAuth: &githubAuth,
			JWTToken:   tokenString,
			User:       user,
		}
		c.Set("user", ctx)

		c.Next()
	}
}

// GetUser retrieves the user context from gin context
func GetUser(c *gin.Context) (database.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return database.User{}, false
	}
	userCtx, ok := user.(UserContext)
	return userCtx.User, ok
}

// GetGitHubAuth retrieves the GitHub authentication details from gin context
func GetGitHubAuth(c *gin.Context) (*database.GitHubAuth, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	userCtx, ok := user.(UserContext)
	return userCtx.GitHubAuth, ok
}

// GetJWTToken retrieves the JWT token from gin context
func GetJWTToken(c *gin.Context) (string, bool) {
	user, exists := c.Get("user")
	if !exists {
		return "", false
	}
	userCtx, ok := user.(UserContext)
	if !ok {
		return "", false
	}

	if userCtx.JWTToken == "" {
		return "", false
	}

	return userCtx.JWTToken, true
}
