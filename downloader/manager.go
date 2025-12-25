package downloader

import (
	"fmt"
	"path/filepath"
	"strigo/downloader/cache"
	"strigo/downloader/core"
	"strigo/downloader/jdk"
	"strigo/downloader/network"
	"strigo/logging"
)

// Manager orchestrates the download and installation process
type Manager struct {
	network     *network.Client
	extractor   *Extractor
	cache       *cache.Manager
	validator   *core.Validator
	certificates *jdk.CertificateManager
}

// NewManager creates a new Manager instance
func NewManager() *Manager {
	return &Manager{
		network:      network.NewClient(),
		extractor:    NewExtractor(),
		cache:        cache.NewManager(),
		validator:    core.NewValidator(),
		certificates: jdk.NewCertificateManager(),
	}
}

// NewManagerWithAuth creates a new Manager instance with authentication
func NewManagerWithAuth(username, password string) *Manager {
	return &Manager{
		network:      network.NewClientWithAuth(username, password),
		extractor:    NewExtractor(),
		cache:        cache.NewManager(),
		validator:    core.NewValidator(),
		certificates: jdk.NewCertificateManager(),
	}
}

// DownloadAndExtract handles the complete download and installation process
func (m *Manager) DownloadAndExtract(opts core.DownloadOptions) error {
	logging.LogDebug("🔍 Starting installation process for %s %s %s", opts.SDKType, opts.Distribution, opts.Version)

	// Check file size
	fileSize, err := m.network.GetFileSize(opts.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	// Validate available space
	if err := m.validator.ValidateSpace(fileSize, opts.CacheDir); err != nil {
		return fmt.Errorf("cache directory space check failed: %w", err)
	}
	if err := m.validator.ValidateSpace(fileSize, filepath.Dir(opts.InstallPath)); err != nil {
		return fmt.Errorf("install directory space check failed: %w", err)
	}

	// Prepare cache
	cachePath, err := m.cache.PrepareCacheDirectory(opts.SDKType, opts.Distribution, opts.Version, opts.CacheDir)
	if err != nil {
		return fmt.Errorf("failed to prepare cache: %w", err)
	}

	// Download file
	cacheFile := filepath.Join(cachePath, filepath.Base(opts.DownloadURL))
	if err := m.network.DownloadFile(opts.DownloadURL, cacheFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Validate and create installation directory
	if err := m.validator.ValidateDirectories(opts.InstallPath); err != nil {
		return fmt.Errorf("failed to prepare installation directory: %w", err)
	}

	// Extract archive
	if err := m.extractor.Extract(cacheFile, opts.InstallPath); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Clean cache if needed
	if err := m.cache.CleanupCache(cachePath, opts.KeepCache); err != nil {
		logging.LogDebug("⚠️ Cache cleanup failed: %v", err)
	}

	// Certificate injection is now handled in cmd/install.go after extraction
	// This allows for proper path detection and optional certificate management

	logging.LogInfo("✅ Successfully extracted %s %s version %s", opts.SDKType, opts.Distribution, opts.Version)
	return nil
}
