package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strigo/config"
	"strigo/logging"
	"strigo/repository/version"
	"strings"
)

// SDKAsset represents an available version of an SDK
type SDKAsset struct {
	Version     string `json:"version"`
	DownloadUrl string `json:"downloadUrl"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
}

// NexusClient implements RepositoryClient for Nexus repositories
type NexusClient struct {
	parser *version.Parser
}

// NewNexusClient creates a new NexusClient with an initialized parser
// Uses default patterns file path (strigopatterns.toml)
func NewNexusClient() (*NexusClient, error) {
	return NewNexusClientWithConfig("")
}

// NewNexusClientWithConfig creates a new NexusClient with a custom patterns file path
// patternsFilePath can be empty to use default (strigopatterns.toml)
func NewNexusClientWithConfig(patternsFilePath string) (*NexusClient, error) {
	parser, err := version.NewParser(patternsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize version parser: %w", err)
	}

	return &NexusClient{
		parser: parser,
	}, nil
}

// NexusAsset represents an asset returned by Nexus API
type NexusAsset struct {
	Path        string            `json:"path"`
	DownloadUrl string            `json:"downloadUrl"`
	Checksum    map[string]string `json:"checksum"`
}

// GetAvailableVersions fetches available versions of a JDK from a Nexus repository.
func (c *NexusClient) GetAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string) ([]SDKAsset, error) {
	var sdkAssets []SDKAsset
	var ignoredFiles []string
	seenVersions := make(map[string]bool) // To track already seen versions

	// Ensure apiURL is correctly formatted and replace placeholders
	logging.LogDebug("🔍 Registry API URL: %s", registry.APIURL)
	logging.LogDebug("🔍 Repository: %s", repo.Repository)
	logging.LogDebug("🔍 Path: %s", repo.Path)

	apiURL := strings.ReplaceAll(registry.APIURL, "{repository}", repo.Repository)
	logging.LogDebug("🔍 API URL after repository replacement: %s", apiURL)

	// Build final request URL with proper URL encoding
	escapedPath := url.QueryEscape(repo.Path)
	requestURL := fmt.Sprintf("%s&path=%s", apiURL, escapedPath)

	logging.LogDebug("🔍 Final Nexus API URL: %s", requestURL)

	// Create HTTP request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Add Basic Auth if credentials are provided
	if registry.Username != "" && registry.Password != "" {
		req.SetBasicAuth(registry.Username, registry.Password)
		logging.LogDebug("🔐 Using Basic Auth with username: %s", registry.Username)
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Nexus API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nexus API returned %d: Check if the path %s exists in Nexus", resp.StatusCode, repo.Path)
	}

	// Parse JSON response
	var data struct {
		Items []NexusAsset `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	logging.LogDebug("🔍 Raw items from Nexus:")
	logging.LogDebug("Found %d items in response", len(data.Items))
	for _, item := range data.Items {
		logging.LogDebug("Item path: %s, downloadUrl: %s", item.Path, item.DownloadUrl)
	}

	// Build full path for distribution
	distributionPath := repo.Path
	logging.LogDebug("Looking for distribution path: %s", distributionPath)

	for _, item := range data.Items {
		logging.LogDebug("   Path: %s", item.Path)

		// Check if the path corresponds to the requested distribution
		if !strings.Contains(item.Path, distributionPath) && distributionPath != "" {
			logging.LogDebug("   Ignoring file: path does not contain %s", distributionPath)
			ignoredFiles = append(ignoredFiles, item.Path)
			continue
		}

		// Use the parser to extract version
		versionName, patternName, err := c.parser.ExtractVersionByType(item.Path, repo.Type)
		if err != nil {
			logging.LogDebug("   No version extracted: %v", err)
			ignoredFiles = append(ignoredFiles, item.Path)
			continue
		}

		logging.LogDebug("   Extracted version: %s from path: %s (pattern: %s)", versionName, item.Path, patternName)

		// Check if this version has already been seen
		if !seenVersions[versionName] {
			seenVersions[versionName] = true
			sdkAsset := SDKAsset{
				Version:     versionName,
				DownloadUrl: item.DownloadUrl,
				Filename:    versionName,
				// Size will be added later if needed
			}
			sdkAssets = append(sdkAssets, sdkAsset)
		}
	}

	if len(ignoredFiles) > 0 {
		logging.LogDebug("❌ Ignored files:")
		for _, f := range ignoredFiles {
			logging.LogDebug("   - %s", f)
		}
	}

	// Filter versions if a filter is specified
	if versionFilter != "" {
		var filteredAssets []SDKAsset
		for _, asset := range sdkAssets {
			if strings.Contains(asset.Version, versionFilter) {
				filteredAssets = append(filteredAssets, asset)
			}
		}
		sdkAssets = filteredAssets
	}

	if len(sdkAssets) == 0 {
		if versionFilter != "" {
			return nil, fmt.Errorf("no version %s found for %s", versionFilter, repo.Path)
		}
		return nil, fmt.Errorf("no versions found for %s", repo.Path)
	}

	// Sort versions
	sort.Slice(sdkAssets, func(i, j int) bool {
		return sdkAssets[i].Version > sdkAssets[j].Version
	})

	return sdkAssets, nil
}
