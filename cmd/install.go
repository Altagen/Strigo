package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/downloader"
	"strigo/downloader/core"
	"strigo/downloader/jdk"
	"strigo/logging"
	"strigo/repository"

	"github.com/spf13/cobra"
)

var (
	jdkCacertsPath     string
	jdkCacertsPassword string
	nodeExtraCaCerts   string
)

func init() {
	installCmd.Flags().StringVar(&jdkCacertsPath, "jdk-cacerts-path", "", "Override cacerts path in JDK (e.g., 'jre/lib/security/cacerts' for Java 8)")
	installCmd.Flags().StringVar(&jdkCacertsPassword, "jdk-cacerts-password", "", "Override cacerts password (default: 'changeit', use '' for password-less PKCS12)")
	installCmd.Flags().StringVar(&nodeExtraCaCerts, "node-extra-ca-certs", "", "Path to PEM bundle for Node.js extra CA certificates (supports multiple certificates)")
}

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
	assets, err := repository.FetchAvailableVersions(sdkRepo, registry, version, true, GetPatternsFilePath()) // true to remove display
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

	// For JDKs, inject custom certificates if configured
	if sdkType == "jdk" && len(cfg.General.CustomCertificates) > 0 {
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
			logging.LogDebug("⚠️  Could not find JDK directory, skipping certificate injection")
		} else {
			// Use the full path for the JDK root
			jdkPath := filepath.Join(installPath, jdkDir)

			// Determine path override (CLI takes precedence over config)
			pathOverride := jdkCacertsPath
			if pathOverride == "" {
				pathOverride = cfg.General.JDKCacertsOverride
			}

			// Determine password (CLI takes precedence over config, default to "changeit")
			password := jdkCacertsPassword
			if password == "" {
				password = cfg.General.JDKCacertsPassword
			}
			if password == "" {
				password = "changeit"
			}

			// Create certificate manager and inject certificates
			certManager := jdk.NewCertificateManager()
			err := certManager.InjectCertificates(
				jdkPath,
				cfg.General.CustomCertificates,
				pathOverride,
				password,
			)

			if err != nil {
				// Non-fatal: log warning but continue installation
				logging.LogDebug("⚠️  Certificate injection failed: %v", err)
				logging.LogInfo("ℹ️  JDK installation is complete but custom certificates were not injected")
				logging.LogInfo("💡 You can manually add certificates using Java's keytool if needed")
			}
		}
	} else if sdkType == "jdk" {
		logging.LogDebug("📋 No custom certificates configured, JDK will use default certificate store")
	}

	// Handle Node.js certificate configuration
	if sdkType == "node" && nodeExtraCaCerts != "" {
		// Validate the certificate path exists
		if _, err := os.Stat(nodeExtraCaCerts); os.IsNotExist(err) {
			logging.LogDebug("⚠️  Node.js certificate file not found: %s", nodeExtraCaCerts)
			logging.LogInfo("ℹ️  Node.js installed but certificate path is invalid")
		} else {
			logging.LogDebug("📋 Node.js will use extra CA certificates from: %s", nodeExtraCaCerts)
		}
	}

	// Save metadata for the installation
	metadata := downloader.SDKMetadata{
		SDKType:      sdkType,
		Distribution: distribution,
		Version:      version,
	}

	// Add Node.js specific metadata if provided
	if sdkType == "node" && nodeExtraCaCerts != "" {
		// Expand tilde if present and convert to absolute path
		expandedPath := nodeExtraCaCerts
		if len(nodeExtraCaCerts) > 0 && nodeExtraCaCerts[0] == '~' {
			home, err := os.UserHomeDir()
			if err == nil {
				expandedPath = filepath.Join(home, nodeExtraCaCerts[1:])
			}
		}

		// Make path absolute
		absPath, err := filepath.Abs(expandedPath)
		if err == nil {
			metadata.NodeExtraCaCerts = absPath
		} else {
			metadata.NodeExtraCaCerts = expandedPath
		}
	}

	if err := downloader.SaveMetadata(installPath, metadata); err != nil {
		logging.LogDebug("⚠️  Failed to save installation metadata: %v", err)
		// Non-fatal, continue
	}

	logging.LogInfo("✅ Successfully installed %s %s version %s", sdkType, distribution, version)
	logging.LogInfo("📂 Installation path: %s", installPath)
	logging.LogInfo("ℹ️  To set this version as active, run: strigo use %s %s %s", sdkType, distribution, version)

	return nil
}
