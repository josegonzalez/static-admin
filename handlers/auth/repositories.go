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

// NewRepositoriesHandler creates a new handler for the repositories page
func NewRepositoriesHandler(config config.Config) (RepositoriesHandler, error) {
	return RepositoriesHandler{
		Database: config.Database,
	}, nil
}

// RepositoriesHandler handles the repositories request
type RepositoriesHandler struct {
	// Database is the database connection
	Database *gorm.DB
}

// AuthRegister registers the handler with the given router
func (h RepositoriesHandler) AuthRegister(auth *gin.RouterGroup) {
	auth.GET("/repositories", h.handler)
}

// handler handles the request for the page
func (h RepositoriesHandler) handler(c *gin.Context) {
	context := session.PageContext(c, nil)
	context["githubRedirectURL"] = middleware.GetLoginURL(c)

	organization := ""
	if organization = c.Query("organization"); organization == "" {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	context["organization"] = organization
	if user, ok := c.Get("githubUser"); ok {
		githubUser := user.(middleware.GithubUser)
		var err error
		var repositories []github.Repository

		if organization == githubUser.Login {
			repositories, err = github.FetchUserRepositories(githubUser.AccessToken)
		} else {
			repositories, err = github.FetchOrgRepositories(github.FetchOrgRepositoriesInput{
				Organization: organization,
				Token:        githubUser.AccessToken,
			})
		}

		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		context["repositories"] = repositories
	}

	c.HTML(http.StatusOK, "repositories.html", context)
}
