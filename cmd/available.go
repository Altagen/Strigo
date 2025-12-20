package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strigo/logging"
	"strigo/repository"
	"strigo/repository/version"
	"strings"

	"github.com/spf13/cobra"
)

// Structures for JSON output
type AvailableOutput struct {
	Types         []string              `json:"types,omitempty"`
	Distributions []string              `json:"distributions,omitempty"`
	Versions      []repository.SDKAsset `json:"versions,omitempty"`
	Error         string                `json:"error,omitempty"`
}

// availableCmd represents the available command
var availableCmd = &cobra.Command{
	Use:   "available [type] <distribution> [version]",
	Short: "List available versions of a specific SDK",
	Long: `List available versions of a specific SDK.
Examples:
  strigo available                  # List all available SDK types
  strigo available jdk             # List all available JDK distributions
  strigo available jdk temurin     # List all Temurin JDK versions
  strigo available jdk temurin 11  # List Temurin JDK versions containing "11"`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Simple validation - config is not loaded yet
		if len(args) > 3 {
			return fmt.Errorf("too many arguments. Use 'strigo available --help' for usage")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			return fmt.Errorf("configuration is not loaded")
		}

		output := &AvailableOutput{}

		// If no arguments, display available SDK types
		if len(args) == 0 {
			return handleNoArgs(output)
		}

		sdkType := args[0]

		// Validate SDK type
		validTypes := getValidSDKTypes()
		if !contains(validTypes, sdkType) {
			return fmt.Errorf("invalid SDK type '%s'. Available types: %s", sdkType, strings.Join(validTypes, ", "))
		}

		// If only type is provided, display distributions
		if len(args) == 1 {
			return handleTypeOnly(sdkType, output)
		}

		distribution := args[1]

		// Validate distribution
		validDists := getValidDistributions(sdkType)
		if !contains(validDists, distribution) {
			return fmt.Errorf("invalid distribution '%s' for type '%s'. Available distributions: %s",
				distribution, sdkType, strings.Join(validDists, ", "))
		}

		var versionFilter string
		if len(args) > 2 {
			versionFilter = args[2]
		}

		return handleFullCommand(sdkType, distribution, versionFilter, output)
	},
}

// Utility functions
func getValidSDKTypes() []string {
	if cfg == nil {
		return []string{}
	}
	types := make([]string, 0, len(cfg.SDKTypes))
	for sdkType := range cfg.SDKTypes {
		types = append(types, sdkType)
	}
	sort.Strings(types)
	return types
}

func getValidDistributions(sdkType string) []string {
	if cfg == nil {
		return []string{}
	}
	var dists []string
	for name, repo := range cfg.SDKRepositories {
		if repo.Type == sdkType {
			dists = append(dists, name)
		}
	}
	sort.Strings(dists)
	return dists
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Handlers for each use case
func handleNoArgs(output *AvailableOutput) error {
	types := getValidSDKTypes()
	output.Types = types

	if len(types) > 0 {
		logging.LogOutput("Available SDK types:")
		logging.LogOutput("─────────────────────")
		for _, sdkType := range types {
			logging.LogOutput("✅ %s", sdkType)
		}
		logging.LogOutput("")
	} else {
		logging.LogOutput("No SDK types available")
	}
	return nil
}

func handleTypeOnly(sdkType string, output *AvailableOutput) error {
	for name, repo := range cfg.SDKRepositories {
		if repo.Type == sdkType {
			output.Distributions = append(output.Distributions, name)
		}
	}

	if len(output.Distributions) > 0 {
		logging.LogOutput("Available %s distributions:", sdkType)
		logging.LogOutput("─────────────────────────")
		for _, dist := range output.Distributions {
			logging.LogOutput("✅ %s", dist)
		}
	}
	return nil
}

// ExtractMajorVersion is deprecated. Use version.ExtractMajor instead.
// Kept for backward compatibility with existing code.
func ExtractMajorVersion(versionStr string) string {
	logging.LogDebug("Extracting major version from: %s", versionStr)

	result := version.ExtractMajor(versionStr)

	if result == "" {
		logging.LogDebug("No major version found")
	} else {
		logging.LogDebug("Found major version: %s", result)
	}

	return result
}

func handleFullCommand(sdkType, distribution, versionFilter string, output *AvailableOutput) error {
	// Check if the distribution exists
	sdkRepo, exists := cfg.SDKRepositories[distribution]
	if !exists {
		err := fmt.Errorf("distribution %s not found in configuration", distribution)
		logging.LogError("❌ %v", err)
		return nil
	}

	// Get registry information
	registry, exists := cfg.Registries[sdkRepo.Registry]
	if !exists {
		err := fmt.Errorf("registry %s not found in configuration", sdkRepo.Registry)
		logging.LogError("❌ %v", err)
		return nil
	}

	// Fetch available versions
	versions, err := repository.FetchAvailableVersions(sdkRepo, registry, "", true, GetPatternsFilePath())
	if err != nil {
		logging.LogError("❌ %v", err)
		return nil
	}

	logging.LogDebug("Found %d versions before filtering", len(versions))

	// Collect all available major versions
	allMajorVersions := make(map[string]bool)
	for _, v := range versions {
		logging.LogDebug("Version before filtering: %s", v.Version)
		majorVersion := ExtractMajorVersion(v.Version)
		if majorVersion != "" {
			allMajorVersions[majorVersion] = true
		}
	}

	// Convert to slice and sort
	var availableMajors []int
	for major := range allMajorVersions {
		if num, err := strconv.Atoi(major); err == nil {
			availableMajors = append(availableMajors, num)
		}
	}
	sort.Ints(availableMajors)

	// Filter versions if a filter is specified
	if versionFilter != "" {
		var filteredVersions []repository.SDKAsset
		for _, v := range versions {
			logging.LogDebug("Checking version %s against filter %s", v.Version, versionFilter)
			if ExtractMajorVersion(v.Version) == versionFilter {
				logging.LogDebug("  ✓ Version matches filter")
				filteredVersions = append(filteredVersions, v)
			} else {
				logging.LogDebug("  ✗ Version does not match filter")
			}
		}

		// If no version matches the filter, display available versions
		if len(filteredVersions) == 0 {
			logging.LogOutput("❌ No version found matching major version %s", versionFilter)
			logging.LogOutput("")
			logging.LogOutput("💡 Available major versions are: %s", joinInts(availableMajors))
			return nil
		}

		versions = filteredVersions
		logging.LogDebug("Found %d versions after filtering", len(versions))
	}

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		return repository.CompareVersions(versions[i].Version, versions[j].Version)
	})

	output.Versions = versions

	displayVersions(versions, sdkType, distribution)
	return nil
}

