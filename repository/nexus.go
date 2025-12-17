package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strigo/config"
	"strigo/logging"
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
type NexusClient struct{}

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
	seenVersions := make(map[string]bool) // Pour suivre les versions déjà vues

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

	resp, err := http.Get(requestURL)
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

	// Construire le chemin complet pour la distribution
	distributionPath := repo.Path
	logging.LogDebug("Looking for distribution path: %s", distributionPath)

	for _, item := range data.Items {
		logging.LogDebug("   Path: %s", item.Path)

		// Vérifier si le chemin correspond à la distribution demandée
		if !strings.Contains(item.Path, distributionPath) && distributionPath != "" {
			logging.LogDebug("   Ignoring file: path does not contain %s", distributionPath)
			ignoredFiles = append(ignoredFiles, item.Path)
			continue
		}

		versionName := ExtractVersionName(item.Path)
		if versionName != "" {
			logging.LogDebug("   Extracted version: %s from path: %s", versionName, item.Path)
			// Vérifier si cette version a déjà été vue
			if !seenVersions[versionName] {
				seenVersions[versionName] = true
				sdkAsset := SDKAsset{
					Version:     versionName,
					DownloadUrl: item.DownloadUrl,
					Filename:    versionName,
					// Size sera ajouté plus tard si nécessaire
				}
				sdkAssets = append(sdkAssets, sdkAsset)
			}
		} else {
			ignoredFiles = append(ignoredFiles, item.Path)
		}
	}

	if len(ignoredFiles) > 0 {
		logging.LogDebug("❌ Ignored files:")
		for _, f := range ignoredFiles {
			logging.LogDebug("   - %s", f)
		}
	}

	// Filtrer les versions si un filtre est spécifié
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

	// Trier les versions
	sort.Slice(sdkAssets, func(i, j int) bool {
		return sdkAssets[i].Version > sdkAssets[j].Version
	})

	return sdkAssets, nil
}

// ExtractVersionName extracts the versioned filename from a Nexus path.
func ExtractVersionName(path string) string {
	logging.LogDebug("Extracting version from path: %s", path)

	// Handle different naming patterns
	patterns := []string{
		`corretto-(\d+\.\d+\.\d+\.\d+)`,             // For Corretto: 11.0.26.4.1
		`jdk-(\d+\.\d+\.\d+_\d+)`,                   // For Temurin: 11.0.26_4
		`jdk_x64_linux_hotspot_(\d+\.\d+\.\d+_\d+)`, // Alternative Temurin pattern
		`(\d+u\d+\w+)`,                              // For older versions: 8u442b06
		`node-v(\d+\.\d+\.\d+)-linux-x64`,           // For Node.js: node-v22.13.1-linux-x64
		`amazon-corretto-(\d+\.\d+\.\d+\.\d+)`,      // For Amazon Corretto
		`zulu\d+\.\d+\.\d+-ca-jdk(\d+\.\d+\.\d+)`,   // For Zulu
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(path); len(matches) > 1 {
			logging.LogDebug("  Found version %s using pattern %s", matches[1], pattern)
			return matches[1]
		}
	}

	// Fallback: try to extract version from path components
	parts := strings.Split(path, "/")
	for _, part := range parts {
		logging.LogDebug("  Checking path component: %s", part)

		// Look for version-like patterns in path components
		if strings.HasPrefix(part, "v") {
			version := strings.TrimPrefix(part, "v")
			if _, err := strconv.Atoi(strings.Split(version, ".")[0]); err == nil {
				logging.LogDebug("  Found version in path component: %s", version)
				return version
			}
		}

		// Check for version in the format jdk-X.Y.Z or jdkX.Y.Z
		if strings.Contains(part, "jdk") {
			version := strings.TrimPrefix(strings.TrimPrefix(part, "jdk-"), "jdk")
			if _, err := strconv.Atoi(strings.Split(version, ".")[0]); err == nil {
				logging.LogDebug("  Found version in JDK component: %s", version)
				return version
			}
		}
	}

	logging.LogDebug("  No version found in path")
	return ""
}

// FindAssetByVersion helps locate a specific asset in the version map
func FindAssetByVersion(versionMap map[string][]NexusAsset, targetVersion string) *NexusAsset {
	for _, assets := range versionMap {
		for _, asset := range assets {
			// Check if the file name contains the target version
			if strings.Contains(asset.Path, targetVersion) {
				return &asset
			}
		}
	}
	return nil
}
