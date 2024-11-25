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

// NewPostsHandler creates a new handler for the posts page
func NewPostsHandler(config config.Config) (PostsHandler, error) {
	return PostsHandler{
		Database: config.Database,
	}, nil
}

// PostsHandler handles the posts request
type PostsHandler struct {
	// Database is the database connection
	Database *gorm.DB
}

// AuthRegister registers the handler with the given router
func (h PostsHandler) AuthRegister(auth *gin.RouterGroup) {
	auth.GET("/posts", h.handler)
}

// handler handles the request for the page
func (h PostsHandler) handler(c *gin.Context) {
	organization := ""
	if organization = c.Query("organization"); organization == "" {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	repository := ""
	if repository = c.Query("repository"); repository == "" {
		c.Redirect(http.StatusFound, "/repositories?organization="+organization)
		return
	}

	branch := ""
	if branch = c.Query("branch"); branch == "" {
		c.Redirect(http.StatusFound, "/repositories?organization="+organization)
		return
	}

	context := session.PageContext(c, nil)
	context["githubRedirectURL"] = middleware.GetLoginURL(c)
	context["organization"] = organization
	context["repository"] = repository
	context["branch"] = branch

	if user, ok := c.Get("githubUser"); ok {
		githubUser := user.(middleware.GithubUser)
		files, err := github.FetchRepoFiles(github.FetchRepoFilesInput{
			Owner: organization,
			Repo:  repository,
			Path:  "_posts",
			Token: githubUser.AccessToken,
			Type:  "file",
		})
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		context["files"] = files
	}

	c.HTML(http.StatusOK, "posts.html", context)
}
