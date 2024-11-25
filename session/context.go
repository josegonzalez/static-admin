package session

import "github.com/gin-gonic/gin"

func PageContext(c *gin.Context, err error) map[string]any {
	githubUser, ok := c.Get("githubUser")
	if !ok {
		githubUser = nil
	}

	return map[string]any{
		"errors":     GetErrors(c, err),
		"githubUser": githubUser,
	}
}
