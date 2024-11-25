package auth

import (
	"net/http"
	"static-admin/config"
	"static-admin/github"
	"static-admin/middleware"
	"static-admin/session"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewDashboardHandler creates a new handler for the dashboard page
func NewDashboardHandler(config config.Config) (DashboardHandler, error) {
	return DashboardHandler{
		Database: config.Database,
	}, nil
}

// DashboardHandler handles the dashboard request
type DashboardHandler struct {
	// Database is the database connection
	Database *gorm.DB
}

// AuthRegister registers the handler with the given router
func (h DashboardHandler) AuthRegister(auth *gin.RouterGroup) {
	auth.GET("/dashboard", h.handler)
}

// handler handles the request for the page
func (h DashboardHandler) handler(c *gin.Context) {
	context := session.PageContext(c, nil)
	context["githubRedirectURL"] = middleware.GetLoginURL(c)

	if user, ok := c.Get("githubUser"); ok {
		githubUser := user.(middleware.GithubUser)
		organizations, err := github.FetchOrganizations(github.FetchOrganizationsInput{
			Username: githubUser.Login,
			Token:    githubUser.AccessToken,
		})
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		context["organizations"] = organizations
	}
	c.HTML(http.StatusOK, "dashboard.html", context)
}
