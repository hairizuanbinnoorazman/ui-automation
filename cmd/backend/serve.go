package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hairizuan-noorazman/ui-automation/cmd/backend/handlers"
	"github.com/hairizuan-noorazman/ui-automation/database"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/project"
	"github.com/hairizuan-noorazman/ui-automation/session"
	"github.com/hairizuan-noorazman/ui-automation/storage"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
	"github.com/hairizuan-noorazman/ui-automation/testrun"
	"github.com/hairizuan-noorazman/ui-automation/user"
	"github.com/spf13/cobra"
)

var configFile string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE:  runServer,
}

func init() {
	serveCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file path")
	rootCmd.AddCommand(serveCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log := logger.NewLogrusLogger(cfg.Log.Level)
	log.Info(ctx, "starting server", map[string]interface{}{
		"version": Version,
		"commit":  Commit,
		"date":    BuildDate,
	})

	// Connect to database
	dbCfg := database.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Database:     cfg.Database.Database,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	defer sqlDB.Close()

	log.Info(ctx, "database connected", map[string]interface{}{
		"host":     cfg.Database.Host,
		"port":     cfg.Database.Port,
		"database": cfg.Database.Database,
	})

	// Initialize storage
	storageConfig := map[string]interface{}{
		"base_dir":       cfg.Storage.BaseDir,
		"bucket":         cfg.Storage.S3Bucket,
		"region":         cfg.Storage.S3Region,
		"presign_expiry": cfg.Storage.S3PresignExpiry,
	}

	blobStorage, err := storage.NewBlobStorage(cfg.Storage.Type, storageConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Log storage initialization
	logFields := map[string]interface{}{"type": cfg.Storage.Type}
	if cfg.Storage.Type == "local" {
		logFields["base_dir"] = cfg.Storage.BaseDir
	} else if cfg.Storage.Type == "s3" {
		logFields["bucket"] = cfg.Storage.S3Bucket
		logFields["region"] = cfg.Storage.S3Region
	}
	log.Info(ctx, "storage initialized", logFields)

	// Initialize stores
	userStore := user.NewMySQLStore(db, log)
	projectStore := project.NewMySQLStore(db, log)
	testProcedureStore := testprocedure.NewMySQLStore(db, log)
	testRunStore := testrun.NewMySQLStore(db, log)
	assetStore := testrun.NewMySQLAssetStore(db, log)

	// Initialize session manager
	sessionManager := session.NewManager(cfg.Session.Duration, log)
	sessionManager.StartCleanup(5 * time.Minute)
	defer sessionManager.StopCleanup()

	log.Info(ctx, "session manager initialized", map[string]interface{}{
		"duration": cfg.Session.Duration.String(),
	})

	// Setup router
	router := mux.NewRouter()

	// Health check endpoint (public)
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")

	// Auth handlers (public)
	authHandler := handlers.NewAuthHandler(
		userStore,
		sessionManager,
		cfg.Session.CookieSecret,
		cfg.Session.CookieName,
		cfg.Session.Secure,
		log,
	)

	router.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/v1/auth/logout", authHandler.Logout).Methods("POST")

	// Protected user routes
	userHandler := handlers.NewUserHandler(userStore, log)
	authMiddleware := handlers.NewAuthMiddleware(sessionManager, cfg.Session.CookieName, log)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(authMiddleware.Handler)

	apiRouter.HandleFunc("/users", userHandler.List).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", userHandler.GetByID).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", userHandler.Update).Methods("PUT")
	apiRouter.HandleFunc("/users/{id}", userHandler.Delete).Methods("DELETE")

	// Project routes (protected)
	projectHandler := handlers.NewProjectHandler(projectStore, log)
	projectAuth := handlers.NewProjectAuthorizationMiddleware(projectStore, log)

	apiRouter.HandleFunc("/projects", projectHandler.List).Methods("GET")
	apiRouter.HandleFunc("/projects", projectHandler.Create).Methods("POST")

	// Project-specific routes with authorization
	projectRouter := apiRouter.PathPrefix("/projects/{id}").Subrouter()
	projectRouter.Use(projectAuth.Handler)
	projectRouter.HandleFunc("", projectHandler.GetByID).Methods("GET")
	projectRouter.HandleFunc("", projectHandler.Update).Methods("PUT")
	projectRouter.HandleFunc("", projectHandler.Delete).Methods("DELETE")

	// Test Procedure routes (protected by project authorization)
	testProcedureHandler := handlers.NewTestProcedureHandler(testProcedureStore, log)

	// List and create procedures for a project
	apiRouter.HandleFunc("/projects/{project_id}/procedures", testProcedureHandler.List).Methods("GET")
	apiRouter.HandleFunc("/projects/{project_id}/procedures", testProcedureHandler.Create).Methods("POST")

	// Individual procedure operations
	apiRouter.HandleFunc("/projects/{project_id}/procedures/{id}", testProcedureHandler.GetByID).Methods("GET")
	apiRouter.HandleFunc("/projects/{project_id}/procedures/{id}", testProcedureHandler.Update).Methods("PUT")
	apiRouter.HandleFunc("/projects/{project_id}/procedures/{id}", testProcedureHandler.Delete).Methods("DELETE")

	// Versioning operations
	apiRouter.HandleFunc("/projects/{project_id}/procedures/{id}/versions", testProcedureHandler.CreateVersion).Methods("POST")
	apiRouter.HandleFunc("/projects/{project_id}/procedures/{id}/versions", testProcedureHandler.GetVersionHistory).Methods("GET")

	// Test Run routes (protected)
	testRunHandler := handlers.NewTestRunHandler(testRunStore, assetStore, blobStorage, log)

	// List and create runs for a procedure
	apiRouter.HandleFunc("/procedures/{procedure_id}/runs", testRunHandler.List).Methods("GET")
	apiRouter.HandleFunc("/procedures/{procedure_id}/runs", testRunHandler.Create).Methods("POST")

	// Individual run operations
	apiRouter.HandleFunc("/runs/{run_id}", testRunHandler.GetByID).Methods("GET")
	apiRouter.HandleFunc("/runs/{run_id}", testRunHandler.Update).Methods("PUT")
	apiRouter.HandleFunc("/runs/{run_id}/start", testRunHandler.Start).Methods("POST")
	apiRouter.HandleFunc("/runs/{run_id}/complete", testRunHandler.Complete).Methods("POST")

	// Asset operations
	apiRouter.HandleFunc("/runs/{run_id}/assets", testRunHandler.UploadAsset).Methods("POST")
	apiRouter.HandleFunc("/runs/{run_id}/assets", testRunHandler.ListAssets).Methods("GET")
	apiRouter.HandleFunc("/runs/{run_id}/assets/{asset_id}", testRunHandler.DownloadAsset).Methods("GET")
	apiRouter.HandleFunc("/runs/{run_id}/assets/{asset_id}", testRunHandler.DeleteAsset).Methods("DELETE")

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info(ctx, "server listening", map[string]interface{}{
			"address": addr,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(ctx, "server error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(ctx, "shutting down server", nil)

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Info(ctx, "server stopped", nil)
	return nil
}
