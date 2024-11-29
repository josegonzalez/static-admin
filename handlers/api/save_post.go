package api

import (
	"net/http"
	"static-admin/blocks"
	"static-admin/config"
	"static-admin/database"
	"static-admin/markdown"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// SavePostRequest represents the JSON request for saving a post's content
type SavePostRequest struct {
	ID          string                      `json:"id"`
	Path        string                      `json:"path"`
	Frontmatter []markdown.FrontmatterField `json:"frontmatter"`
	Blocks      []blocks.Block              `json:"blocks"`
}

// SavePostResponse represents the JSON response for saving a post's content
type SavePostResponse struct {
	Message  string          `json:"message"`
	Content  SavePostRequest `json:"content"`
	Path     string          `json:"path"`
	Markdown string          `json:"markdown"`
}

// NewSavePostHandler creates a new handler for saving post content
func NewSavePostHandler(config config.Config) (SavePostHandler, error) {
	return SavePostHandler{
		Database:  config.Database,
		JWTSecret: []byte(config.JWTSecret),
	}, nil
}

// SavePostHandler handles the save post request
type SavePostHandler struct {
	Database  *gorm.DB
	JWTSecret []byte
}

// GroupRegister registers the handler with the given router group
func (h SavePostHandler) GroupRegister(r *gin.RouterGroup) {
	r.POST("/sites/:siteId/posts/:postId", h.handler)
}

// handler handles the POST request for saving post content
func (h SavePostHandler) handler(c *gin.Context) {
	// Extract bearer token
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing bearer token",
		})
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	// Extract user ID from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process token",
		})
		return
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID in token",
		})
		return
	}

	// Get site ID and post ID from URL
	siteID := c.Param("siteId")
	postID := c.Param("postId")
	if siteID == "" || postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Site ID and Post ID are required",
		})
		return
	}

	// Verify site ownership
	var site database.Site
	if err := h.Database.Where("id = ? AND user_id = ?", siteID, uint(userID)).First(&site).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Site not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch site details",
		})
		return
	}

	// Parse request body
	var req SavePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// decode the id into a path
	path, err := fromBase62(req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to decode post ID",
		})
		return
	}

	// Generate frontmatter YAML
	frontmatterYaml, err := markdown.FrontmatterFieldToYaml(req.Frontmatter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate frontmatter",
		})
		return
	}

	// Generate markdown from blocks
	contentMarkdown, err := blocks.ParseBlocksToMarkdown(req.Blocks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate markdown",
		})
		return
	}

	// Combine frontmatter and content
	fullMarkdown := frontmatterYaml + contentMarkdown

	c.JSON(http.StatusOK, SavePostResponse{
		Message:  "Received post content",
		Content:  req,
		Path:     path,
		Markdown: fullMarkdown,
	})
}
