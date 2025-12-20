package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/logging"
	"strings"

	"github.com/pelletier/go-toml"
)

// GeneralConfig holds general configuration parameters
type GeneralConfig struct {
	LogLevel          string `toml:"log_level"`
	SDKInstallDir     string `toml:"sdk_install_dir"`
	CacheDir          string `toml:"cache_dir"`
	LogPath           string `toml:"log_path"`
	KeepCache         bool   `toml:"keep_cache"`
	JDKSecurityPath   string `toml:"jdk_security_path"`
	SystemCacertsPath string `toml:"system_cacerts_path"`
	ShellConfigPath   string `toml:"shell_config_path"`
	PatternsFile      string `toml:"patterns_file"` // Path to strigopatterns.toml (default: strigopatterns.toml)
}

// SDKType represents a referenced SDK type configuration
type SDKType struct {
	Type       string `toml:"type"`
	InstallDir string `toml:"install_dir"`
}

// Registry represents a remote registry configuration
type Registry struct {
	Type     string `toml:"type"`
	APIURL   string `toml:"api_url"`
	Username string `toml:"username,omitempty"` // Optional: for authenticated registries
	Password string `toml:"password,omitempty"` // Optional: for authenticated registries
}

// SDKRepository represents a referenced SDK configuration
type SDKRepository struct {
	Type       string `toml:"type"`
	Registry   string `toml:"registry"`
	Repository string `toml:"repository"`
	Path       string `toml:"path"`
}

// Config represents the main configuration structure
type Config struct {
	General         GeneralConfig            `toml:"general"`
	Registries      map[string]Registry      `toml:"registries"`
	SDKTypes        map[string]SDKType       `toml:"sdk_types"`
	SDKRepositories map[string]SDKRepository `toml:"sdk_repositories"`
}

// ExpandTilde expands ~ to the user's home directory
func ExpandTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

// LoadConfig loads and parses the configuration file
// Priority: cliPath > STRIGO_CONFIG_PATH env var > ./strigo.toml
func LoadConfig(cliPath string) (*Config, error) {
	// Detect configuration file with priority
	var configPath string
	if cliPath != "" {
		// 1. CLI flag has highest priority
		configPath = cliPath
	} else if envPath := os.Getenv("STRIGO_CONFIG_PATH"); envPath != "" {
		// 2. Environment variable
		configPath = envPath
	} else {
		// 3. Default fallback
		configPath = "strigo.toml"
	}

	// Prelog for capture before InitLogger
	logging.PreLog("DEBUG", "📂 Loading configuration from: %s", configPath)

	// Read configuration file
	file, err := os.ReadFile(configPath)
	if err != nil {
		logging.PreLog("ERROR", "❌ Failed to read config file: %v", err)
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Debug: Display raw file content
	logging.PreLog("DEBUG", "📜 Raw file content:\n%s", string(file))

	// Unmarshal TOML file
	var cfg Config
	err = toml.Unmarshal(file, &cfg)
	if err != nil {
		// Always print to stderr for parsing errors (before log level is set)
		fmt.Fprintf(os.Stderr, "❌ Failed to parse config file '%s': %v\n", configPath, err)

		// Check for common issues
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "defined twice") || strings.Contains(errMsg, "already defined") {
			fmt.Fprintf(os.Stderr, "\n💡 Hint: You have duplicate keys in your TOML file. Each repository/registry name must be unique.\n")
		}

		logging.PreLog("ERROR", "❌ Failed to parse config file: %v", err)
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Debug: Display decoded structure
	logging.PreLog("DEBUG", "🔍 Decoded Config: %+v", cfg)

	// Check required fields (LogPath can be empty)
	if cfg.General.SDKInstallDir == "" || cfg.General.CacheDir == "" {
		logging.PreLog("ERROR", "❌ Configuration values are empty! Check your `strigo.toml`.")
		logging.PreLog("DEBUG", "🔍 Debug: SDKInstallDir=%q | CacheDir=%q", cfg.General.SDKInstallDir, cfg.General.CacheDir)
		return nil, fmt.Errorf("one or more required configuration paths are empty")
	}

	// Apply temporary log level to filter PreLog()
	logging.SetPreLogLevel(cfg.General.LogLevel)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logging.PreLog("ERROR", "❌ Configuration validation failed: %v", err)
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logging.PreLog("DEBUG", "✅ Configuration successfully loaded and validated.")
	return &cfg, nil
}

// EnsureDirectoriesExist checks and creates required directories
func EnsureDirectoriesExist(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil, cannot ensure directories")
	}

	// Debug: Display directory paths
	logging.LogDebug("🔍 Checking directory paths: LogPath=%s, SDKInstallDir=%s, CacheDir=%s",
		cfg.General.LogPath, cfg.General.SDKInstallDir, cfg.General.CacheDir)

	// Check that required paths are not empty
	if cfg.General.SDKInstallDir == "" || cfg.General.CacheDir == "" {
		return fmt.Errorf("one or more required directory paths are empty in configuration")
	}

	// List of mandatory directories to create
	mandatoryPaths := []string{cfg.General.SDKInstallDir, cfg.General.CacheDir}

	// LogPath is optional, add it only if defined
	if cfg.General.LogPath != "" {
		mandatoryPaths = append(mandatoryPaths, cfg.General.LogPath)
	} else {
		logging.LogDebug("LogPath is empty. Logs will be written only to stdout.")
	}

	// Create directories if needed
	for _, path := range mandatoryPaths {
		logging.LogDebug("📂 Ensuring directory exists: %s", path)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			logging.LogError("❌ Failed to create directory %s: %v", path, err)
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	logging.LogDebug("✅ Directories verified.")
	return nil
}

// Validate checks the configuration validity
func (c *Config) Validate() error {
	// Expand tilde in shell_config_path if set
	if c.General.ShellConfigPath != "" {
		expandedPath, err := ExpandTilde(c.General.ShellConfigPath)
		if err != nil {
			return fmt.Errorf("failed to expand shell_config_path: %w", err)
		}

		// Check if the file exists
		if _, err := os.Stat(expandedPath); err != nil {
			return fmt.Errorf("shell configuration file not found: %s", expandedPath)
		}

		c.General.ShellConfigPath = expandedPath
	}

	// Check that required paths for JDKs are defined
	if c.General.JDKSecurityPath == "" {
		return fmt.Errorf("jdk_security_path must be set")
	}
	if c.General.SystemCacertsPath == "" {
		return fmt.Errorf("system_cacerts_path must be set")
	}

	// Check if system_cacerts_path exists
	if _, err := os.Stat(c.General.SystemCacertsPath); os.IsNotExist(err) {
		return fmt.Errorf("system_cacerts_path does not exist: %s", c.General.SystemCacertsPath)
	}

	return nil
}
