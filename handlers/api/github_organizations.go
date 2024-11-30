package api

import (
	"context"
	"net/http"
	"sort"
	"static-admin/config"
	"static-admin/middleware"

	"static-admin/handlers/api/util"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// OrganizationResponse represents a GitHub organization in the JSON response
type OrganizationResponse struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// NewGitHubOrganizationsHandler creates a new handler for the GitHub organizations endpoint
func NewGitHubOrganizationsHandler(config config.Config) (GitHubOrganizationsHandler, error) {
	return GitHubOrganizationsHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// GitHubOrganizationsHandler handles the GitHub organizations request
type GitHubOrganizationsHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h GitHubOrganizationsHandler) GroupRegister(r *gin.RouterGroup) {
	r.GET("/github/organizations", h.handler)
	r.OPTIONS("/github/organizations", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the GET request for GitHub organizations
func (h GitHubOrganizationsHandler) handler(c *gin.Context) {
	githubAuth, exists := middleware.GetGitHubAuth(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "GitHub authentication required",
		})
		return
	}

	// Create GitHub client with user's access token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAuth.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// List organizations for authenticated user
	orgs, _, err := client.Organizations.List(ctx, "", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch organizations from GitHub",
		})
		return
	}

	// Convert to response format
	response := make([]OrganizationResponse, len(orgs))
	for i, org := range orgs {
		response[i] = OrganizationResponse{
			Login: util.StringFromPointer(org.Login),
			Name:  util.StringFromPointer(org.Name),
			URL:   util.StringFromPointer(org.URL),
		}
	}

	sort.Slice(response, func(i, j int) bool {
		return response[i].Login < response[j].Login
	})

	first := OrganizationResponse{
		Login: githubAuth.Login,
		Name:  githubAuth.Name,
		URL:   githubAuth.URL,
	}

	response = append([]OrganizationResponse{first}, response...)

	c.JSON(http.StatusOK, response)
}
