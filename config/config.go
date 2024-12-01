package config

import (
	"embed"
	"log"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type Config struct {
	// Database is the database connection
	Database *gorm.DB

	// GithubRedirectURL is the URL to redirect to after the GitHub login
	GithubRedirectURL string

	// GithubClientID is the client ID used to authenticate with the GitHub API
	GithubClientID string

	// GithubClientSecret is the client secret used to authenticate with the GitHub API
	GithubClientSecret string

	// GithubScopes is the list of scopes to request from the GitHub API
	GithubScopes []string

	// JWTSecret is the secret used to sign the JWT tokens
	JWTSecret string

	// Port is the port to run the server on
	Port int

	// StaticFiles is the embedded static files
	StaticFiles embed.FS
}

func NewConfig(database *gorm.DB, staticFiles embed.FS) Config {
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8080"
	}
	port, err := strconv.Atoi(defaultPort)
	if err != nil {
		log.Fatalf("Invalid PORT environment variable: %v", err)
	}

	config := Config{
		Database:           database,
		GithubRedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		GithubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GithubScopes:       []string{"repo", "read:user"},
		JWTSecret:          os.Getenv("JWT_SECRET"),
		StaticFiles:        staticFiles,
		Port:               port,
	}

	if config.GithubRedirectURL == "" {
		log.Fatal("GITHUB_REDIRECT_URL environment variable is required")
	}

	if config.GithubClientID == "" {
		log.Fatal("GITHUB_CLIENT_ID environment variable is required")
	}

	if config.GithubClientSecret == "" {
		log.Fatal("GITHUB_CLIENT_SECRET environment variable is required")
	}

	if config.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	if config.Port == 0 {
		config.Port = 8080
	}

	return config
}
