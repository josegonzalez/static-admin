package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"static-admin/config"
	"static-admin/database"
	"static-admin/github"
	"static-admin/handlers"
	api_handlers "static-admin/handlers/api"
	auth_handlers "static-admin/handlers/auth"
	"static-admin/markdown"
	"static-admin/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed assets/*
var staticFiles embed.FS

type Handler interface {
	Register(r *gin.Engine)
}

type AuthHandler interface {
	GroupRegister(r *gin.RouterGroup)
}

type Registry struct {
	Engine    *gin.Engine
	AuthGroup *gin.RouterGroup
	ApiGroup  *gin.RouterGroup
}

func (r *Registry) Register(handler Handler, err error) {
	if err != nil {
		log.Fatalf("Failed to create editor handler: %v", err)
	}

	handler.Register(r.Engine)
}

func (r *Registry) ApiRegister(handler AuthHandler, err error) {
	if err != nil {
		log.Fatalf("Failed to create auth handler: %v", err)
	}

	handler.GroupRegister(r.ApiGroup)
}

func (r *Registry) AuthRegister(handler AuthHandler, err error) {
	if err != nil {
		log.Fatalf("Failed to create auth handler: %v", err)
	}

	handler.GroupRegister(r.AuthGroup)
}

func main() {
	// Initialize the database
	db, err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	defer func() {
		dbInstance, _ := db.DB()
		_ = dbInstance.Close()
	}()

	// Create a quit channel to signal the cache cleaner
	quit := make(chan struct{})
	github.StartCacheCleaner(quit)

	config := config.NewConfig(db, staticFiles)
	middleware.Github(config)

	r := gin.Default()
	r.Use(middleware.Session(config.SessionSecret))
	r.Use(middleware.GithubAuth())

	r.SetHTMLTemplate(template.Must(template.ParseFS(staticFiles, "assets/*.html")))
	r.StaticFS("/static", http.FS(staticFiles))

	apiUnauthenticated := r.Group("/api")
	apiUnauthenticated.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
		}, // Frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	auth := r.Group("/")
	// auth.Use(middleware.Auth(db, []byte(config.JWTSecret)))

	registry := &Registry{Engine: r, AuthGroup: auth, ApiGroup: apiUnauthenticated}

	registry.Register(handlers.NewLoginHandler(config))
	registry.Register(handlers.NewRegisterHandler(config))

	registry.AuthRegister(auth_handlers.NewGithubCallbackHandler(config))
	registry.ApiRegister(api_handlers.NewLoginHandler(config))
	registry.ApiRegister(api_handlers.NewCreateAccountHandler(config))
	registry.ApiRegister(api_handlers.NewGitHubAuthURLHandler(config))
	registry.ApiRegister(api_handlers.NewRevalidateHandler(config))
	registry.ApiRegister(api_handlers.NewGitHubOrganizationsHandler(config))
	registry.ApiRegister(api_handlers.NewSitesHandler(config))
	registry.ApiRegister(api_handlers.NewDeleteSiteHandler(config))
	registry.ApiRegister(api_handlers.NewCreateSiteHandler(config))
	registry.ApiRegister(api_handlers.NewGitHubRepositoriesHandler(config))

	// registry.AuthRegister(auth_handlers.NewEditorHandler(config))
	// registry.AuthRegister(auth_handlers.NewDashboardHandler(config))
	// registry.AuthRegister(auth_handlers.NewConfigurationHandler(config))
	// registry.AuthRegister(auth_handlers.NewPostsHandler(config))

	// render the blocks json
	r.GET("/blocks", func(c *gin.Context) {
		source, err := staticFiles.ReadFile("assets/example.md")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read example.md: %s", err.Error())
			return
		}

		_, content, err := markdown.ExtractFrontMatter([]byte(source))
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
