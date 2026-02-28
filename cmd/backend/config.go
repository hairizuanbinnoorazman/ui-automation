package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// AgentConfig holds agent pipeline configuration.
type AgentConfig struct {
	MaxIterations       int
	TimeLimit           time.Duration
	BedrockRegion       string
	BedrockModel        string
	BedrockAccessKey    string
	BedrockSecretKey    string
	PlaywrightMCPURL    string
	AgentScriptPath     string
	MaxConcurrentWorkers int
}

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Storage  StorageConfig
	Log      LogConfig
	Agent    AgentConfig
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

// StorageConfig holds blob storage configuration.
type StorageConfig struct {
	Type            string        // "local" or "s3"
	BaseDir         string        // For local: "./uploads"
	S3Bucket        string        // For S3: bucket name
	S3Region        string        // For S3: AWS region
	S3PresignExpiry time.Duration // Presigned URL expiration
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
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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

	v.SetDefault("storage.type", "local")
	v.SetDefault("storage.base_dir", "./uploads")
	v.SetDefault("storage.s3_bucket", "")
	v.SetDefault("storage.s3_region", "us-east-1")
	v.SetDefault("storage.s3_presign_expiry", "15m")

	v.SetDefault("log.level", "info")

	v.SetDefault("agent.max_iterations", 50)
	v.SetDefault("agent.time_limit", "10m")
	v.SetDefault("agent.bedrock_region", "us-east-1")
	v.SetDefault("agent.bedrock_model", "anthropic.claude-sonnet-4-6")
	v.SetDefault("agent.bedrock_access_key", "")
	v.SetDefault("agent.bedrock_secret_key", "")
	v.SetDefault("agent.playwright_mcp_url", "http://localhost:3000")
	v.SetDefault("agent.script_path", "/app/agent/agent_runner.py")
	v.SetDefault("agent.max_concurrent_workers", 1)

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

	config.Storage.Type = v.GetString("storage.type")
	config.Storage.BaseDir = v.GetString("storage.base_dir")
	config.Storage.S3Bucket = v.GetString("storage.s3_bucket")
	config.Storage.S3Region = v.GetString("storage.s3_region")
	config.Storage.S3PresignExpiry = v.GetDuration("storage.s3_presign_expiry")

	config.Log.Level = v.GetString("log.level")

	config.Agent.MaxIterations = v.GetInt("agent.max_iterations")
	config.Agent.TimeLimit = v.GetDuration("agent.time_limit")
	config.Agent.BedrockRegion = v.GetString("agent.bedrock_region")
	config.Agent.BedrockModel = v.GetString("agent.bedrock_model")
	config.Agent.BedrockAccessKey = v.GetString("agent.bedrock_access_key")
	config.Agent.BedrockSecretKey = v.GetString("agent.bedrock_secret_key")
	config.Agent.PlaywrightMCPURL = v.GetString("agent.playwright_mcp_url")
	config.Agent.AgentScriptPath = v.GetString("agent.script_path")
	config.Agent.MaxConcurrentWorkers = v.GetInt("agent.max_concurrent_workers")

	return &config, nil
}