// joinInts converts a slice of integers to a string
func joinInts(numbers []int) string {
	var strNumbers []string
	for _, num := range numbers {
		strNumbers = append(strNumbers, strconv.Itoa(num))
	}
	return strings.Join(strNumbers, ", ")
}

func displayVersions(versions []repository.SDKAsset, sdkType, distribution string) {
	logging.LogDebug("Processing %d versions for display", len(versions))

	// Group versions by major version
	versionGroups := make(map[string][]string)
	allMajorVersions := make(map[string]bool)

	// Retrieve all available major versions
	for _, asset := range versions {
		logging.LogDebug("Processing version: %s", asset.Version)
		majorVersion := ExtractMajorVersion(asset.Version)
		logging.LogDebug("  Extracted major version: %s", majorVersion)
		if majorVersion != "" {
			allMajorVersions[majorVersion] = true
			versionGroups[majorVersion] = append(versionGroups[majorVersion], asset.Version)
			logging.LogDebug("  Added to version groups. Current groups: %v", versionGroups)
		}
	}

	// Get sorted major versions
	var majorVersions []int
	for major := range versionGroups {
		if num, err := strconv.Atoi(major); err == nil {
			majorVersions = append(majorVersions, num)
		}
	}
	sort.Ints(majorVersions)

	logging.LogOutput("🔹 Available versions:")
	logging.LogOutput("─────────────────────────")

	// If no version is found
	if len(majorVersions) == 0 {
		logging.LogOutput("❌ No major version found matching your criteria")
		logging.LogOutput("")

		// Create the list of available major versions
		var availableMajors []int
		for major := range allMajorVersions {
			if num, err := strconv.Atoi(major); err == nil {
				availableMajors = append(availableMajors, num)
			}
		}
		sort.Ints(availableMajors)

		// Convert versions to strings for display
		var majorStrings []string
		for _, num := range availableMajors {
			majorStrings = append(majorStrings, strconv.Itoa(num))
		}

		logging.LogOutput("💡 Available major versions are: %s", strings.Join(majorStrings, ", "))
		return
	}

	// Display versions by group
	for _, majorNum := range majorVersions {
		major := strconv.Itoa(majorNum)
		versions := versionGroups[major]

		// Sort versions in each group
		sort.Slice(versions, func(i, j int) bool {
			return repository.CompareVersions(versions[i], versions[j])
		})

		logging.LogOutput("-%d :", majorNum)
		for _, version := range versions {
			logging.LogOutput("    ✅ %s", version)
		}
		logging.LogOutput("") // Empty line between groups
	}

	logging.LogOutput("💡 To install a specific version:")
	logging.LogOutput(fmt.Sprintf("   strigo install %s %s [version]", sdkType, distribution))
}
