package unit

import (
	"strigo/repository/version"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtractMajor tests the major version extraction function
func TestExtractMajor(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		// Standard semantic versioning
		{
			name:     "Standard Java 11",
			version:  "11.0.26_4",
			expected: "11",
		},
		{
			name:     "Standard Java 17",
			version:  "17.0.11_10",
			expected: "17",
		},
		{
			name:     "Standard Java 21",
			version:  "21.0.6_7",
			expected: "21",
		},

		// Legacy Java 8 format
		{
			name:     "Java 8 legacy",
			version:  "8u442b06",
			expected: "8",
		},
		{
			name:     "Java 8 legacy variant",
			version:  "8u432b06",
			expected: "8",
		},

		// Node.js versions
		{
			name:     "Node.js",
			version:  "22.13.1",
			expected: "22",
		},
		{
			name:     "Node.js older",
			version:  "20.18.1",
			expected: "20",
		},

		// Versions with prefixes
		{
			name:     "JDK prefix",
			version:  "jdk-17.0.11",
			expected: "17",
		},
		{
			name:     "JDK prefix with underscore",
			version:  "jdk_11.0.26_4",
			expected: "11",
		},

		// Edge cases
		{
			name:     "Empty string",
			version:  "",
			expected: "",
		},
		{
			name:     "Single digit",
			version:  "21",
			expected: "",
		},
		{
			name:     "No version info",
			version:  "invalid",
			expected: "",
		},

		// Amazon Corretto format
		{
			name:     "Corretto",
			version:  "11.0.29.7.1",
			expected: "11",
		},
		{
			name:     "Corretto Java 8",
			version:  "8.472.08.1",
			expected: "8",
		},

		// GraalVM format
		{
			name:     "GraalVM",
			version:  "25.0.1",
			expected: "25",
		},
		{
			name:     "GraalVM with build",
			version:  "17.0.11+7.1",
			expected: "17",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := version.ExtractMajor(tt.version)
			assert.Equal(t, tt.expected, result, "ExtractMajor(%q) should return %q", tt.version, tt.expected)
		})
	}
}

// TestCompareVersions tests version comparison logic
func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool // true if v1 < v2
	}{
		// Standard comparisons
		{
			name:     "11.0.26 < 11.0.27",
			v1:       "11.0.26_4",
			v2:       "11.0.27_5",
			expected: true,
		},
		{
			name:     "11.0.27 > 11.0.26",
			v1:       "11.0.27_5",
			v2:       "11.0.26_4",
			expected: false,
		},
		{
			name:     "Same version",
			v1:       "11.0.26_4",
			v2:       "11.0.26_4",
			expected: false,
		},

		// Different major versions
		{
			name:     "Java 11 < Java 17",
			v1:       "11.0.26_4",
			v2:       "17.0.11_10",
			expected: true,
		},
		{
			name:     "Java 17 > Java 11",
			v1:       "17.0.11_10",
			v2:       "11.0.26_4",
			expected: false,
		},

		// Legacy Java 8 format
		{
			name:     "8u432 < 8u442",
			v1:       "8u432b06",
			v2:       "8u442b06",
			expected: true,
		},
		{
			name:     "8u442 > 8u432",
			v1:       "8u442b06",
			v2:       "8u432b06",
			expected: false,
		},

		// Different lengths
		{
			name:     "Shorter version is older",
			v1:       "21.0.6",
			v2:       "21.0.6_7",
			expected: true,
		},
		{
			name:     "Longer version is newer",
			v1:       "21.0.6_7",
			v2:       "21.0.6",
			expected: false,
		},

		// Build numbers
		{
			name:     "Same version different build",
			v1:       "11.0.26_4",
			v2:       "11.0.26_5",
			expected: true,
		},

		// Node.js versions
		{
			name:     "Node 20 < Node 22",
			v1:       "20.18.1",
			v2:       "22.13.1",
			expected: true,
		},

		// Corretto format
		{
			name:     "Corretto versions",
			v1:       "11.0.28.6.1",
			v2:       "11.0.29.7.1",
			expected: true,
		},

		// Mixed formats
		{
			name:     "Java 8 legacy < Java 11",
			v1:       "8u442b06",
			v2:       "11.0.26_4",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := version.CompareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result, "CompareVersions(%q, %q) should return %v", tt.v1, tt.v2, tt.expected)
		})
	}
}

// TestExtractMajorConsistency verifies that ExtractMajor produces consistent results
func TestExtractMajorConsistency(t *testing.T) {
	// Test that same version extracts to same major
	testCases := []string{
		"11.0.26_4",
		"11.0.27_5",
		"11.0.28_6",
	}

	for _, ver := range testCases {
		major := version.ExtractMajor(ver)
		assert.Equal(t, "11", major, "All Java 11 versions should extract to '11'")
	}
}

// TestCompareVersionsSymmetry verifies that comparison is asymmetric
func TestCompareVersionsSymmetry(t *testing.T) {
	pairs := []struct {
		v1 string
		v2 string
	}{
		{"11.0.26_4", "11.0.27_5"},
		{"8u432b06", "8u442b06"},
		{"21.0.6", "21.0.6_7"},
	}

	for _, pair := range pairs {
		forward := version.CompareVersions(pair.v1, pair.v2)
		backward := version.CompareVersions(pair.v2, pair.v1)

		// If v1 < v2, then v2 should NOT be < v1
		if forward {
			assert.False(t, backward, "CompareVersions should be asymmetric: if %q < %q, then %q should not be < %q",
				pair.v1, pair.v2, pair.v2, pair.v1)
		}
	}
}
