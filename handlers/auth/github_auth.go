package auth

import (
	"context"
	"fmt"
	"net/http"

	"static-admin/config"
	"static-admin/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

// NewGithubCallbackHandler creates a new handler for the github callback page
func NewGithubCallbackHandler(config config.Config) (GithubCallbackHandler, error) {
	return GithubCallbackHandler{}, nil
}

// GithubCallbackHandler handles the github callback request
type GithubCallbackHandler struct{}

// AuthRegister registers the handler with the given router
func (h GithubCallbackHandler) AuthRegister(auth *gin.RouterGroup) {
	auth.GET("/auth/github/callback", h.handler)
}

// handler handles the request for the page
func (h GithubCallbackHandler) handler(c *gin.Context) {
	session := sessions.Default(c)

	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: %s", retrievedState))
		return
	}

	stdctx := context.Background()
	tok, err := middleware.GithubConfig.Exchange(stdctx, c.Query("code"))
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to do exchange: %w", err))
		return
	}

	client := github.NewClient(middleware.GithubConfig.Client(stdctx, tok))
	user, _, err := client.Users.Get(stdctx, "")
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to get user: %w", err))
		return
	}

	// Protection: fields used in userinfo might be nil-pointers
	githubUser := middleware.GithubUser{
		Login:       stringFromPointer(user.Login),
		Name:        stringFromPointer(user.Name),
		URL:         stringFromPointer(user.URL),
		AccessToken: tok.AccessToken,
	}

	// save userinfo, which could be used in Handlers
	c.Set("githubUser", githubUser)

	// populate cookie
	session.Set("ginoauthgh", githubUser)
	if err := session.Save(); err != nil {
		glog.Errorf("Failed to save session: %v", err)
	}
	c.Redirect(http.StatusFound, "/dashboard")
}

func stringFromPointer(strPtr *string) (res string) {
	if strPtr == nil {
		res = ""
		return res
	}
	res = *strPtr
	return res
}
