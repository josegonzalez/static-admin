package api

import (
	"fmt"
	"net/http"
	"static-admin/blocks"
	"static-admin/config"
	"static-admin/database"
	"static-admin/github"
	"static-admin/markdown"
	"static-admin/middleware"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jxskiss/base62"
	"gorm.io/gorm"
)

// PostContentResponse represents the JSON response for a post's content
type PostContentResponse struct {
	ID          string                      `json:"id"`
	Path        string                      `json:"path"`
	Frontmatter []markdown.FrontmatterField `json:"frontmatter"`
	Blocks      []blocks.Block              `json:"blocks"`
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
	site, err := database.GetSite(h.Database, siteID, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
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
