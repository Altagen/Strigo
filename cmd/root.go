package cmd

import (
	"fmt"
	"os"
	"strigo/config"
	"strigo/logging"

	"github.com/spf13/cobra"
)

// Global config variable
var cfg *config.Config

// Global flags
var configFile string
var patternsFile string

// GetPatternsFilePath returns the patterns file path with priority resolution
// Priority: CLI flag > env var > config value
func GetPatternsFilePath() string {
	if patternsFile != "" {
		return patternsFile
	}
	if envPath := os.Getenv("STRIGO_PATTERNS_PATH"); envPath != "" {
		return envPath
	}
	return cfg.General.PatternsFile
}

// Root command
var rootCmd = &cobra.Command{
	Use:           "strigo",
	Short:         "Strigo - SDK & JDK Version Manager",
	Long:          `Strigo is a command-line tool that helps you manage multiple versions of different SDKs (like JDK) on your system.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration with optional config file override
		var err error
		cfg, err = config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Ensure required directories exist
		if err := config.EnsureDirectoriesExist(cfg); err != nil {
			return fmt.Errorf("error ensuring directories: %w", err)
		}

		// Initialize logger with JSON format if requested
		if err := logging.InitLogger(cfg.General.LogPath, cfg.General.LogLevel, jsonOutput || jsonLogs); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		return nil
	},
}

func init() {
	// Pre-log important startup messages before logger is initialized
	logging.PreLog("DEBUG", "Initializing Strigo...")

	// Add subcommands
	rootCmd.AddCommand(availableCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(listCmd)

	// Allow flags to be placed after arguments
	rootCmd.Flags().SetInterspersed(true)

	// Add flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (default: STRIGO_CONFIG_PATH or ./strigo.toml)")
	rootCmd.PersistentFlags().StringVarP(&patternsFile, "patterns", "p", "", "Path to patterns file (default: STRIGO_PATTERNS_PATH or config value)")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&jsonLogs, "json-logs", false, "Output logs in JSON format")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ExitWithError(err)
	}
}
