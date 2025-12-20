package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/downloader"
	"strigo/downloader/core"
	"strigo/logging"
	"strigo/repository"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [type] [distribution] [version]",
	Short: "Install a specific SDK version",
	Long: `Install a specific SDK version. For example:
	strigo install jdk temurin 11.0.24_8
	strigo install jdk corretto 8u442b06

Available SDK types:
	jdk     Java Development Kit

Available distributions for jdk:
	temurin    Eclipse Temurin (AdoptOpenJDK)
	corretto   Amazon Corretto`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("\n❌ Invalid number of arguments\n\n" +
				"Usage:\n" +
				"  strigo install [type] [distribution] [version]\n\n" +
				"Example:\n" +
				"  strigo install jdk temurin 11.0.24_8\n\n" +
				"To see available versions:\n" +
				"  strigo available jdk temurin")
		}
		return nil
	},
	Run: install,
	Example: `  # Install Temurin JDK 11
  strigo install jdk temurin 11.0.24_8

  # Install Corretto JDK 8
  strigo install jdk corretto 8u442b06

  # To see available versions:
  strigo available jdk temurin`,
}

func install(cmd *cobra.Command, args []string) {
	sdkType := args[0]
	distribution := args[1]
	version := args[2]

	if err := handleInstall(sdkType, distribution, version); err != nil {
		logging.LogError("❌ Error executing command: %v", err)
		return
	}
}

func handleInstall(sdkType, distribution, version string) error {
	logging.LogDebug("🔧 Starting installation of %s %s version %s", sdkType, distribution, version)

	// Check if the SDK type exists
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		logging.LogError("❌ SDK type %s not found in configuration", sdkType)
		return fmt.Errorf("SDK type %s not found", sdkType)
	}

	// Check if the distribution exists
	sdkRepo, exists := cfg.SDKRepositories[distribution]
	if !exists {
		logging.LogError("❌ Distribution %s not found in configuration", distribution)
		return fmt.Errorf("distribution %s not found", distribution)
	}

	// Verify that the distribution's type matches the requested type
	if sdkRepo.Type != sdkTypeConfig.Type {
		logging.LogError("❌ Distribution %s is not of type %s", distribution, sdkType)
		return fmt.Errorf("distribution %s is not of type %s", distribution, sdkType)
	}

	// Get registry information
	registry, exists := cfg.Registries[sdkRepo.Registry]
	if !exists {
		logging.LogError("❌ Registry %s not found in configuration", sdkRepo.Registry)
		return fmt.Errorf("registry %s not found", sdkRepo.Registry)
	}

	// Fetch available versions with filter
	assets, err := repository.FetchAvailableVersions(sdkRepo, registry, version, true, cfg.General.PatternsFile) // true to remove display
	if err != nil {
		logging.LogError("❌ Failed to fetch versions: %v", err)
		return fmt.Errorf("failed to fetch versions: %w", err)
	}

	// Find exact version match
	var matchedAsset *repository.SDKAsset
	for i := range assets {
		if assets[i].Version == version {
			matchedAsset = &assets[i]
			break
		}
	}

	if matchedAsset == nil {
		logging.LogError("❌ Version %s not found", version)
		logging.LogInfo("💡 Use 'strigo available %s %s' to see available versions", sdkType, distribution)
		return fmt.Errorf("version %s not found", version)
	}

	logging.LogInfo("✅ Found version %s, preparing for installation...", version)

	// Get installation path
	installPath, err := GetInstallPath(cfg, sdkType, distribution, version)
	if err != nil {
		logging.LogError("❌ Failed to get installation path: %v", err)
		return fmt.Errorf("failed to get installation path: %w", err)
	}

	// Check if already installed
	if _, err := os.Stat(installPath); err == nil {
		logging.LogError("❌ Version %s is already installed at %s", version, installPath)
		return fmt.Errorf("version %s is already installed", version)
	}

	// Create installation directory
	if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
		logging.LogError("❌ Failed to create installation directory: %v", err)
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Prepare certificate configuration
	certConfig := core.CertConfig{
		JDKSecurityPath:   cfg.General.JDKSecurityPath,
		SystemCacertsPath: cfg.General.SystemCacertsPath,
	}

	// Download and extract - create manager with auth if credentials are provided
	var manager *downloader.Manager
	if registry.Username != "" && registry.Password != "" {
		logging.LogDebug("🔐 Creating download manager with authentication")
		manager = downloader.NewManagerWithAuth(registry.Username, registry.Password)
	} else {
		manager = downloader.NewManager()
	}

	opts := core.DownloadOptions{
		DownloadURL:  matchedAsset.DownloadUrl,
		CacheDir:     cfg.General.CacheDir,
		InstallPath:  installPath,
		SDKType:      sdkType,
		Distribution: distribution,
		Version:      version,
		KeepCache:    cfg.General.KeepCache,
		CertConfig:   certConfig,
		Username:     registry.Username,
		Password:     registry.Password,
	}
	err = manager.DownloadAndExtract(opts)

	if err != nil {
		logging.LogError("❌ Installation failed: %v", err)
		// Cleanup on failure
		os.RemoveAll(installPath)
		return fmt.Errorf("installation failed: %w", err)
	}

	// For JDKs, manage certificates
	if sdkType == "jdk" {
		// Find the extracted JDK folder
		entries, err := os.ReadDir(installPath)
		if err != nil {
			return fmt.Errorf("failed to read installation directory: %w", err)
		}

		// JDK directory selection logic
		var jdkDir string
		dirCount := 0
		
		// Count directories and remember the first one
		for _, entry := range entries {
			if entry.IsDir() {
				dirCount++
				// If it's the first directory, remember it
				if jdkDir == "" {
					jdkDir = entry.Name()
				}
			}
		}
		
		// If multiple directories exist, it's ambiguous
		if dirCount > 1 {
			jdkDir = ""
		}

		if jdkDir == "" {
			return fmt.Errorf("could not find JDK directory in %s", installPath)
		}

		// Use the full path for certificates
		jdkPath := filepath.Join(installPath, jdkDir)
		jdkSecPath := filepath.Join(jdkPath, cfg.General.JDKSecurityPath)

		// 1. Remove default JDK certificates
		logging.LogDebug("🗑️ Removing default JDK certificates...")
		if err := os.RemoveAll(jdkSecPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove default certificates: %w", err)
		}

		// 2. Create a symbolic link to system certificates
		logging.LogDebug("🔗 Creating link to system certificates...")
		if err := os.MkdirAll(filepath.Dir(jdkSecPath), 0755); err != nil {
			return fmt.Errorf("failed to create security directory: %w", err)
		}

		if err := os.Symlink(cfg.General.SystemCacertsPath, jdkSecPath); err != nil {
			return fmt.Errorf("failed to create symlink to system certificates: %w", err)
		}
		logging.LogInfo("✅ Successfully linked system certificates")
	}

	logging.LogInfo("✅ Successfully installed %s %s version %s", sdkType, distribution, version)
	logging.LogInfo("📂 Installation path: %s", installPath)
	logging.LogInfo("ℹ️  To set this version as active, run: strigo use %s %s %s", sdkType, distribution, version)

	return nil
}
