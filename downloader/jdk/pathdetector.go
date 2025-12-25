package jdk

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/logging"
)

// CacertsPathDetector handles automatic detection of JDK cacerts location
type CacertsPathDetector struct{}

// NewCacertsPathDetector creates a new path detector instance
func NewCacertsPathDetector() *CacertsPathDetector {
	return &CacertsPathDetector{}
}

// DetectCacertsPath automatically detects the cacerts file location in a JDK installation
// It tries common paths in this order:
//   1. CLI override (if provided)
//   2. jre/lib/security/cacerts (Java 8 and earlier)
//   3. lib/security/cacerts (Java 9+)
//
// Returns the absolute path to cacerts or an error if not found
func (d *CacertsPathDetector) DetectCacertsPath(jdkRootPath string, cliOverride string) (string, error) {
	logging.LogDebug("🔍 Detecting cacerts path in JDK at %s", jdkRootPath)

	// 1. Check CLI override first
	if cliOverride != "" {
		overridePath := filepath.Join(jdkRootPath, cliOverride)
		if info, err := os.Stat(overridePath); err == nil && !info.IsDir() {
			logging.LogDebug("✅ Using CLI override path: %s", overridePath)
			return overridePath, nil
		}
		logging.LogDebug("⚠️  CLI override path not found or is a directory: %s", overridePath)
	}

	// 2. Try Java 8 path first (jre/lib/security/cacerts)
	java8Path := filepath.Join(jdkRootPath, "jre", "lib", "security", "cacerts")
	if info, err := os.Stat(java8Path); err == nil && !info.IsDir() {
		logging.LogDebug("✅ Detected Java 8 cacerts at: %s", java8Path)
		return java8Path, nil
	}

	// 3. Try Java 11+ path (lib/security/cacerts)
	java11Path := filepath.Join(jdkRootPath, "lib", "security", "cacerts")
	if info, err := os.Stat(java11Path); err == nil && !info.IsDir() {
		logging.LogDebug("✅ Detected Java 11+ cacerts at: %s", java11Path)
		return java11Path, nil
	}

	return "", fmt.Errorf("cacerts file not found in JDK at %s. Tried paths:\n  - %s\n  - %s\n💡 Tip: Use --jdk-cacerts-path to specify a custom path",
		jdkRootPath, java8Path, java11Path)
}

// DetectKeystoreFormat attempts to determine if a keystore is JKS or PKCS12
// Returns "JKS" or "PKCS12" or an error
func (d *CacertsPathDetector) DetectKeystoreFormat(cacertsPath string) (string, error) {
	file, err := os.Open(cacertsPath)
	if err != nil {
		return "", fmt.Errorf("failed to open cacerts: %w", err)
	}
	defer file.Close()

	// Read first 4 bytes to detect format
	magic := make([]byte, 4)
	_, err = file.Read(magic)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore magic bytes: %w", err)
	}

	// JKS magic number: 0xFEEDFEED
	if magic[0] == 0xFE && magic[1] == 0xED && magic[2] == 0xFE && magic[3] == 0xED {
		logging.LogDebug("🔍 Detected JKS keystore format")
		return "JKS", nil
	}

	// PKCS12 magic: 0x30 (ASN.1 SEQUENCE tag)
	if magic[0] == 0x30 {
		logging.LogDebug("🔍 Detected PKCS12 keystore format")
		return "PKCS12", nil
	}

	return "", fmt.Errorf("unknown keystore format (magic bytes: %02x %02x %02x %02x)", magic[0], magic[1], magic[2], magic[3])
}
