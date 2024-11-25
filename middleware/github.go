package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"

	"static-admin/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
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

func GetLoginURL(c *gin.Context) string {
	state := randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	_ = session.Save()
	return GithubConfig.AuthCodeURL(state)
}

func randToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		glog.Fatalf("[Gin-OAuth] Failed to read rand: %v\n", err)
	}
	return base64.StdEncoding.EncodeToString(b)
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
