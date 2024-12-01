package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"static-admin/config"
	"static-admin/database"
	"static-admin/embedded_box/frontend"
	"static-admin/github"
	"static-admin/handlers"
	api_handlers "static-admin/handlers/api"
	auth_handlers "static-admin/handlers/auth"
	"static-admin/middleware"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
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

func GetManagerBoxEngine() *goview.ViewEngine {
	config := goview.DefaultConfig
	config.Root = frontend.Box.Root()
	config.Extension = ""
	config.Partials = frontend.Box.Partials()
	config.Delims = frontend.Box.Delims()
	config.Master = ""
	config.Funcs = template.FuncMap{}
	engine := goview.New(config)
	engine.SetFileHandler(frontend.Box.GoviewFileHandler())
	return engine
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

	r.SetHTMLTemplate(template.Must(template.ParseFS(staticFiles, "assets/*.html")))
	r.StaticFS("/static", http.FS(staticFiles))
	r.HTMLRender = ginview.Wrap(GetManagerBoxEngine())
	r.NoRoute(handlers.Index)

	apiUnauthenticated := r.Group("/api")
	apiUnauthenticated.Use(cors.New(cors.Config{
		// for testing purposes
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:" + strconv.Itoa(config.Port),
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	apiUnauthenticated.Use(middleware.User(middleware.UserMiddleware{
		Database:  db,
		JWTSecret: []byte(config.JWTSecret),
		IgnoredPaths: []string{
			"/auth/github/callback",
			"/api/login",
			"/api/register",
		},
	}))

	auth := r.Group("/")
	// auth.Use(middleware.Auth(db, []byte(config.JWTSecret)))

	registry := &Registry{Engine: r, AuthGroup: auth, ApiGroup: apiUnauthenticated}

	registry.AuthRegister(auth_handlers.NewGithubCallbackHandler(config))
	registry.ApiRegister(api_handlers.NewLoginHandler(config))
	registry.ApiRegister(api_handlers.NewCreateAccountHandler(config))
	registry.ApiRegister(api_handlers.NewGitHubAuthURLHandler(config))
	registry.ApiRegister(api_handlers.NewRevalidateHandler(config))
	registry.ApiRegister(api_handlers.NewGitHubRepositoriesHandler(config))
	registry.ApiRegister(api_handlers.NewGitHubOrganizationsHandler(config))
	registry.ApiRegister(api_handlers.NewSitesHandler(config))
	registry.ApiRegister(api_handlers.NewSiteCreateHandler(config))
	registry.ApiRegister(api_handlers.NewSiteDeleteHandler(config))
	registry.ApiRegister(api_handlers.NewPostsHandler(config))
	registry.ApiRegister(api_handlers.NewPostHandler(config))
	registry.ApiRegister(api_handlers.NewPostSaveHandler(config))
	registry.ApiRegister(api_handlers.NewTemplatesHandler(config))
	registry.ApiRegister(api_handlers.NewTemplateHandler(config))
	registry.ApiRegister(api_handlers.NewTemplateCreateHandler(config))
	registry.ApiRegister(api_handlers.NewTemplateUpdateHandler(config))
	registry.ApiRegister(api_handlers.NewTemplateDeleteHandler(config))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Port),
		Handler: r,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	log.Println("Server started on :" + strconv.Itoa(config.Port))

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
