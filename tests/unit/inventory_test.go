package unit

import (
	"bufio"
	"os"
	"path/filepath"
	"strigo/repository/version"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInventoryJDKsParsing tests that we can parse all real-world JDK versions from inventory-jdks
func TestInventoryJDKsParsing(t *testing.T) {
	parser, err := version.NewParser("")
	require.NoError(t, err)

	// Read inventory file
	inventoryPath := filepath.Join("..", "inventory-jdks.txt")
	file, err := os.Open(inventoryPath)
	require.NoError(t, err, "Failed to open inventory-jdks.txt file")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	successCount := 0
	failedLines := []string{}

	currentSection := ""

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Track sections (comments)
		if strings.HasPrefix(line, "#") {
			currentSection = line
			continue
		}

		// Try to extract version from this line
		_, patternName, err := parser.ExtractVersion(line)

		if err != nil {
			failedLines = append(failedLines, line)
			t.Logf("❌ Line %d [%s]: FAILED to parse: %s", lineNum, currentSection, line)
		} else {
			successCount++
			t.Logf("✅ Line %d [%s]: %s (pattern: %s)", lineNum, currentSection, line, patternName)
		}
	}

	require.NoError(t, scanner.Err())

	// Report results
	totalLines := lineNum
	t.Logf("\n========== INVENTORY PARSING RESULTS ==========")
	t.Logf("Total lines processed: %d", totalLines)
	t.Logf("Successfully parsed: %d", successCount)
	t.Logf("Failed to parse: %d", len(failedLines))

	if len(failedLines) > 0 {
		t.Logf("\nFailed lines:")
		for _, line := range failedLines {
			t.Logf("  - %s", line)
		}
	}

	// Calculate success rate based on actual JDK lines (excluding comments/empty)
	actualJDKLines := successCount + len(failedLines)
	successRate := float64(successCount) / float64(actualJDKLines) * 100
	t.Logf("Success rate: %.1f%% (%d/%d actual JDK lines)", successRate, successCount, actualJDKLines)

	// Assert that we can parse most versions (should be 100%)
	assert.Greater(t, successRate, 95.0, "Should be able to parse at least 95% of inventory JDK versions")
}

