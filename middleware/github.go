package middleware

import (
	"encoding/gob"

	"static-admin/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	zgithub "github.com/zalando/gin-oauth2/github"
	"golang.org/x/oauth2"
	oauth2gh "golang.org/x/oauth2/github"
)

var (
	GithubConfig *oauth2.Config
)

func Github(config config.Config) {
	credentials := zgithub.Credentials{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
	}
	GithubConfig = &oauth2.Config{
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
		RedirectURL:  config.GithubRedirectURL,
		Scopes:       config.GithubScopes,
		Endpoint:     oauth2gh.Endpoint,
	}
}

func GetLoginURL(c *gin.Context, state string) string {
	return GithubConfig.AuthCodeURL(state)
}

type GithubUser struct {
	AccessToken string `json:"access_token"`
	Expiry      string `json:"expiry"`
	Login       string `json:"login"`
	Name        string `json:"name"`
	URL         string `json:"url"`
}

func init() {
	gob.Register(GithubUser{})
}

func GithubAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Handle the exchange code to initiate a transport.
		session := sessions.Default(ctx)
		mysession := session.Get("ginoauthgh")
		if githubUser, ok := mysession.(GithubUser); ok {
			ctx.Set("githubUser", githubUser)
		}

		ctx.Next()
	}
}
