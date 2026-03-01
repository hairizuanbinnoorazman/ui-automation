package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg *viper.Viper

func initConfig() error {
	cfg = viper.New()
	cfg.SetConfigName(".ui-automation")
	cfg.SetConfigType("yaml")

	home, err := os.UserHomeDir()
	if err == nil {
		cfg.AddConfigPath(home)
	}

	cfg.SetDefault("url", "http://localhost:8080")
	cfg.SetDefault("token", "")

	cfg.SetEnvPrefix("UI_AUTOMATION")
	cfg.AutomaticEnv()

	// Read config file (ignore if not found)
	cfg.ReadInConfig()

	// CLI flags take highest priority
	if flagURL != "" {
		cfg.Set("url", flagURL)
	}
	if flagToken != "" {
		cfg.Set("token", flagToken)
	}

	return nil
}

func getConfigURL() string {
	return strings.TrimRight(cfg.GetString("url"), "/")
}

func getConfigToken() string {
	return cfg.GetString("token")
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigShowCmd())
	return cmd
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a config file template at ~/.ui-automation.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			configPath := filepath.Join(home, ".ui-automation.yaml")

			if _, err := os.Stat(configPath); err == nil {
				printMessage("Config file already exists at " + configPath)
				return nil
			}

			template := `# UI Automation CLI configuration
url: http://localhost:8080
token: ""
`
			if err := os.WriteFile(configPath, []byte(template), 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			printMessage("Config file created at " + configPath)
			return nil
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print the resolved configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := getConfigURL()
			token := getConfigToken()

			masked := "(not set)"
			if token != "" {
				if len(token) > 8 {
					masked = token[:4] + "..." + token[len(token)-4:]
				} else {
					masked = "****"
				}
			}

			printMessage(fmt.Sprintf("URL:   %s", url))
			printMessage(fmt.Sprintf("Token: %s", masked))

			if cfgFile := cfg.ConfigFileUsed(); cfgFile != "" {
				printMessage(fmt.Sprintf("Config file: %s", cfgFile))
			} else {
				printMessage("Config file: (none)")
			}

			return nil
		},
	}
}
