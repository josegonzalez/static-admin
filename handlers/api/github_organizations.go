package api

import (
	"context"
	"net/http"
	"sort"
	"static-admin/config"
	"static-admin/database"
	"strings"

	"static-admin/handlers/api/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
