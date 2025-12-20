package unit

import (
	"fmt"
	"strigo/repository/version"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCaseInsensitivePatterns tests that patterns work with different case variations
func TestCaseInsensitivePatterns(t *testing.T) {
	parser, err := version.NewParser("strigo-patterns.toml")
	require.NoError(t, err)

	tests := []struct {
		name        string
		paths       []string // Different case variations
		expectedVer string
	}{
		{
			name: "Temurin - OpenJDK variations",
			paths: []string{
				"/repo/OpenJDK11U-jdk_x64_linux_hotspot_11.0.29_7.tar.gz",  // Original
				"/repo/openjdk11u-jdk_x64_linux_hotspot_11.0.29_7.tar.gz",  // Lowercase
				"/repo/OPENJDK11U-JDK_X64_LINUX_HOTSPOT_11.0.29_7.tar.gz",  // Uppercase
				"/repo/OpenJdk11U-jdk_X64_Linux_Hotspot_11.0.29_7.tar.gz",  // Mixed
			},
			expectedVer: "11.0.29_7",
		},
		{
			name: "Corretto variations",
			paths: []string{
				"/repo/amazon-corretto-11.0.29.7.1-linux-x64.tar.gz",  // Original
				"/repo/Amazon-Corretto-11.0.29.7.1-linux-x64.tar.gz",  // Capitalized
				"/repo/AMAZON-CORRETTO-11.0.29.7.1-linux-x64.tar.gz",  // Uppercase
				"/repo/amazon-CORRETTO-11.0.29.7.1-linux-x64.tar.gz",  // Mixed
			},
			expectedVer: "11.0.29.7.1",
		},
		{
			name: "Zulu variations",
			paths: []string{
				"/repo/zulu11.84.17-ca-jdk11.0.29-linux_x64.tar.gz",  // Original
				"/repo/Zulu11.84.17-ca-jdk11.0.29-linux_x64.tar.gz",  // Capitalized
				"/repo/ZULU11.84.17-CA-JDK11.0.29-linux_x64.tar.gz",  // Uppercase
				"/repo/ZuLu11.84.17-Ca-JdK11.0.29-linux_x64.tar.gz",  // Mixed
			},
			expectedVer: "11.0.29",
		},
		{
			name: "GraalVM variations",
			paths: []string{
				"/repo/graalvm-community-jdk-25.0.1_linux-x64.tar.gz",  // Original
				"/repo/GraalVM-community-jdk-25.0.1_linux-x64.tar.gz",  // Capitalized
				"/repo/GRAALVM-COMMUNITY-JDK-25.0.1_linux-x64.tar.gz",  // Uppercase
				"/repo/GraalVm-Community-Jdk-25.0.1_linux-x64.tar.gz",  // Mixed
			},
			expectedVer: "25.0.1",
		},
		{
			name: "Node.js variations",
			paths: []string{
				"/repo/node-v22.13.1-linux-x64.tar.gz",  // Original
				"/repo/Node-v22.13.1-linux-x64.tar.gz",  // Capitalized
				"/repo/NODE-V22.13.1-linux-x64.tar.gz",  // Uppercase
				"/repo/NoDe-V22.13.1-linux-x64.tar.gz",  // Mixed
			},
			expectedVer: "22.13.1",
		},
		{
			name: "Python variations",
			paths: []string{
				"/repo/Python-3.12.1.tar.gz",  // Original
				"/repo/python-3.12.1.tar.gz",  // Lowercase
				"/repo/PYTHON-3.12.1.tar.gz",  // Uppercase
				"/repo/PyThOn-3.12.1.tar.gz",  // Mixed
			},
			expectedVer: "3.12.1",
		},
		{
			name: "Microsoft OpenJDK variations",
			paths: []string{
				"/repo/microsoft-jdk-21.0.9.tar.gz",  // Original
				"/repo/Microsoft-jdk-21.0.9.tar.gz",  // Capitalized
				"/repo/MICROSOFT-JDK-21.0.9.tar.gz",  // Uppercase
				"/repo/MicroSoft-JDK-21.0.9.tar.gz",  // Mixed
			},
			expectedVer: "21.0.9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, path := range tt.paths {
				caseName := []string{"Original", "Capitalized/Lowercase", "Uppercase", "Mixed"}[i]

				ver, patternName, err := parser.ExtractVersion(path)

				assert.NoError(t, err, "Should extract version from %s case", caseName)
				assert.Equal(t, tt.expectedVer, ver, "Version should match for %s case", caseName)
				assert.NotEmpty(t, patternName, "Pattern name should be set for %s case", caseName)

				t.Logf("✅ %s (%s): %s → %s (pattern: %s)", tt.name, caseName, path, ver, patternName)
			}
		})
	}
}

