package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/github"
	"strings"

	"github.com/jxskiss/base62"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// generatePostID creates a base62-encoded hash of the file path
func generatePostID(path string) string {
	return base62.EncodeToString([]byte(path))
}

// PostResponse represents a post in the JSON response
type PostResponse struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// NewPostsHandler creates a new handler for the posts endpoint
func NewPostsHandler(config config.Config) (PostsHandler, error) {
	return PostsHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// PostsHandler handles the posts request
type PostsHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h PostsHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/sites/:siteId/posts", h.handler)
	r.OPTIONS("/sites/:siteId/posts", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the GET request for posts
func (h PostsHandler) handler(c *gin.Context) {
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

	// Get site ID from URL
	siteID := c.Param("siteId")
	if siteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Site ID is required",
		})
		return
	}

	// Fetch site details
	var site database.Site
	if err := h.Database.Where("id = ? AND user_id = ?", siteID, uint(userID)).First(&site).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Site not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch site details",
		})
		return
	}

	// Get GitHub auth details
	var githubAuth database.GitHubAuth
	if err := h.Database.Where("user_id = ?", uint(userID)).First(&githubAuth).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "GitHub authentication required",
		})
		return
	}

	// Extract owner and repo from repository URL
	// Format: https://github.com/owner/repo
	urlParts := strings.Split(site.RepositoryURL, "/")
	if len(urlParts) < 5 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid repository URL",
		})
		return
	}
	owner := urlParts[len(urlParts)-2]
	repo := urlParts[len(urlParts)-1]

	// Fetch files from GitHub
	files, err := github.FetchRepoFiles(github.FetchRepoFilesInput{
		Owner: owner,
		Repo:  repo,
		Path:  "_posts",
		Token: githubAuth.AccessToken,
		Ref:   site.DefaultBranch,
		Type:  "file",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch posts from GitHub",
		})
		return
	}

	// Convert to response format
	response := make([]PostResponse, len(files))
	for i, file := range files {
		response[i] = PostResponse{
			ID:   generatePostID(file.Path),
			Path: file.Path,
		}
	}

	c.JSON(http.StatusOK, response)
}
