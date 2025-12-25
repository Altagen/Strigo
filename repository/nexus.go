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
// It handles pagination using continuationToken to retrieve all assets.
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

	// Collect all items across all pages using pagination
	var allItems []NexusAsset
	continuationToken := ""
	pageCount := 0

	for {
		pageCount++
		logging.LogDebug("📄 Fetching page %d from Nexus...", pageCount)

		// Build request URL with continuation token if present
		requestURL := apiURL
		if continuationToken != "" {
			requestURL = fmt.Sprintf("%s&continuationToken=%s", apiURL, url.QueryEscape(continuationToken))
		}

		logging.LogDebug("🔍 Nexus API URL: %s", requestURL)

		// Create HTTP request
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request: %v", err)
		}

		// Add Basic Auth if credentials are provided
		if registry.Username != "" && registry.Password != "" {
			req.SetBasicAuth(registry.Username, registry.Password)
			if pageCount == 1 {
				logging.LogDebug("🔐 Using Basic Auth with username: %s", registry.Username)
			}
		}

		// Execute request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to query Nexus API: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("nexus API returned %d: Check if the path %s exists in Nexus", resp.StatusCode, repo.Path)
		}

		// Parse JSON response
		var data struct {
			Items             []NexusAsset `json:"items"`
			ContinuationToken string       `json:"continuationToken,omitempty"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode JSON response: %v", err)
		}
		resp.Body.Close()

		logging.LogDebug("📦 Received %d items on page %d", len(data.Items), pageCount)

		// Accumulate items from this page
		allItems = append(allItems, data.Items...)

		// Check if there are more pages
		if data.ContinuationToken != "" {
			continuationToken = data.ContinuationToken
			logging.LogDebug("➡️  More pages available, continuing pagination...")
		} else {
			logging.LogDebug("✅ Pagination complete. Total items: %d", len(allItems))
			break
		}
	}

	// Process all collected items
	logging.LogDebug("🔍 Processing %d total items from Nexus", len(allItems))

	// Build full path for distribution
	distributionPath := repo.Path
	logging.LogDebug("Looking for distribution path: %s", distributionPath)

	// Normalize path prefix for matching
	// Ensure it starts with "/" and doesn't end with "/"
	pathPrefix := "/" + strings.TrimPrefix(distributionPath, "/")
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}

	for _, item := range allItems {
		logging.LogDebug("   Path: %s", item.Path)

		// Check if the path starts with the requested distribution path
		// This ensures exact prefix matching (e.g., "/jdk/adoptium/temurin/" matches
		// "/jdk/adoptium/temurin/17/..." but NOT "/jdk/adoptium/temurin-test/...")
		if distributionPath != "" && !strings.HasPrefix(item.Path, pathPrefix) {
			logging.LogDebug("   Ignoring file: path does not start with %s", pathPrefix)
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
