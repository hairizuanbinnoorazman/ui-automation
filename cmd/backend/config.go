package main

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Log      LogConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
}

// SessionConfig holds session management configuration.
type SessionConfig struct {
	CookieName   string
	CookieSecret string
	Duration     time.Duration
	Secure       bool
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string
}

// LoadConfig loads configuration from file and environment variables.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Enable environment variable overrides
	v.AutomaticEnv()

	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "password")
	v.SetDefault("database.database", "ui_automation")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)

	v.SetDefault("session.cookie_name", "session_id")
	v.SetDefault("session.cookie_secret", "change-this-secret-in-production-min-32-chars")
	v.SetDefault("session.duration", "24h")
	v.SetDefault("session.secure", false)

	v.SetDefault("log.level", "info")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; using defaults
	}

	// Parse configuration
	var config Config

	config.Server.Host = v.GetString("server.host")
	config.Server.Port = v.GetInt("server.port")
	config.Server.ReadTimeout = v.GetDuration("server.read_timeout")
	config.Server.WriteTimeout = v.GetDuration("server.write_timeout")

	config.Database.Host = v.GetString("database.host")
	config.Database.Port = v.GetInt("database.port")
	config.Database.User = v.GetString("database.user")
	config.Database.Password = v.GetString("database.password")
	config.Database.Database = v.GetString("database.database")
	config.Database.MaxOpenConns = v.GetInt("database.max_open_conns")
	config.Database.MaxIdleConns = v.GetInt("database.max_idle_conns")

	config.Session.CookieName = v.GetString("session.cookie_name")
	config.Session.CookieSecret = v.GetString("session.cookie_secret")
	config.Session.Duration = v.GetDuration("session.duration")
	config.Session.Secure = v.GetBool("session.secure")

	config.Log.Level = v.GetString("log.level")

	return &config, nil
}
