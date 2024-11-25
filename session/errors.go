package session

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AddError adds an error message to the session
func AddError(c *gin.Context, message string) error {
	session := sessions.Default(c)
	errors := session.Get("errors")

	var errorList []string
	if errors != nil {
		errorList = errors.([]string)
	}
	errorList = append(errorList, message)

	session.Set("errors", errorList)
	return session.Save()
}

// HasErrors checks if there are any error messages in the session
func HasErrors(c *gin.Context) bool {
	session := sessions.Default(c)
	errors := session.Get("errors")

	var errorList []string
	if errors != nil {
		errorList = errors.([]string)
	}

	return len(errorList) > 0
}

// GetErrors retrieves and clears all error messages from the session
func GetErrors(c *gin.Context, err error) []string {
	session := sessions.Default(c)
	errors := session.Get("errors")

	errorList := []string{}
	if errors != nil {
		errorList = errors.([]string)
	}

	if err != nil {
		errorList = append(errorList, err.Error())
	}

	session.Delete("errors")
	err = session.Save()
	if err != nil {
		errorList = append(errorList, err.Error())
	}

	return errorList
}
