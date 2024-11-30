package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/middleware"

	"static-admin/github"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RepositoryResponse represents a GitHub repository in the JSON response
type RepositoryResponse struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	Url           string `json:"url"`
	HtmlURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
}

// NewGitHubRepositoriesHandler creates a new handler for the GitHub repositories endpoint
func NewGitHubRepositoriesHandler(config config.Config) (GitHubRepositoriesHandler, error) {
	return GitHubRepositoriesHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// GitHubRepositoriesHandler handles the GitHub repositories request
type GitHubRepositoriesHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h GitHubRepositoriesHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/github/organizations/:org/repositories", h.handler)
	r.OPTIONS("/github/organizations/:org/repositories", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the GET request for GitHub repositories
func (h GitHubRepositoriesHandler) handler(c *gin.Context) {
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

	var err error
	var allRepos []github.Repository
	orgName := c.Param("org")
	if orgName == githubAuth.Login {
		allRepos, err = github.FetchUserRepositories(github.FetchUserRepositoriesInput{
			Username: githubAuth.Login,
			Token:    githubAuth.AccessToken,
			UserID:   user.ID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user repositories",
			})
			return
		}
	} else {
		allRepos, err = github.FetchOrgRepositories(github.FetchOrgRepositoriesInput{
			Organization: orgName,
			Token:        githubAuth.AccessToken,
			UserID:       user.ID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch organization repositories",
			})
			return
		}
	}

	// Convert to response format
	response := make([]RepositoryResponse, len(allRepos))
	for i, repo := range allRepos {
		response[i] = RepositoryResponse(repo)
	}

	c.JSON(http.StatusOK, response)
}
