package version

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strigo/logging"

	"github.com/pelletier/go-toml"
)

// Pattern represents a single regex pattern for version extraction
type Pattern struct {
	Name        string   `toml:"name"`
	Type        string   `toml:"type"`
	Description string   `toml:"description"`
	Patterns    []string `toml:"patterns"`
}

// PatternConfig holds all pattern configurations
type PatternConfig struct {
	Patterns []Pattern `toml:"patterns"`
}

// Parser handles version extraction from SDK paths
type Parser struct {
	patterns []Pattern
}

// GetPatternsFilePath returns the path to the patterns configuration file
// Priority order:
//   1. STRIGO_PATTERNS_PATH env var (highest priority)
//   2. Config file setting (if provided)
//   3. Default: strigopatterns.toml in current directory
func GetPatternsFilePath(configPath string) string {
	// 1. Check environment variable first (highest priority)
	if envPath := os.Getenv("STRIGO_PATTERNS_PATH"); envPath != "" {
		return envPath
	}

	// 2. Use config file setting if provided
	if configPath != "" {
		return configPath
	}

	// 3. Default fallback
	return "strigopatterns.toml"
}

// EnsurePatternsFile creates the patterns file with default content if it doesn't exist
func EnsurePatternsFile(configPath string) error {
	patternsPath := GetPatternsFilePath(configPath)

	// Check if file already exists
	if _, err := os.Stat(patternsPath); err == nil {
		logging.LogDebug("📦 Patterns file already exists: %s", patternsPath)
		return nil
	}

	logging.LogDebug("📦 Creating default patterns file: %s", patternsPath)

	// Read the builtin patterns file as default
	builtinPath := filepath.Join("repository", "version", "patterns", "builtin.toml")
	defaultContent, err := os.ReadFile(builtinPath)
	if err != nil {
		// If builtin.toml doesn't exist, use embedded default
		logging.LogDebug("⚠️  Could not read builtin.toml, using minimal default")
		defaultContent = []byte(getMinimalDefaultPatterns())
	}

	// Update header comment to reflect that this is the user-editable file
	header := `# Strigo Version Parsing Patterns
#
# This file defines regex patterns for extracting versions from SDK distribution paths.
# You can edit this file to add your own custom patterns.
#
# IMPORTANT: After modifying this file, restart strigo or run the command again.
#
# PATTERN STRUCTURE:
# [[patterns]]
# name = "provider-name"           # Unique identifier for the pattern set
# type = "jdk"                      # SDK type (jdk, node, python, etc.)
# description = "..."               # Human-readable description
# patterns = [                      # Array of regex patterns (tried in order)
#     "(?i)pattern1...",            # Case-insensitive pattern 1
#     "(?i)pattern2...",            # Case-insensitive pattern 2
# ]
#
# TIPS:
# - Use (?i) at the start of patterns for case-insensitive matching
# - Patterns are tried in the order they appear in this file
# - You can add new [[patterns]] sections for custom distributions
#

`

	// Write the file
	if err := os.WriteFile(patternsPath, []byte(header+string(defaultContent)), 0644); err != nil {
		return fmt.Errorf("failed to create patterns file: %w", err)
	}

	logging.LogDebug("✅ Created patterns file: %s", patternsPath)
	return nil
}

// getMinimalDefaultPatterns returns a minimal set of patterns if builtin.toml is not available
func getMinimalDefaultPatterns() string {
	return `# Minimal default patterns (builtin.toml not found)

[[patterns]]
name = "temurin"
type = "jdk"
description = "Eclipse Temurin (AdoptOpenJDK)"
patterns = [
    "(?i)jdk-(\\d+\\.\\d+\\.\\d+_\\d+)",
    "(?i)OpenJDK\\d+U-jdk_x64_linux_hotspot_(\\d+\\.\\d+\\.\\d+_\\d+)",
]

[[patterns]]
name = "corretto"
type = "jdk"
description = "Amazon Corretto"
patterns = [
    "(?i)corretto-(\\d+\\.\\d+\\.\\d+\\.\\d+(?:\\.\\d+)?)",
    "(?i)amazon-corretto-(\\d+\\.\\d+\\.\\d+\\.\\d+(?:\\.\\d+)?)",
]

[[patterns]]
name = "generic-version"
type = "*"
description = "Generic semantic versioning"
patterns = [
    "(\\d+\\.\\d+\\.\\d+)",
]
`
}

