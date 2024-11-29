package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"strings"

	"static-admin/github"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
	orgName := c.Param("org")
	if orgName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Organization name is required",
		})
		return
	}

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

	// Get GitHub auth data for user
	var githubAuth database.GitHubAuth
	if err := h.Database.Where("user_id = ?", uint(userID)).First(&githubAuth).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "GitHub authentication required",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch GitHub authentication",
		})
		return
	}

	var allRepos []github.Repository
	if orgName == githubAuth.Login {
		allRepos, err = github.FetchUserRepositories(github.FetchUserRepositoriesInput{
			Username: githubAuth.Login,
			Token:    githubAuth.AccessToken,
			UserID:   uint(userID),
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
			UserID:       uint(userID),
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
