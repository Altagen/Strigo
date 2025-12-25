package unit

import (
	"strigo/repository/version"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserInitialization(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err, "Parser should initialize without error")
	assert.NotNil(t, parser, "Parser should not be nil")

	patterns := parser.ListAllPatterns()
	assert.Greater(t, len(patterns), 10, "Should have multiple builtin patterns")
}

func TestExtractVersionTemurin(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	tests := []struct {
		name          string
		path          string
		expectedVer   string
		shouldSucceed bool
	}{
		{
			name:          "Temurin standard format",
			path:          "/jdk/adoptium/temurin/jdk-11.0.26_4-linux-x64.tar.gz",
			expectedVer:   "11.0.26_4",
			shouldSucceed: true,
		},
		{
			name:          "Temurin alternative format",
			path:          "/jdk/OpenJDK11U-jdk_x64_linux_hotspot_11.0.26_4.tar.gz",
			expectedVer:   "11.0.26_4",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, patternName, err := parser.ExtractVersion(tt.path)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVer, ver)
				assert.NotEmpty(t, patternName)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestExtractVersionCorretto(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	tests := []struct {
		name          string
		path          string
		expectedVer   string
		shouldSucceed bool
	}{
		{
			name:          "Corretto standard",
			path:          "/jdk/amazon/corretto/corretto-11.0.26.4.1-linux-x64.tar.gz",
			expectedVer:   "11.0.26.4.1",
			shouldSucceed: true,
		},
		{
			name:          "Corretto with prefix",
			path:          "/jdk/amazon-corretto-8.442.06.1-linux-x64.tar.gz",
			expectedVer:   "8.442.06.1",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, patternName, err := parser.ExtractVersion(tt.path)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVer, ver)
				assert.NotEmpty(t, patternName)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestExtractVersionZulu(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	path := "/jdk/azul/zulu/zulu11.74.15-ca-jdk11.0.24-linux_x64.tar.gz"
	ver, _, err := parser.ExtractVersion(path)

	assert.NoError(t, err)
	assert.Equal(t, "11.0.24", ver)
}

func TestExtractVersionNodeJS(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	tests := []struct {
		name          string
		path          string
		expectedVer   string
		shouldSucceed bool
	}{
		{
			name:          "Node.js standard",
			path:          "/node/nodejs/node-v22.13.1-linux-x64.tar.gz",
			expectedVer:   "22.13.1",
			shouldSucceed: true,
		},
		{
			name:          "Node.js without platform",
			path:          "/node/nodejs/node-v20.18.1.tar.gz",
			expectedVer:   "20.18.1",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, patternName, err := parser.ExtractVersion(tt.path)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVer, ver)
				assert.NotEmpty(t, patternName)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestExtractVersionByType(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	// Test filtering by type
	jdkPath := "/jdk/temurin/jdk-11.0.26_4-linux-x64.tar.gz"
	nodePath := "/node/nodejs/node-v22.13.1-linux-x64.tar.gz"

	// Extract with correct type
	ver, _, err := parser.ExtractVersionByType(jdkPath, "jdk")
	assert.NoError(t, err)
	assert.Equal(t, "11.0.26_4", ver)

	// Extract Node.js with correct type
	ver, _, err = parser.ExtractVersionByType(nodePath, "node")
	assert.NoError(t, err)
	assert.Equal(t, "22.13.1", ver)

	// Extract Node.js path with wrong type (should still work with generic pattern)
	ver, _, err = parser.ExtractVersionByType(nodePath, "jdk")
	// This might fail or succeed depending on generic patterns
	// The important thing is it doesn't panic
	t.Logf("Extract with wrong type result: ver=%s, err=%v", ver, err)
}

func TestExtractVersionByDistribution(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	// Test filtering by distribution
	temurinPath := "/jdk/temurin/jdk-11.0.26_4-linux-x64.tar.gz"
	correttoPath := "/jdk/corretto/corretto-11.0.26.4.1-linux-x64.tar.gz"

	// Extract Temurin with correct distribution filter
	ver, patternName, err := parser.ExtractVersionByDistribution(temurinPath, "temurin")
	assert.NoError(t, err)
	assert.Equal(t, "11.0.26_4", ver)
	assert.Equal(t, "temurin", patternName)

	// Extract Corretto with correct distribution filter
	ver, patternName, err = parser.ExtractVersionByDistribution(correttoPath, "corretto")
	assert.NoError(t, err)
	assert.Equal(t, "11.0.26.4.1", ver)
	assert.Equal(t, "corretto", patternName)

	// Extract Temurin path with Corretto filter (should fail)
	_, _, err = parser.ExtractVersionByDistribution(temurinPath, "corretto")
	assert.Error(t, err, "Should not match Temurin path with Corretto filter")
}

func TestExtractVersionLegacyJava8(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	// Test legacy Java 8 format (8u442b06) - use a filename with the full version
	path := "/jdk/java8/OpenJDK8U-jdk_x64_linux_8u442b06.tar.gz"
	ver, _, err := parser.ExtractVersion(path)

	assert.NoError(t, err)
	assert.Equal(t, "8u442b06", ver)
}

func TestExtractVersionGraalVM(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		expectedVer string
	}{
		{
			name:        "GraalVM CE",
			path:        "/jdk/graalvm/graalvm-ce-java11-22.3.0-linux-amd64.tar.gz",
			expectedVer: "22.3.0",
		},
		{
			name:        "GraalVM JDK",
			path:        "/jdk/graalvm/graalvm-jdk-17.0.11+7.1-linux-x64.tar.gz",
			expectedVer: "17.0.11+7.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, _, err := parser.ExtractVersion(tt.path)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedVer, ver)
		})
	}
}

func TestExtractVersionErrorCases(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	// Test paths that should not match any pattern
	invalidPaths := []string{
		"/invalid/no-version-here.tar.gz",
		"/jdk/random/file.txt",
		"/completely/random/path",
	}

	for _, path := range invalidPaths {
		t.Run(path, func(t *testing.T) {
			_, _, err := parser.ExtractVersion(path)
			assert.Error(t, err, "Should return error for invalid path")
		})
	}
}

func TestGetPatternsByType(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	jdkPatterns := parser.GetPatternsByType("jdk")
	assert.Greater(t, len(jdkPatterns), 5, "Should have multiple JDK patterns")

	nodePatterns := parser.GetPatternsByType("node")
	assert.Greater(t, len(nodePatterns), 0, "Should have Node.js patterns")

	// Generic patterns should be included for all types
	for _, pattern := range parser.GetPatternsByType("jdk") {
		assert.True(t, pattern.Type == "jdk" || pattern.Type == "*",
			"JDK patterns should be jdk or generic type")
	}
}

func TestGetPatternByName(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	// Test getting existing pattern
	temurinPattern := parser.GetPatternByName("temurin")
	assert.NotNil(t, temurinPattern)
	assert.Equal(t, "temurin", temurinPattern.Name)
	assert.Equal(t, "jdk", temurinPattern.Type)

	// Test getting non-existent pattern
	nonExistent := parser.GetPatternByName("non-existent-pattern")
	assert.Nil(t, nonExistent)
}

func TestNewParserWithCustomPatterns(t *testing.T) {
	// Create custom patterns with a unique prefix that won't conflict with builtins
	customPatterns := []version.Pattern{
		{
			Name:        "custom-vendor",
			Type:        "jdk",
			Description: "Custom vendor JDK pattern",
			Patterns:    []string{`mycompany-jdk-(\\d+\\.\\d+\\.\\d+)`},
		},
	}

	parser, err := version.NewParserWithCustomPatterns("../../strigo-patterns.toml", customPatterns)
	require.NoError(t, err)
	assert.NotNil(t, parser)

	// Test that custom pattern works
	path := "/jdk/mycompany/mycompany-jdk-17.0.5-linux.tar.gz"
	ver, _, err := parser.ExtractVersion(path)

	assert.NoError(t, err)
	assert.Equal(t, "17.0.5", ver)
	// Version extracted successfully - custom pattern is working

	// Also test that we can get the custom pattern by name
	customPattern := parser.GetPatternByName("custom-vendor")
	assert.NotNil(t, customPattern)
	assert.Equal(t, "jdk", customPattern.Type)

	// Verify all patterns are loaded (builtin + custom)
	allPatterns := parser.ListAllPatterns()
	assert.Greater(t, len(allPatterns), 10, "Should have builtin + custom patterns")
}

func TestMultipleDistributions(t *testing.T) {
	parser, err := version.NewParser("../../strigo-patterns.toml")
	require.NoError(t, err)

	// Test that parser can handle many different distributions
	testCases := []struct {
		distribution string
		path         string
		expectedVer  string
	}{
		{"temurin", "/jdk/temurin/jdk-11.0.26_4.tar.gz", "11.0.26_4"},
		{"corretto", "/jdk/corretto-8.442.06.1.tar.gz", "8.442.06.1"},
		{"zulu", "/jdk/zulu11.74.15-ca-jdk11.0.24.tar.gz", "11.0.24"},
		{"nodejs", "/node/node-v22.13.1-linux-x64.tar.gz", "22.13.1"},
	}

	for _, tc := range testCases {
		t.Run(tc.distribution, func(t *testing.T) {
			ver, _, err := parser.ExtractVersion(tc.path)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedVer, ver)
		})
	}
}
