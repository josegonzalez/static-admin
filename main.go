package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	pg "github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/josegonzalez/static-admin/markdown"
)

// CacheEntry represents a cached file or folder content
type CacheEntry struct {
	ID        int64     `pg:"id,pk"`
	Path      string    `pg:"path,unique"`
	Content   string    `pg:"content"`    // JSON string for content
	IsFile    bool      `pg:"is_file"`    // true if it's a file, false if it's a folder
	ExpiresAt time.Time `pg:"expires_at"` // Cache expiration timestamp
}

// GitHubProfile model
type GitHubProfile struct {
	ID           int64     `pg:"id,pk"`
	UserID       int64     `pg:"user_id,unique"`
	GitHubID     int64     `pg:"github_id,unique"`
	AccessToken  string    `pg:"access_token"`
	RefreshToken string    `pg:"refresh_token"`
	Username     string    `pg:"username"`
	ProfileURL   string    `pg:"profile_url"`
	AvatarURL    string    `pg:"avatar_url"`
	TokenExpiry  time.Time `pg:"token_expiry"`
	LinkedAt     time.Time `pg:"linked_at"`
}

// User model
type User struct {
	ID        int64     `pg:"id,pk"`
	Username  string    `pg:"username,unique"`
	Email     string    `pg:"email,unique"`
	Password  string    `pg:"password"`
	CreatedAt time.Time `pg:"created_at"`
}

// ClearExpiredCache removes expired cache entries from the database
func ClearExpiredCache(db *pg.DB) {
	_, err := db.Model((*CacheEntry)(nil)).
		Where("expires_at <= NOW()").
		Delete()
	if err != nil {
		log.Printf("Failed to clear expired cache: %v", err)
	} else {
		log.Println("Expired cache cleared successfully")
	}
}

//go:embed assets/*
var staticFiles embed.FS

// StartCacheCleaner starts a background process to clear expired cache periodically and stops on receiving a quit signal
func StartCacheCleaner(db *pg.DB, interval time.Duration, quit chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Running cache cleanup...")
				ClearExpiredCache(db)
			case <-quit:
				log.Println("Cache cleaner shutting down...")
				return
			}
		}
	}()
}

// Initialize the database connection
func initDB() *pg.DB {
	opts := &pg.Options{
		Addr:     "localhost:5432",
		User:     "postgres",
		Password: "password",
		Database: "user_auth",
	}
	db := pg.Connect(opts)

	// Create the users table if it doesn't exist
	err := db.Model((*User)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	})
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
	}

	// Create the github_profiles table if it doesn't exist
	err = db.Model((*GitHubProfile)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	})
	if err != nil {
		log.Fatalf("Error creating GitHub profiles table: %v", err)
	}

	return db
}

// Set a JWT token in a cookie
func setTokenCookie(c *gin.Context, token string) {
	c.SetCookie("jwt", token, 3600*24, "/", "localhost", false, true) // Secure should be true in production with HTTPS
}

// Get JWT token from the cookie
func getTokenFromCookie(c *gin.Context) (string, error) {
	return c.Cookie("jwt")
}

// AuthMiddleware checks if the user is authenticated and injects the user profile into the context
func AuthMiddleware(db *pg.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the JWT token from the cookie
		tokenString, err := getTokenFromCookie(c)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}

		// Parse the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			jwtSecret, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}

		// Extract user details from the token claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}

		// Fetch the user profile from the database
		user := &User{}
		err = db.Model(user).Where("username = ?", username).Select()
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}

		// Attach the user profile to the context
		c.Set("user", user)
		c.Next()
	}
}

func main() {
	// Initialize the database
	// db := initDB()
	// defer db.Close()

	// Initialize Gin router
	r := gin.Default()

	// Create a quit channel to signal the cache cleaner
	quit := make(chan struct{})

	// Start the cache cleaner with a 10-minute interval
	// StartCacheCleaner(db, 10*time.Minute, quit)

	// Serve static files
	r.StaticFS("/static", http.FS(staticFiles))

	// Editor page route
	r.GET("/editor", RenderEditorPage)

	// render the blocks json
	r.GET("/blocks", func(c *gin.Context) {
		source, err := staticFiles.ReadFile("assets/example.md")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read example.md: %s", err.Error())
			return
		}

		_, content, err := ExtractFrontMatter([]byte(source))
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to extract frontmatter: %s", err.Error())
			return
		}
		blocks, err := markdown.ParseMarkdownToBlocks(content)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to parse markdown: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, blocks)
	})

	// Start the Gin server in a separate Goroutine
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	log.Println("Server started on :8080")

	// Signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...", sig)

	// Signal the cache cleaner to stop
	close(quit)

	// Gracefully shut down the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
