package version

import (
	"regexp"
	"strconv"
	"strings"
)

// ExtractMajor extracts the major version from a full version string.
//
// Examples:
//   - "11.0.26_4" → "11"
//   - "8u442b06" → "8"
//   - "21.0.6_7" → "21"
//   - "22.13.1" → "22"
//   - "jdk-17.0.11" → "17"
//   - "" → ""
//
// This function consolidates version extraction logic that was previously
// duplicated across cmd/available.go and repository/fetcher.go.
func ExtractMajor(version string) string {
	// Handle empty version
	if version == "" {
		return ""
	}

	// Try regex-based patterns first (fast path)
	// Note: These patterns require a separator (. or u) to ensure we're not
	// matching a standalone number which isn't a valid version format
	patterns := []string{
		`^(\d+)\..*`, // For 11.0.26_4, 21.0.6_7 (requires dot)
		`^(\d+)u.*`,  // For 8u442b06 (requires u separator)
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(version); len(matches) > 1 {
			return matches[1]
		}
	}

	// Fallback: Handle versions with non-numeric prefixes (like "jdk-17.0.11")
	// Remove any non-numeric prefix
	cleanVersion := strings.TrimLeft(version, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_")

	// Only proceed if the cleaned version has a separator (. or u)
	// This ensures we don't extract from standalone numbers like "21"
	if !strings.Contains(cleanVersion, ".") && !strings.Contains(cleanVersion, "u") {
		return ""
	}

	// Find first number which represents major version
	parts := strings.Split(cleanVersion, ".")
	if len(parts) > 0 {
		// For versions like "8u442b06", extract the 8
		majorPart := strings.Split(parts[0], "u")[0]
		if _, err := strconv.Atoi(majorPart); err == nil {
			return majorPart
		}
	}

	// No major version found
	return ""
}

// CompareVersions compares two versions and returns true if v1 is older than v2.
//
// This function normalizes version strings by converting separators (u, _) to dots,
// then compares each numeric part sequentially.
//
// Examples:
//   - CompareVersions("11.0.26_4", "11.0.27_5") → true (11.0.26 < 11.0.27)
//   - CompareVersions("8u442b06", "8u432b06") → false (442 > 432)
//   - CompareVersions("21.0.6", "21.0.6_7") → true (6 < 6.7)
func CompareVersions(v1, v2 string) bool {
	// Normalize versions to handle different formats
	// Convert: "8u442b06" → "8.442.06"
	//          "11.0.26_4" → "11.0.26.4"
	v1Parts := strings.Split(strings.Replace(strings.Replace(v1, "u", ".", -1), "_", ".", -1), ".")
	v2Parts := strings.Split(strings.Replace(strings.Replace(v2, "u", ".", -1), "_", ".", -1), ".")

	// Compare each numeric part
	minLen := len(v1Parts)
	if len(v2Parts) < minLen {
		minLen = len(v2Parts)
	}

	for i := 0; i < minLen; i++ {
		n1, err1 := strconv.Atoi(v1Parts[i])
		n2, err2 := strconv.Atoi(v2Parts[i])

		// If either part is not numeric, compare as strings
		if err1 != nil || err2 != nil {
			if v1Parts[i] < v2Parts[i] {
				return true
			}
			if v1Parts[i] > v2Parts[i] {
				return false
			}
			continue
		}

		// Compare numerically
		if n1 < n2 {
			return true
		}
		if n1 > n2 {
			return false
		}
	}

	// If all parts are equal so far, the shorter version is "older"
	// Example: "21.0.6" < "21.0.6_7"
	return len(v1Parts) < len(v2Parts)
}
