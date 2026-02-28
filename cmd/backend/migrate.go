package main

import (
	"fmt"

	"github.com/hairizuanbinnoorazman/ui-automation/database"
	"github.com/spf13/cobra"
)

var (
	migrationsPath string
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

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

		// Run migrations
		if err := database.RunMigrations(sqlDB, migrationsPath); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		fmt.Println("Migrations applied successfully")
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the most recent migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

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

		// Rollback migration
		if err := database.RollbackMigration(sqlDB, migrationsPath); err != nil {
			return fmt.Errorf("failed to rollback migration: %w", err)
		}

		fmt.Println("Migration rolled back successfully")
		return nil
	},
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)

	migrateCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	migrateCmd.PersistentFlags().StringVarP(&migrationsPath, "path", "p", "database/migrations", "migrations directory path")

	rootCmd.AddCommand(migrateCmd)
}
