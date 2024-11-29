package handlers

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"
	"static-admin/session"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRequest represents the form data for registering a new user
type RegisterRequest struct {
	Name     string `form:"name" binding:"required"`           // Name of the user
	Email    string `form:"email" binding:"required,email"`    // Email address
	Password string `form:"password" binding:"required,min=6"` // Password (minimum 6 characters)
}

// NewRegisterHandler creates a new handler for the register page
func NewRegisterHandler(config config.Config) (RegisterHandler, error) {
	return RegisterHandler{
		Database: config.Database,
	}, nil
}

// RegisterHandler handles the register request
type RegisterHandler struct {
	// Database is the database connection
	Database *gorm.DB
}

// Register registers the handler with the given router
func (h RegisterHandler) Register(r *gin.Engine) {
	r.GET("/register", h.handler)
	r.POST("/register", h.handler)
}

// handler handles the request for the page
func (h RegisterHandler) handler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		context := session.PageContext(c, nil)
		c.HTML(http.StatusOK, "register.html", context)
		return
	}

	var req RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		context := session.PageContext(c, err)
		c.HTML(http.StatusBadRequest, "register.html", context)
		return
	}

	// Check if the email already exists
	var existingUser database.User
	if err := h.Database.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		// Email already exists
		c.HTML(http.StatusConflict, "register.html", gin.H{"errors": []string{"Email already in use"}})
		return
	} else if err != gorm.ErrRecordNotFound {
		// Handle unexpected database errors
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"errors": []string{"An unexpected error occurred"}})
		return
	}

	// Create the new user
	newUser := database.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password, // In a real application, hash the password
	}
	if err := h.Database.Create(&newUser).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"errors": []string{"Failed to create user"}})
		return
	}

	// Redirect after successful registration
	c.Redirect(http.StatusCreated, "/login")
}
