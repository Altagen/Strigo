package repository

import (
	"fmt"
	"sort"
	"strconv"
	"strigo/config"
	"strigo/logging"
	"strigo/repository/version"
)

// RepositoryClient defines the interface for fetching available versions
type RepositoryClient interface {
	GetAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string) ([]SDKAsset, error)
}

// FetchAvailableVersions fetches available versions with optional JSON output control
// opts[0]: jsonOutput (bool) - whether to suppress display output
// opts[1]: patternsFilePath (string) - custom patterns file path (empty for default)
func FetchAvailableVersions(repo config.SDKRepository, registry config.Registry, versionFilter string, opts ...interface{}) ([]SDKAsset, error) {
	var client RepositoryClient

	// Parse options
	jsonOutput := false
	patternsFilePath := ""

	if len(opts) > 0 {
		if b, ok := opts[0].(bool); ok {
			jsonOutput = b
		}
	}
	if len(opts) > 1 {
		if s, ok := opts[1].(string); ok {
			patternsFilePath = s
		}
	}

	switch registry.Type {
	case "nexus":
		nexusClient, err := NewNexusClientWithConfig(patternsFilePath)
		if err != nil {
			logging.LogError("❌ Failed to initialize Nexus client: %v", err)
			return nil, fmt.Errorf("failed to initialize Nexus client: %w", err)
		}
		client = nexusClient
	default:
		logging.LogError("❌ Unsupported repository type: %s", registry.Type)
		return nil, fmt.Errorf("unsupported repository type: %s", registry.Type)
	}

	assets, err := client.GetAvailableVersions(repo, registry, versionFilter)
	if err != nil {
		return nil, err
	}

	// If not in JSON mode, display versions
	if !jsonOutput {
		displayVersions(assets)
	}

	return assets, nil
}

// displayVersions handles the user-friendly output
func displayVersions(assets []SDKAsset) {
	// Create a map to group by major version
	versionGroups := make(map[string][]string)

	// Extract major version and group
	for _, asset := range assets {
		majorVersion := ExtractMajorVersion(asset.Version)
		versionGroups[majorVersion] = append(versionGroups[majorVersion], asset.Version)
	}

	// Get numerically sorted major versions
	var majorVersions []int
	for major := range versionGroups {
		if num, err := strconv.Atoi(major); err == nil {
			majorVersions = append(majorVersions, num)
		}
	}
	sort.Ints(majorVersions)

	logging.LogOutput("🔹 Available versions:")
	for _, majorNum := range majorVersions {
		major := strconv.Itoa(majorNum)
		versions := versionGroups[major]

		// Sort versions in each group
		sort.Slice(versions, func(i, j int) bool {
			return CompareVersions(versions[i], versions[j])
		})

		logging.LogOutput("  - %s:", major)
		for _, version := range versions {
			logging.LogOutput("    ✅ %s", version)
		}
	}

	logging.LogOutput("\n💡 To install a specific version:")
	logging.LogOutput("   strigo install jdk [distribution] [version]")
}

// ExtractMajorVersion is deprecated. Use version.ExtractMajor instead.
// Kept for backward compatibility with existing code.
func ExtractMajorVersion(versionStr string) string {
	result := version.ExtractMajor(versionStr)
	// Maintain backward compatibility: return "unknown" if empty
	if result == "" {
		return "unknown"
	}
	return result
}

// CompareVersions is deprecated. Use version.CompareVersions instead.
// Kept for backward compatibility with existing code.
func CompareVersions(v1, v2 string) bool {
	return version.CompareVersions(v1, v2)
}