// NewParser creates a new Parser instance by loading patterns from file
// configPatternsPath is the path from strigo.toml config (can be empty)
func NewParser(configPatternsPath string) (*Parser, error) {
	// Ensure patterns file exists
	if err := EnsurePatternsFile(configPatternsPath); err != nil {
		logging.LogDebug("⚠️  Failed to ensure patterns file: %v", err)
		// Continue anyway, will try to load
	}

	// Load patterns from file
	patternsPath := GetPatternsFilePath(configPatternsPath)
	file, err := os.ReadFile(patternsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read patterns file %s: %w", patternsPath, err)
	}

	var config PatternConfig
	if err := toml.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse patterns file: %w", err)
	}

	logging.LogDebug("📦 Loaded %d patterns from %s", len(config.Patterns), patternsPath)

	return &Parser{
		patterns: config.Patterns,
	}, nil
}

// NewParserWithCustomPatterns creates a parser with additional custom patterns
// configPatternsPath is the path from strigo.toml config (can be empty)
func NewParserWithCustomPatterns(configPatternsPath string, customPatterns []Pattern) (*Parser, error) {
	parser, err := NewParser(configPatternsPath)
	if err != nil {
		return nil, err
	}

	// Prepend custom patterns (they will be tried before builtin patterns)
	parser.patterns = append(customPatterns, parser.patterns...)

	logging.LogDebug("📦 Added %d custom patterns (total: %d)", len(customPatterns), len(parser.patterns))

	return parser, nil
}

// ExtractVersion extracts a version from a path using all available patterns
// Returns the version string and the pattern name that matched
func (p *Parser) ExtractVersion(path string) (version string, patternName string, err error) {
	logging.LogDebug("🔍 Extracting version from path: %s", path)

	for _, pattern := range p.patterns {
		for _, regexStr := range pattern.Patterns {
			re, err := regexp.Compile(regexStr)
			if err != nil {
				logging.LogDebug("⚠️  Invalid regex pattern %s: %v", regexStr, err)
				continue
			}

			if matches := re.FindStringSubmatch(path); len(matches) > 1 {
				version := matches[1]
				logging.LogDebug("✅ Matched pattern '%s' (%s): extracted version %s", pattern.Name, pattern.Description, version)
				return version, pattern.Name, nil
			}
		}
	}

	return "", "", fmt.Errorf("no pattern matched for path: %s", path)
}

// ExtractVersionByType extracts a version using only patterns for a specific SDK type
func (p *Parser) ExtractVersionByType(path string, sdkType string) (version string, patternName string, err error) {
	logging.LogDebug("🔍 Extracting version from path (type filter: %s): %s", sdkType, path)

	for _, pattern := range p.patterns {
		// Skip patterns that don't match the requested type
		if pattern.Type != sdkType && pattern.Type != "*" {
			continue
		}

		for _, regexStr := range pattern.Patterns {
			re, err := regexp.Compile(regexStr)
			if err != nil {
				logging.LogDebug("⚠️  Invalid regex pattern %s: %v", regexStr, err)
				continue
			}

			if matches := re.FindStringSubmatch(path); len(matches) > 1 {
				version := matches[1]
				logging.LogDebug("✅ Matched pattern '%s' (%s): extracted version %s", pattern.Name, pattern.Description, version)
				return version, pattern.Name, nil
			}
		}
	}

	return "", "", fmt.Errorf("no pattern matched for path: %s (type: %s)", path, sdkType)
}

// ExtractVersionByDistribution extracts a version using only patterns for a specific distribution
func (p *Parser) ExtractVersionByDistribution(path string, distribution string) (version string, patternName string, err error) {
	logging.LogDebug("🔍 Extracting version from path (distribution filter: %s): %s", distribution, path)

	for _, pattern := range p.patterns {
		// Skip patterns that don't match the requested distribution
		if pattern.Name != distribution {
			continue
		}

		for _, regexStr := range pattern.Patterns {
			re, err := regexp.Compile(regexStr)
			if err != nil {
				logging.LogDebug("⚠️  Invalid regex pattern %s: %v", regexStr, err)
				continue
			}

			if matches := re.FindStringSubmatch(path); len(matches) > 1 {
				version := matches[1]
				logging.LogDebug("✅ Matched pattern '%s' (%s): extracted version %s", pattern.Name, pattern.Description, version)
				return version, pattern.Name, nil
			}
		}
	}

	return "", "", fmt.Errorf("no pattern matched for path: %s (distribution: %s)", path, distribution)
}

// GetPatternsByType returns all patterns for a specific SDK type
func (p *Parser) GetPatternsByType(sdkType string) []Pattern {
	var result []Pattern
	for _, pattern := range p.patterns {
		if pattern.Type == sdkType || pattern.Type == "*" {
			result = append(result, pattern)
		}
	}
	return result
}

// GetPatternByName returns a specific pattern by name
func (p *Parser) GetPatternByName(name string) *Pattern {
	for _, pattern := range p.patterns {
		if pattern.Name == name {
			return &pattern
		}
	}
	return nil
}

// ListAllPatterns returns all available patterns
func (p *Parser) ListAllPatterns() []Pattern {
	return p.patterns
}