// TestSpecificInventoryFormats tests specific format categories from the inventory
func TestSpecificInventoryFormats(t *testing.T) {
	parser, err := version.NewParser("")
	require.NoError(t, err)

	tests := []struct {
		name          string
		path          string
		expectedVer   string
		shouldSucceed bool
	}{
		// Adoptium Temurin - Java 8 legacy format
		{
			name:          "Temurin Java 8 legacy",
			path:          "/jdk/temurin/OpenJDK8U-jdk_x64_linux_hotspot_8u472b08.tar.gz",
			expectedVer:   "8u472b08",
			shouldSucceed: true,
		},
		// Adoptium Temurin - Java 11+
		{
			name:          "Temurin Java 11",
			path:          "/jdk/temurin/OpenJDK11U-jdk_x64_linux_hotspot_11.0.29_7.tar.gz",
			expectedVer:   "11.0.29_7",
			shouldSucceed: true,
		},
		// Adoptium Temurin - Java 23+ (no U suffix)
		{
			name:          "Temurin Java 23 (no U)",
			path:          "/jdk/temurin/OpenJDK23-jdk_x64_linux_hotspot_23_37.tar.gz",
			expectedVer:   "23_37",
			shouldSucceed: true,
		},
		// Amazon Corretto
		{
			name:          "Corretto Java 8",
			path:          "/jdk/corretto/amazon-corretto-8.472.08.1-linux-x64.tar.gz",
			expectedVer:   "8.472.08.1",
			shouldSucceed: true,
		},
		{
			name:          "Corretto Java 11",
			path:          "/jdk/corretto/amazon-corretto-11.0.29.7.1-linux-x64.tar.gz",
			expectedVer:   "11.0.29.7.1",
			shouldSucceed: true,
		},
		// SDKMAN IDs - Zulu
		{
			name:          "SDKMAN Zulu",
			path:          "/jdk/zulu/25.0.1-zulu.tar.gz",
			expectedVer:   "25.0.1",
			shouldSucceed: true,
		},
		// SDKMAN IDs - Liberica
		{
			name:          "SDKMAN Liberica",
			path:          "/jdk/liberica/23.0.2-librca.tar.gz",
			expectedVer:   "23.0.2",
			shouldSucceed: true,
		},
		// SDKMAN IDs - Microsoft
		{
			name:          "SDKMAN Microsoft",
			path:          "/jdk/microsoft/21.0.9-ms.tar.gz",
			expectedVer:   "21.0.9",
			shouldSucceed: true,
		},
		// SDKMAN IDs - Semeru
		{
			name:          "SDKMAN Semeru",
			path:          "/jdk/semeru/17.0.17-sem.tar.gz",
			expectedVer:   "17.0.17",
			shouldSucceed: true,
		},
		// SDKMAN IDs - Oracle
		{
			name:          "SDKMAN Oracle",
			path:          "/jdk/oracle/25.0.1-oracle.tar.gz",
			expectedVer:   "25.0.1",
			shouldSucceed: true,
		},
		// SDKMAN IDs - SapMachine
		{
			name:          "SDKMAN SapMachine",
			path:          "/jdk/sapmachine/21.0.9-sapmchn.tar.gz",
			expectedVer:   "21.0.9",
			shouldSucceed: true,
		},
		// SDKMAN IDs - Dragonwell
		{
			name:          "SDKMAN Dragonwell",
			path:          "/jdk/dragonwell/21.0.9-albba.tar.gz",
			expectedVer:   "21.0.9",
			shouldSucceed: true,
		},
		// SDKMAN IDs - GraalVM CE
		{
			name:          "SDKMAN GraalVM CE",
			path:          "/jdk/graalvm/25.0.1-graalce.tar.gz",
			expectedVer:   "25.0.1",
			shouldSucceed: true,
		},
		// SDKMAN IDs - GraalVM Oracle
		{
			name:          "SDKMAN GraalVM Oracle",
			path:          "/jdk/graalvm/23.0.2-graal.tar.gz",
			expectedVer:   "23.0.2",
			shouldSucceed: true,
		},
		// SDKMAN IDs - Mandrel
		{
			name:          "SDKMAN Mandrel",
			path:          "/jdk/mandrel/25.0.1.r25-mandrel.tar.gz",
			expectedVer:   "25.0.1.r25",
			shouldSucceed: true,
		},
		// Azul Zulu CDN - standard
		{
			name:          "Zulu CDN standard",
			path:          "/jdk/zulu/zulu11.84.17-ca-jdk11.0.29-linux_x64.tar.gz",
			expectedVer:   "11.0.29",
			shouldSucceed: true,
		},
		// Azul Zulu CDN - CRAC variant
		{
			name:          "Zulu CDN CRAC",
			path:          "/jdk/zulu/zulu17.60.17-ca-crac-jdk17.0.16-linux_x64.tar.gz",
			expectedVer:   "17.0.16",
			shouldSucceed: true,
		},
		// Azul Zulu CDN - FX variant
		{
			name:          "Zulu CDN FX",
			path:          "/jdk/zulu/zulu21.44.17-ca-fx-jdk21.0.8-linux_x64.tar.gz",
			expectedVer:   "21.0.8",
			shouldSucceed: true,
		},
		// Azul Zulu CDN - Beta
		{
			name:          "Zulu CDN Beta",
			path:          "/jdk/zulu/zulu23.0.57-beta-jdk23.0.0-beta.23-linux_x64.tar.gz",
			expectedVer:   "23.0.0-beta.23",
			shouldSucceed: true,
		},
		// Mandrel binary format
		{
			name:          "Mandrel binary",
			path:          "/jdk/mandrel/mandrel-java21-linux-amd64-23.1.9.0-Final.tar.gz",
			expectedVer:   "23.1.9.0",
			shouldSucceed: true,
		},
		// GraalVM Community new format
		{
			name:          "GraalVM Community",
			path:          "/jdk/graalvm/graalvm-community-jdk-25.0.1_linux-x64_bin.tar.gz",
			expectedVer:   "25.0.1",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, patternName, err := parser.ExtractVersion(tt.path)

			if tt.shouldSucceed {
				assert.NoError(t, err, "Should extract version successfully")
				assert.Equal(t, tt.expectedVer, ver, "Extracted version should match")
				t.Logf("✅ %s: %s (pattern: %s)", tt.name, ver, patternName)
			} else {
				assert.Error(t, err, "Should fail to extract version")
			}
		})
	}
}
