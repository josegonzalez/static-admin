package api

import (
	"net/http"
	"time"

	"static-admin/config"
	"static-admin/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginRequest represents the JSON data for logging in
type LoginRequest struct {
	// Email is the email address of the user
	Email string `json:"email" binding:"required,email"`

	// Password is the password of the user
	Password string `json:"password" binding:"required,max=40"`
}

// LoginResponse represents the JSON response for a successful login
type LoginResponse struct {
	Token string `json:"token"`
}

// NewLoginHandler creates a new handler for the API login endpoint
func NewLoginHandler(config config.Config) (LoginHandler, error) {
	return LoginHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// LoginHandler handles the login request
type LoginHandler struct {
	// Database is the database connection
	Database *gorm.DB

	// JWTSecret is the secret key for JWT
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h LoginHandler) GroupRegister(r *gin.RouterGroup) {
	r.POST("/login", h.handler)
	r.OPTIONS("/login", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}

// handler handles the POST request for login
func (h LoginHandler) handler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	var user database.User
	if err := h.Database.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	token, err := generateJWT(generateJWTInput{
		User:      user,
		JWTSecret: h.JWTSecret,
		Database:  h.Database,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
	})
}

// generateJWTInput contains the input parameters for generating a JWT token
type generateJWTInput struct {
	// User is the user for whom the token is generated
	User database.User

	// JWTSecret is the secret key for JWT
	JWTSecret []byte

	// Database is the database connection
	Database *gorm.DB
}

// generateJWT generates a JWT token for the given user
func generateJWT(input generateJWTInput) (string, error) {
	// Check if user has GitHub authentication
	var githubAuth database.GitHubAuth
	hasGitHubAuth := input.Database.Where("user_id = ?", input.User.ID).First(&githubAuth).Error == nil

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     input.User.ID,
		"name":        input.User.Name,
		"email":       input.User.Email,
		"github_auth": hasGitHubAuth,
		"exp":         time.Now().Add(time.Hour * 72).Unix(), // 3 days expiration
	})
	return token.SignedString(input.JWTSecret)
}
