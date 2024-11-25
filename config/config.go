package config

import (
	"embed"
	"log"
	"os"

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

	// GithubToken is the token used to authenticate with the GitHub API
	GithubToken string

	// JWTSecret is the secret used to sign the JWT tokens
	JWTSecret string

	// SessionSecret is the secret used to encrypt the session
	SessionSecret string

	// StaticFiles is the embedded static files
	StaticFiles embed.FS
}

func NewConfig(database *gorm.DB, staticFiles embed.FS) Config {
	config := Config{
		Database:           database,
		GithubToken:        os.Getenv("GITHUB_TOKEN"),
		GithubRedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		GithubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GithubScopes:       []string{"repo", "read:user"},
		JWTSecret:          os.Getenv("JWT_SECRET"),
		SessionSecret:      os.Getenv("SESSION_SECRET"),
		StaticFiles:        staticFiles,
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

	if config.GithubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	if config.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	if config.SessionSecret == "" {
		log.Fatal("SESSION_SECRET environment variable is required")
	}

	return config
}
