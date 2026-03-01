package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var (
	flagURL   string
	flagToken string
	flagJSON  bool
	flagDebug bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "uictl",
		Short: "CLI for UI Automation backend",
		Long:  "A command-line interface for managing projects, test procedures, test runs, and API tokens in the UI Automation system.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	}

	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "API server URL (env: UI_AUTOMATION_URL)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (env: UI_AUTOMATION_TOKEN)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug output")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("uictl %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newProjectsCmd())
	rootCmd.AddCommand(newProceduresCmd())
	rootCmd.AddCommand(newRunsCmd())
	rootCmd.AddCommand(newTokensCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
