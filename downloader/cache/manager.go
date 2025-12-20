package cache

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strigo/logging"
)

// Manager handles cache for downloaded files
type Manager struct{}

// NewManager creates a new Manager instance
func NewManager() *Manager {
	return &Manager{}
}

// PrepareCacheDirectory prepares the cache directory
func (m *Manager) PrepareCacheDirectory(sdkType, distribution, version, cacheDir string) (string, error) {
	cachePath := filepath.Join(cacheDir, sdkType, distribution, version)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return cachePath, nil
}

// CleanupCache cleans up the cache if needed
func (m *Manager) CleanupCache(cachePath string, keepCache bool) error {
	if !keepCache {
		logging.LogDebug("🧹 Cleaning up cache directory: %s", cachePath)
		return m.cleanupCacheDirectory(cachePath)
	}
	return nil
}

func (m *Manager) cleanupCacheDirectory(cachePath string) error {
	if err := os.RemoveAll(cachePath); err != nil {
		return fmt.Errorf("failed to remove cache directory: %w", err)
	}

	// Clean up empty parent directories
	parent := filepath.Dir(cachePath)
	for parent != filepath.Dir(parent) {
		if empty, err := m.isDirEmpty(parent); err != nil || !empty {
			break
		}
		if err := os.Remove(parent); err != nil {
			break
		}
		parent = filepath.Dir(parent)
	}
	return nil
}

func (m *Manager) isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		return false, nil // Directory not empty
	}
	if errors.Is(err, io.EOF) {
		return true, nil // Directory empty
	}
	return false, fmt.Errorf("failed to check if directory is empty: %w", err)
}
