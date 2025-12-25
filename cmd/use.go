package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/downloader"
	"strigo/logging"
	"strings"

	"github.com/spf13/cobra"
)

var (
	setEnvVar bool
	unsetEnv  bool
)

// getHomeDir returns the user's home directory with proper error handling
func getHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	return home, nil
}

// getShell returns the current shell with a fallback to /bin/bash
func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	return shell
}

func init() {
	useCmd.Flags().BoolVarP(&setEnvVar, "set-env", "e", false, "Set environment variables in shell configuration file (~/.bashrc or ~/.zshrc)")
	useCmd.Flags().BoolVar(&unsetEnv, "unset", false, "Remove environment variables from shell configuration file")
}

var useCmd = &cobra.Command{
	Use:   "use [type] [distribution] [version]",
	Short: "Set a specific SDK version as active",
	Long: `Set a specific SDK version as active. For example:
strigo use jdk temurin 11.0.24_8

This will create a symbolic link to the specified version.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if unsetEnv {
			if len(args) != 1 || (args[0] != "jdk" && args[0] != "node") {
				return fmt.Errorf("\n❌ Invalid arguments for --unset\n\n" +
					"Usage:\n" +
					"  strigo use [jdk|node] --unset")
			}
			return nil
		}

		if len(args) != 3 {
			return fmt.Errorf("\n❌ Invalid number of arguments\n\n" +
				"Usage:\n" +
				"  strigo use [type] [distribution] [version]\n\n" +
				"Example:\n" +
				"  strigo use jdk temurin 11.0.24_8\n\n" +
				"To see installed versions:\n" +
				"  strigo list jdk temurin")
		}
		return nil
	},
	Run: use,
	Example: `  # Use Temurin JDK 11
  strigo use jdk temurin 11.0.24_8

  # Use Corretto JDK 8
  strigo use jdk corretto 8u442b06`,
}

func use(cmd *cobra.Command, args []string) {
	if unsetEnv {
		if err := handleUnset(args[0]); err != nil {
			ExitWithError(err)
		}
		return
	}

	if err := handleUse(args[0], args[1], args[2]); err != nil {
		ExitWithError(err)
	}
}

func getSDKBinPath(basePath string, sdkType string) (string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read installation directory: %w", err)
	}

	// SDK directory selection logic
	var sdkDir string
	dirCount := 0
	
	// Count directories and remember the first one
	for _, entry := range entries {
		if entry.IsDir() {
			dirCount++
			// If it's the first directory, remember it
			if sdkDir == "" {
				sdkDir = entry.Name()
			}
		}
	}
	
	// If multiple directories exist, it's ambiguous
	if dirCount > 1 {
		sdkDir = ""
	}

	if sdkDir == "" {
		return "", fmt.Errorf("could not find %s directory in %s", strings.ToUpper(sdkType), basePath)
	}

	return filepath.Join(basePath, sdkDir), nil
}

func findRcFile() (string, error) {
	// Check if shell_config_path is set in config
	if cfg.General.ShellConfigPath != "" {
		return cfg.General.ShellConfigPath, nil
	}

	// Auto-detect based on current shell
	shell := getShell()
	home, err := getHomeDir()
	if err != nil {
		return "", err
	}

	// List of possible RC files
	var rcFiles []string

	// Determine the order based on the shell
	if strings.HasSuffix(shell, "zsh") {
		rcFiles = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".bashrc"), // fallback
		}
	} else if strings.HasSuffix(shell, "bash") {
		rcFiles = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"), // fallback
		}
	} else {
		// Unrecognized shell, try both
		rcFiles = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
		}
	}

	// Find the first existing RC file
	for _, file := range rcFiles {
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
	}

	return "", fmt.Errorf("no shell configuration file found (.zshrc or .bashrc). Please set shell_config_path in strigo.toml")
}

func handleUnset(sdkType string) error {
	if cfg == nil {
		return fmt.Errorf("configuration is not loaded")
	}

	if sdkType != "jdk" && sdkType != "node" {
		return fmt.Errorf("unset is only supported for JDK and Node.js")
	}

	rcFile, err := findRcFile()
	if err != nil {
		return fmt.Errorf("could not find shell configuration file: %w", err)
	}

	// Expand tilde if present
	expandedPath := rcFile
	if strings.HasPrefix(rcFile, "~") {
		home, err := getHomeDir()
		if err != nil {
			return err
		}
		expandedPath = filepath.Join(home, rcFile[1:])
	}

	// Read the current content
	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", expandedPath, err)
	}

	// Remove the Strigo configuration block
	lines := strings.Split(string(content), "\n")
	var newLines []string
	var removed bool
	inStrigoBlock := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// If we find the Strigo comment
		if strings.Contains(line, fmt.Sprintf("# Added by Strigo - %s configuration", strings.ToUpper(sdkType))) {
			inStrigoBlock = true
			removed = true
			continue
		}
		// Skip all export lines in the Strigo block
		if inStrigoBlock {
			if strings.HasPrefix(strings.TrimSpace(line), "export ") {
				continue
			} else if strings.TrimSpace(line) == "" {
				// Empty line marks end of block
				inStrigoBlock = false
				continue
			}
			// If we encounter a non-export, non-empty line, block has ended
			inStrigoBlock = false
		}
		newLines = append(newLines, line)
	}

	if !removed {
		logging.LogInfo("ℹ️  No Strigo %s configuration found in %s", strings.ToUpper(sdkType), rcFile)
		return nil
	}

	// Write the file
	newContent := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(expandedPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update %s: %w", expandedPath, err)
	}

	logging.LogInfo("✅ Successfully removed Strigo %s configuration from %s", strings.ToUpper(sdkType), expandedPath)
	logging.LogInfo("ℹ️  To apply these changes, run: source %s", expandedPath)

	return nil
}

func handleUse(sdkType, distribution, version string) error {
	if cfg == nil {
		return fmt.Errorf("configuration is not loaded")
	}

	// Check if the SDK type exists
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Build the installation path
	installPath := filepath.Join(cfg.General.SDKInstallDir, sdkTypeConfig.InstallDir, distribution, version)

	// Check if the SDK is installed
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return fmt.Errorf("version %s %s %s is not installed", sdkType, distribution, version)
	}

	// Get the binary path
	sdkPath, err := getSDKBinPath(installPath, sdkType)
	if err != nil {
		return fmt.Errorf("failed to find SDK binary path: %w", err)
	}

	// Create the symbolic link
	linkPath := filepath.Join(cfg.General.SDKInstallDir, fmt.Sprintf("current-%s", sdkType))

	// Remove the existing link if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove existing symbolic link: %w", err)
		}
	}

	// Create the new link
	if err := os.Symlink(sdkPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symbolic link: %w", err)
	}

	logging.LogInfo("✅ Successfully set %s %s version %s as active", sdkType, distribution, version)

	// Load metadata for the installation
	metadata, err := downloader.LoadMetadata(installPath)
	if err != nil {
		logging.LogDebug("⚠️  Failed to load installation metadata: %v", err)
		// Non-fatal, continue with default behavior
	}

	// If --set-env is specified, configure the environment variables
	if setEnvVar {
		if err := configureEnvironment(sdkType, sdkPath, metadata); err != nil {
			return fmt.Errorf("failed to configure environment: %w", err)
		}
	} else {
		if sdkType == "jdk" {
			logging.LogInfo("ℹ️  To use this JDK, set these environment variables:")
			logging.LogInfo("   export JAVA_HOME=%s", sdkPath)
			logging.LogInfo("   export PATH=$JAVA_HOME/bin:$PATH")
			logging.LogInfo("")
			logging.LogInfo("💡 Or use --set-env to set them automatically in your shell configuration")
		} else if sdkType == "node" {
			logging.LogInfo("ℹ️  To use this Node.js version, set these environment variables:")
			logging.LogInfo("   export NODE_HOME=%s", sdkPath)
			logging.LogInfo("   export PATH=$NODE_HOME/bin:$PATH")
			if metadata != nil && metadata.NodeExtraCaCerts != "" {
				logging.LogInfo("   export NODE_EXTRA_CA_CERTS=%s", metadata.NodeExtraCaCerts)
			}
			logging.LogInfo("")
			logging.LogInfo("💡 Or use --set-env to set them automatically in your shell configuration")
		}
	}

	return nil
}

func configureEnvironment(sdkType, sdkPath string, metadata *downloader.SDKMetadata) error {
	// Find the appropriate RC file
	rcFile, err := findRcFile()
	if err != nil {
		return err
	}

	// Expand tilde if present
	expandedPath := rcFile
	if strings.HasPrefix(rcFile, "~") {
		home, err := getHomeDir()
		if err != nil {
			return err
		}
		expandedPath = filepath.Join(home, rcFile[1:])
	}

	// Read the current content
	content, err := os.ReadFile(expandedPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read rc file: %w", err)
	}

	// Prepare the new lines
	var envVar string
	var newConfig string
	if sdkType == "jdk" {
		envVar = "JAVA_HOME"
		newConfig = fmt.Sprintf("\n# Added by Strigo - %s configuration\nexport %s=%s\nexport PATH=$%s/bin:$PATH\n",
			strings.ToUpper(sdkType), envVar, sdkPath, envVar)
	} else if sdkType == "node" {
		envVar = "NODE_HOME"
		if metadata != nil && metadata.NodeExtraCaCerts != "" {
			// Include NODE_EXTRA_CA_CERTS if configured
			newConfig = fmt.Sprintf("\n# Added by Strigo - %s configuration\nexport %s=%s\nexport PATH=$%s/bin:$PATH\nexport NODE_EXTRA_CA_CERTS=%s\n",
				strings.ToUpper(sdkType), envVar, sdkPath, envVar, metadata.NodeExtraCaCerts)
		} else {
			newConfig = fmt.Sprintf("\n# Added by Strigo - %s configuration\nexport %s=%s\nexport PATH=$%s/bin:$PATH\n",
				strings.ToUpper(sdkType), envVar, sdkPath, envVar)
		}
	}

	// Remove the old configuration if it exists
	lines := strings.Split(string(content), "\n")
	var newLines []string
	inStrigoBlock := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// If we find the Strigo comment
		if strings.Contains(line, fmt.Sprintf("# Added by Strigo - %s configuration", strings.ToUpper(sdkType))) {
			inStrigoBlock = true
			continue
		}
		// Skip all export lines in the Strigo block
		if inStrigoBlock {
			if strings.HasPrefix(strings.TrimSpace(line), "export ") {
				continue
			} else if strings.TrimSpace(line) == "" {
				// Empty line marks end of block
				inStrigoBlock = false
				continue
			}
			// If we encounter a non-export, non-empty line, block has ended
			inStrigoBlock = false
		}
		newLines = append(newLines, line)
	}

	// Add the new configuration
	newContent := strings.Join(newLines, "\n") + newConfig

	// Write the new content
	if err := os.WriteFile(expandedPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update rc file: %w", err)
	}

	logging.LogInfo("✅ Successfully configured environment in %s", expandedPath)
	logging.LogInfo("ℹ️  To apply these changes, run: source %s", expandedPath)

	return nil
}