// TestCaseInsensitiveRealWorldScenarios tests real-world repository scenarios
func TestCaseInsensitiveRealWorldScenarios(t *testing.T) {
	parser, err := version.NewParser("strigo-patterns.toml")
	require.NoError(t, err)

	scenarios := []struct {
		scenario string
		path     string
		expected string
	}{
		{
			scenario: "Legacy repository with UPPERCASE naming",
			path:     "/legacy/OPENJDK17U-JDK_X64_LINUX_HOTSPOT_17.0.15_6.TAR.GZ",
			expected: "17.0.15_6",
		},
		{
			scenario: "Build system with lowercase naming",
			path:     "/builds/openjdk21u-jdk_x64_linux_hotspot_21.0.9_10.tar.gz",
			expected: "21.0.9_10",
		},
		{
			scenario: "Mixed case from manual upload",
			path:     "/uploads/Amazon-CORRETTO-23.0.2.7.1-Linux-X64.tar.gz",
			expected: "23.0.2.7.1",
		},
		{
			scenario: "Mirror repository with inconsistent casing",
			path:     "/mirror/ZULU17.62.17-CA-JDK17.0.17-linux_x64.tar.gz",
			expected: "17.0.17",
		},
		{
			scenario: "Enterprise repository with Title Case",
			path:     "/enterprise/Microsoft-Jdk-25.0.1-Linux-X64.tar.gz",
			expected: "25.0.1",
		},
	}

	for _, s := range scenarios {
		t.Run(s.scenario, func(t *testing.T) {
			ver, patternName, err := parser.ExtractVersion(s.path)

			assert.NoError(t, err, "Should handle %s", s.scenario)
			assert.Equal(t, s.expected, ver, "Version should match")

			t.Logf("✅ %s: %s → %s (pattern: %s)", s.scenario, s.path, ver, patternName)
		})
	}
}

// TestCaseInsensitiveWithInventory validates that case-insensitive works with inventory
func TestCaseInsensitiveWithInventory(t *testing.T) {
	parser, err := version.NewParser("strigo-patterns.toml")
	require.NoError(t, err)

	// Simulate variations of inventory entries
	variations := []struct {
		original string
		variant  string
		expected string
	}{
		{
			original: "OpenJDK11U-jdk_x64_linux_hotspot_11.0.29_7",
			variant:  "openjdk11u-jdk_x64_linux_hotspot_11.0.29_7",
			expected: "11.0.29_7",
		},
		{
			original: "amazon-corretto-11.0.29.7.1",
			variant:  "AMAZON-CORRETTO-11.0.29.7.1",
			expected: "11.0.29.7.1",
		},
		{
			original: "zulu11.84.17-ca-jdk11.0.29-linux_x64.tar.gz",
			variant:  "ZULU11.84.17-CA-JDK11.0.29-linux_x64.tar.gz",
			expected: "11.0.29",
		},
	}

	for i, v := range variations {
		t.Run(fmt.Sprintf("Variation_%d", i+1), func(t *testing.T) {
			// Test original
			verOrig, _, errOrig := parser.ExtractVersion(v.original)
			assert.NoError(t, errOrig)
			assert.Equal(t, v.expected, verOrig)

			// Test variant
			verVar, _, errVar := parser.ExtractVersion(v.variant)
			assert.NoError(t, errVar)
			assert.Equal(t, v.expected, verVar)

			// Both should give same result
			assert.Equal(t, verOrig, verVar, "Original and variant should extract same version")

			t.Logf("✅ Original & Variant both extract: %s", v.expected)
		})
	}
}
