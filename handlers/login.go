package handlers

import (
	"net/http"
	"time"

	"static-admin/config"
	"static-admin/database"
	"static-admin/session"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// LoginRequest represents the form data for logging in
type LoginRequest struct {
	// Email is the email address of the user
	Email string `form:"email" binding:"required,email"`

	// Password is the password of the user
	Password string `form:"password" binding:"required"`
}

// NewLoginHandler creates a new handler for the login page
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

// Register registers the handler with the given router
func (h LoginHandler) Register(r *gin.Engine) {
	r.GET("/login", h.handler)
	r.POST("/login", h.handler)
}

// handler handles the request for the page
func (h LoginHandler) handler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		context := session.PageContext(c, nil)
		c.HTML(http.StatusOK, "login.html", context)
		return
	}

	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		context := session.PageContext(c, err)
		c.HTML(http.StatusBadRequest, "login.html", context)
		return
	}

	var user database.User
	if err := h.Database.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"errors": []string{"Invalid credentials"}})
		return
	}

	if req.Password != user.Password { // In a real app, compare hashed passwords
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"errors": []string{"Invalid credentials"}})
		return
	}

	token, err := generateJWT(generateJWTInput{
		User:      user,
		JWTSecret: h.JWTSecret,
	})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"errors": []string{"Failed to generate token"}})
		return
	}

	c.SetCookie("jwt", token, 3600*72, "/", "", false, true) // Adjust to your requirements
	c.Redirect(http.StatusFound, "/dashboard")               // Redirect after successful login
}

// generateJWTInput contains the input parameters for generating a JWT token
type generateJWTInput struct {
	// User is the user for whom the token is generated
	User database.User

	// JWTSecret is the secret key for JWT
	JWTSecret []byte
}

// generateJWT generates a JWT token for the given user
func generateJWT(input generateJWTInput) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": input.User.ID,
		"name":    input.User.Name,
		"email":   input.User.Email,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // 3 days expiration
	})
	return token.SignedString(input.JWTSecret)
}
