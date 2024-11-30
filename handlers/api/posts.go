package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/github"
	"static-admin/middleware"
	"strings"

	"github.com/jxskiss/base62"

	"github.com/gin-gonic/gin"
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
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	githubAuth, exists := middleware.GetGitHubAuth(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "GitHub authentication required",
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
	site, err := database.GetSite(h.Database, siteID, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
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
