package downloader

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SDKMetadata contains metadata about an installed SDK
type SDKMetadata struct {
	SDKType      string `json:"sdk_type"`
	Distribution string `json:"distribution"`
	Version      string `json:"version"`

	// Node.js specific
	NodeExtraCaCerts string `json:"node_extra_ca_certs,omitempty"` // Path to PEM bundle
}

// SaveMetadata writes metadata to .strigo-metadata.json in the installation directory
func SaveMetadata(installPath string, metadata SDKMetadata) error {
	metadataPath := filepath.Join(installPath, ".strigo-metadata.json")

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// LoadMetadata reads metadata from .strigo-metadata.json in the installation directory
func LoadMetadata(installPath string) (*SDKMetadata, error) {
	metadataPath := filepath.Join(installPath, ".strigo-metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No metadata file, not an error
		}
		return nil, err
	}

	var metadata SDKMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}
