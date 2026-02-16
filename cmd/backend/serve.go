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
	"github.com/hairizuan-noorazman/ui-automation/session"
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

	// Initialize stores
	userStore := user.NewMySQLStore(db, log)

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
