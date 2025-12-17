package cache

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strigo/logging"
)

// Manager gère le cache des fichiers téléchargés
type Manager struct{}

// NewManager crée une nouvelle instance de Manager
func NewManager() *Manager {
	return &Manager{}
}

// PrepareCacheDirectory prépare le répertoire de cache
func (m *Manager) PrepareCacheDirectory(sdkType, distribution, version, cacheDir string) (string, error) {
	cachePath := filepath.Join(cacheDir, sdkType, distribution, version)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return cachePath, nil
}

// CleanupCache nettoie le cache si nécessaire
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

	// Nettoyer les répertoires parents vides
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
		return false, nil // Répertoire non vide
	}
	if errors.Is(err, io.EOF) {
		return true, nil // Répertoire vide
	}
	return false, fmt.Errorf("failed to check if directory is empty: %w", err)
}
