package version

import (
	"fmt"
	"os"
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

// NewParser creates a new Parser instance by loading patterns from file
// patternsPath should be the resolved path (already prioritized: CLI > env var > config)
func NewParser(patternsPath string) (*Parser, error) {
	// Validate that a patterns file path was provided
	if patternsPath == "" {
		return nil, fmt.Errorf("patterns file path not configured. Please set 'patterns_file' in strigo.toml, use --patterns flag, or set STRIGO_PATTERNS_PATH environment variable")
	}

	// Read patterns file
	file, err := os.ReadFile(patternsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read patterns file %s: %w\n💡 Hint: Create this file with patterns from the documentation or copy strigopatterns.toml from the Strigo repository", patternsPath, err)
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
// patternsPath should be the resolved path (already prioritized: CLI > env var > config)
func NewParserWithCustomPatterns(patternsPath string, customPatterns []Pattern) (*Parser, error) {
	parser, err := NewParser(patternsPath)
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
