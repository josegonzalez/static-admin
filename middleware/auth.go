package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"static-admin/database"
	"static-admin/session"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// AuthMiddleware ensures the user is authenticated and not deleted
func Auth(db *gorm.DB, jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the cookie
		tokenString, err := c.Cookie("jwt")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				_ = session.AddError(c, "User is not authenticated and cannot access the page")
			} else {
				_ = session.AddError(c, "Failed to get token from cookie")
			}
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Parse and validate the JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.SetCookie("jwt", "", -1, "/", "", false, true)
			_ = session.AddError(c, "Invalid or expired token: "+err.Error())
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		if !token.Valid {
			c.SetCookie("jwt", "", -1, "/", "", false, true)
			_ = session.AddError(c, "Invalid or expired token")
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Extract claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.SetCookie("jwt", "", -1, "/", "", false, true)
			_ = session.AddError(c, "Invalid token claims")
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		userID := claims["user_id"]
		if userID == nil {
			c.SetCookie("jwt", "", -1, "/", "", false, true)
			_ = session.AddError(c, "Invalid token payload")
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Check if the user exists and is not deleted
		var user database.User
		if err := db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
			c.SetCookie("jwt", "", -1, "/", "", false, true)
			if err == gorm.ErrRecordNotFound {
				_ = session.AddError(c, "User not found or deleted")
			} else {
				_ = session.AddError(c, "Failed to verify user")
			}
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Add user information to the context
		c.Set("user", user)
		c.Next()
	}
}
