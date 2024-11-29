package api

import (
	"fmt"
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/github"
	"static-admin/markdown"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jxskiss/base62"
	"gorm.io/gorm"
)

// PostContentResponse represents the JSON response for a post's content
type PostContentResponse struct {
	ID          string                      `json:"id"`
	Path        string                      `json:"path"`
	Frontmatter []markdown.FrontmatterField `json:"frontmatter"`
	Blocks      []markdown.Block            `json:"blocks"`
}

// NewPostHandler creates a new handler for the post content endpoint
func NewPostHandler(config config.Config) (PostHandler, error) {
	return PostHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// PostHandler handles the post content request
type PostHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h PostHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/sites/:siteId/posts/:postId", h.handler)
	r.OPTIONS("/sites/:siteId/posts/:postId", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// fromBase62 converts a base62 string back to a path
func fromBase62(encoded string) (string, error) {
	decoded, err := base62.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base62: %w", err)
	}
	return string(decoded), nil
}

// handler handles the GET request for post content
func (h PostHandler) handler(c *gin.Context) {
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

	// Get site ID and post ID from URL
	siteID := c.Param("siteId")
	postID := c.Param("postId")
	if siteID == "" || postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Site ID and Post ID are required",
		})
		return
	}

	postPath, err := fromBase62(postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to decode post ID",
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
	urlParts := strings.Split(site.RepositoryURL, "/")
	if len(urlParts) < 5 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid repository URL",
		})
		return
	}
	owner := urlParts[len(urlParts)-2]
	repo := urlParts[len(urlParts)-1]

	branch := site.DefaultBranch
	if branch == "" {
		branch = "master"
	}

	// Fetch file content from GitHub
	content, err := github.FetchFileFromGitHub(github.GitHubFileRequest{
		RepoOwner: owner,
		RepoName:  repo,
		FilePath:  postPath,
		Branch:    branch,
		Token:     githubAuth.AccessToken,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch file content",
		})
		return
	}

	// Extract frontmatter and parse markdown
	frontmatter, markdownContent, err := markdown.ExtractFrontMatter([]byte(content))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to extract frontmatter",
		})
		return
	}

	// Parse markdown into blocks
	blocks, err := markdown.ParseMarkdownToBlocks(markdownContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse markdown",
		})
		return
	}

	c.JSON(http.StatusOK, PostContentResponse{
		ID:          postID,
		Path:        postPath,
		Frontmatter: frontmatter,
		Blocks:      blocks,
	})
}
