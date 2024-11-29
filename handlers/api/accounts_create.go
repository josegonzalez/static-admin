package api

import (
	"net/http"
	"static-admin/config"
	"static-admin/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateAccountRequest represents the JSON data for creating a new account
type CreateAccountRequest struct {
	Email    string `json:"email" binding:"required,email"`           // Email address
	Password string `json:"password" binding:"required,min=6,max=40"` // Password (6-40 characters)
}

// CreateAccountResponse represents the JSON response for a successful account creation
type CreateAccountResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

// NewCreateAccountHandler creates a new handler for the account creation endpoint
func NewCreateAccountHandler(config config.Config) (CreateAccountHandler, error) {
	return CreateAccountHandler{
		Database: config.Database,
	}, nil
}

// CreateAccountHandler handles the account creation request
type CreateAccountHandler struct {
	// Database is the database connection
	Database *gorm.DB
}

// GroupRegister registers the handler with the given router group
func (h CreateAccountHandler) GroupRegister(r *gin.RouterGroup) {
	r.PUT("/accounts", h.handler)
	r.OPTIONS("/accounts", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the PUT request for account creation
func (h CreateAccountHandler) handler(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Check if the email already exists
	var existingUser database.User
	if err := h.Database.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		// Email already exists
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already in use",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		// Handle unexpected database errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "An unexpected error occurred",
		})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password",
		})
		return
	}

	// Create the new user with hashed password
	newUser := database.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	if err := h.Database.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	// Return the created user without the password
	c.JSON(http.StatusCreated, CreateAccountResponse{
		ID:    newUser.ID,
		Email: newUser.Email,
	})
}
