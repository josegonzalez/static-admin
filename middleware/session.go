package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var (
	store cookie.Store
)

func Session(secret string) gin.HandlerFunc {
	store = cookie.NewStore([]byte(secret))
	return sessions.Sessions("static-admin", store)
}
