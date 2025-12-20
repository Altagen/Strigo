package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strigo/config"
	"strigo/repository"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [type] [distribution]",
	Short: "List installed SDK versions",
	Long: `List installed SDK versions. For example:
strigo list              # List all installed SDKs
strigo list jdk         # List all installed JDK distributions
strigo list jdk temurin # List installed Temurin JDK versions`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 2 {
			return fmt.Errorf("\n❌ Too many arguments\n\n" +
				"Usage:\n" +
				"  strigo list                    # List all SDK types\n" +
				"  strigo list jdk               # List all JDK distributions\n" +
				"  strigo list jdk temurin      # List Temurin JDK versions\n")
		}
		return nil
	},
	Run: list,
	Example: `  # List all SDK types
  strigo list

  # List all JDK distributions
  strigo list jdk

  # List installed Temurin JDK versions
  strigo list jdk temurin`,
}

func list(cmd *cobra.Command, args []string) {
	if err := handleList(args); err != nil {
		ExitWithError(err)
	}
}

func handleList(args []string) error {
	output := &CommandOutput{}

	switch len(args) {
	case 0:
		return listSDKTypes(cfg, output)
	case 1:
		return listDistributions(cfg, args[0], output)
	case 2:
		return listVersions(cfg, args[0], args[1], output)
	default:
		return fmt.Errorf("too many arguments")
	}
}

func listSDKTypes(cfg *config.Config, output *CommandOutput) error {
	var types []string
	for sdkType := range cfg.SDKTypes {
		types = append(types, sdkType)
	}
	sort.Strings(types)
	output.Types = types

	if jsonOutput {
		return OutputJSON(output)
	}

	if len(types) == 0 {
		fmt.Printf("No SDKs installed\n")
		return nil
	}

	fmt.Printf("Available SDK types:\n")
	fmt.Printf("─────────────────────\n")
	for _, sdkType := range types {
		fmt.Printf("✅ %s\n", sdkType)
	}
	fmt.Printf("\n")

	return nil
}

func listDistributions(cfg *config.Config, sdkType string, output *CommandOutput) error {
	// Check if SDK type exists
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Build base path
	basePath := filepath.Join(cfg.General.SDKInstallDir, sdkTypeConfig.InstallDir)

	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			if jsonOutput {
				return OutputJSON(output)
			}
			fmt.Printf("No distributions installed for %s\n", sdkType)
			return nil
		}
		return fmt.Errorf("failed to read distributions directory: %w", err)
	}

	var dists []string
	for _, entry := range entries {
		if entry.IsDir() {
			dists = append(dists, entry.Name())
		}
	}
	sort.Strings(dists)
	output.Distributions = dists

	if jsonOutput {
		return OutputJSON(output)
	}

	if len(dists) == 0 {
		fmt.Printf("No distributions installed for %s\n", sdkType)
		return nil
	}

	fmt.Printf("Installed %s distributions:\n", sdkType)
	fmt.Printf("─────────────────────────────\n")
	for _, dist := range dists {
		fmt.Printf("✅ %s\n", dist)
	}
	fmt.Printf("\n")

	return nil
}

func listVersions(cfg *config.Config, sdkType, distribution string, output *CommandOutput) error {
	// Check if SDK type exists
	sdkTypeConfig, exists := cfg.SDKTypes[sdkType]
	if !exists {
		return fmt.Errorf("SDK type %s not found in configuration", sdkType)
	}

	// Build base path
	basePath := filepath.Join(cfg.General.SDKInstallDir, sdkTypeConfig.InstallDir, distribution)

	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			if jsonOutput {
				return OutputJSON(output)
			}
			fmt.Printf("No versions installed for %s %s\n", sdkType, distribution)
			return nil
		}
		return fmt.Errorf("failed to read versions directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		return repository.CompareVersions(versions[i], versions[j])
	})
	output.Versions = versions

	if jsonOutput {
		return OutputJSON(output)
	}

	if len(versions) == 0 {
		fmt.Printf("No versions installed for %s %s\n", sdkType, distribution)
		return nil
	}

	// Group versions by major version
	versionGroups := make(map[string][]string)
	for _, version := range versions {
		majorVersion := repository.ExtractMajorVersion(version)
		versionGroups[majorVersion] = append(versionGroups[majorVersion], version)
	}

	// Get sorted major versions
	var majorVersions []int
	for major := range versionGroups {
		if num, err := strconv.Atoi(major); err == nil {
			majorVersions = append(majorVersions, num)
		}
	}
	sort.Ints(majorVersions)

	fmt.Printf("\nInstalled %s %s versions:\n", sdkType, distribution)
	fmt.Println("─────────────────────────────────")

	for _, majorNum := range majorVersions {
		major := strconv.Itoa(majorNum)
		versions := versionGroups[major]

		// Sort versions in each group
		sort.Slice(versions, func(i, j int) bool {
			return repository.CompareVersions(versions[i], versions[j])
		})

		fmt.Printf("-%s :\n", major)
		for _, version := range versions {
			fmt.Printf("    ✅ %s\n", version)
		}
		fmt.Println()
	}

	return nil
}
