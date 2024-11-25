package auth

import (
	"net/http"
	"static-admin/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewConfigurationHandler creates a new handler for the configuration page
func NewConfigurationHandler(config config.Config) (ConfigurationHandler, error) {
	return ConfigurationHandler{
		Database: config.Database,
	}, nil
}

// ConfigurationHandler handles the configuration request
type ConfigurationHandler struct {
	// Database is the database connection
	Database *gorm.DB
}

// AuthRegister registers the handler with the given router
func (h ConfigurationHandler) AuthRegister(auth *gin.RouterGroup) {
	auth.GET("/configuration", h.handler)
}

// handler handles the request for the page
func (h ConfigurationHandler) handler(c *gin.Context) {
	c.HTML(http.StatusOK, "configuration.html", nil)
}
