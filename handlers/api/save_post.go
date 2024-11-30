package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"static-admin/blocks"
	"static-admin/config"
	"static-admin/database"
	"static-admin/github"
	"static-admin/markdown"
	"static-admin/middleware"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// SavePostRequest represents the JSON request for saving a post's content
type SavePostRequest struct {
	ID          string                      `json:"id"`
	Path        string                      `json:"path"`
	Frontmatter []markdown.FrontmatterField `json:"frontmatter"`
	Blocks      []blocks.Block              `json:"blocks"`
}

// SavePostResponse represents the JSON response for saving a post's content
type SavePostResponse struct {
	Message  string          `json:"message"`
	Request  SavePostRequest `json:"content"`
	Path     string          `json:"path"`
	Markdown string          `json:"markdown"`
	PRURL    string          `json:"pr_url"`
}

// NewSavePostHandler creates a new handler for saving post content
func NewSavePostHandler(config config.Config) (SavePostHandler, error) {
	return SavePostHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// SavePostHandler handles the save post request
type SavePostHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h SavePostHandler) GroupRegister(r *gin.RouterGroup) {
	r.POST("/sites/:siteId/posts/:postId", h.handler)
}

// handler handles the POST request for saving post content
func (h SavePostHandler) handler(c *gin.Context) {
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

	// Verify site ownership
	site, err := database.GetSite(h.Database, siteID, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse request body
	var req SavePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// decode the id into a path
	path, err := fromBase62(req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to decode post ID",
		})
		return
	}

	// Generate markdown content
	frontmatterYaml, err := markdown.FrontmatterFieldToYaml(req.Frontmatter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate frontmatter",
		})
		return
	}

	contentMarkdown, err := blocks.ParseBlocksToMarkdown(req.Blocks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate markdown",
		})
		return
	}

	fullMarkdown := frontmatterYaml + "\n" + contentMarkdown + "\n"

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

	fileName := filepath.Base(path)
	branchName := fmt.Sprintf("update-%s", slug.Make(fileName))

	err = github.CreateBranchAndUpdateFile(github.CreateBranchAndUpdateFileInput{
		Owner:      owner,
		Repo:       repo,
		Path:       path,
		Content:    fullMarkdown,
		Branch:     branchName,
		BaseBranch: site.DefaultBranch,
		CommitMsg:  fmt.Sprintf("Update %s", path),
		Token:      githubAuth.AccessToken,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create branch and update file: %v", err),
		})
		return
	}

	prNumber, err := github.CreatePullRequestIfNecessary(github.CreatePullRequestIfNecessaryInput{
		Owner:      owner,
		Repo:       repo,
		Branch:     branchName,
		BaseBranch: site.DefaultBranch,
		Title:      fmt.Sprintf("Update %s", fileName),
		Body:       fmt.Sprintf("Updates content for %s", path),
		Token:      githubAuth.AccessToken,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create pull request: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, SavePostResponse{
		Message:  "Created pull request for changes",
		Request:  req,
		Path:     path,
		Markdown: fullMarkdown,
		PRURL:    fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, prNumber),
	})
}
