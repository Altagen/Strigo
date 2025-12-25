package jdk

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strigo/config"
	"strigo/logging"
	"time"

	keystore "github.com/pavlo-v-chernykh/keystore-go/v4"
)

// CertificateManager handles certificate injection into JDK keystores
type CertificateManager struct {
	pathDetector *CacertsPathDetector
}

// NewCertificateManager creates a new CertificateManager instance
func NewCertificateManager() *CertificateManager {
	return &CertificateManager{
		pathDetector: NewCacertsPathDetector(),
	}
}

// InjectCertificates adds custom certificates to the JDK keystore
// Parameters:
//   - jdkRootPath: Root directory of the extracted JDK
//   - customCerts: List of certificates with their explicit aliases
//   - pathOverride: Optional CLI override for cacerts path
//   - password: Keystore password (default: "changeit", "" for password-less)
//
// Returns error if injection fails (non-fatal - JDK installation continues)
func (cm *CertificateManager) InjectCertificates(jdkRootPath string, customCerts []config.CertificateEntry, pathOverride string, password string) error {
	// Skip if no custom certificates configured
	if len(customCerts) == 0 {
		logging.LogDebug("📋 No custom certificates configured, skipping certificate injection")
		return nil
	}

	logging.LogInfo("🔐 Starting certificate injection into JDK keystore...")

	// Step 1: Detect cacerts path
	cacertsPath, err := cm.pathDetector.DetectCacertsPath(jdkRootPath, pathOverride)
	if err != nil {
		return fmt.Errorf("failed to detect cacerts path: %w", err)
	}
	logging.LogDebug("📂 Using cacerts at: %s", cacertsPath)

	// Step 2: Detect keystore format (for logging purposes)
	format, err := cm.pathDetector.DetectKeystoreFormat(cacertsPath)
	if err != nil {
		logging.LogDebug("⚠️  Could not detect keystore format: %v", err)
	} else {
		logging.LogDebug("📦 Keystore format: %s", format)
	}

	// Step 3: Create backup of original cacerts
	backupPath := cacertsPath + ".original"
	if err := cm.backupCacerts(cacertsPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup cacerts: %w", err)
	}
	logging.LogDebug("💾 Backed up original cacerts to: %s", backupPath)

	// Step 4: Load existing keystore with password fallback
	ks, actualPassword, err := cm.loadKeystoreWithFallback(cacertsPath, password)
	if err != nil {
		return fmt.Errorf("failed to load keystore: %w", err)
	}
	logging.LogDebug("✅ Loaded existing keystore with %d entries", len(ks.Aliases()))

	// Step 5: Add custom certificates
	totalCertsAdded := 0
	for _, certEntry := range customCerts {
		if err := cm.addCertificateFromFile(ks, certEntry); err != nil {
			logging.LogDebug("⚠️  Failed to add certificate from %s: %v", certEntry.Path, err)
			continue
		}
		totalCertsAdded++
		logging.LogDebug("✅ Added certificate '%s' from %s", certEntry.Alias, certEntry.Path)
	}

	if totalCertsAdded == 0 {
		// Restore backup since no certificates were added
		if err := os.Rename(backupPath, cacertsPath); err != nil {
			logging.LogDebug("⚠️  Failed to restore backup: %v", err)
		}
		return fmt.Errorf("no certificates were successfully added")
	}

	// Step 6: Save updated keystore
	if err := cm.saveKeystore(ks, cacertsPath, actualPassword); err != nil {
		// Restore backup on failure
		if restoreErr := os.Rename(backupPath, cacertsPath); restoreErr != nil {
			logging.LogDebug("⚠️  Failed to restore backup: %v", restoreErr)
		}
		return fmt.Errorf("failed to save keystore: %w", err)
	}

	logging.LogInfo("✅ Successfully injected %d custom certificate(s) into JDK keystore", totalCertsAdded)
	return nil
}

// loadKeystoreWithFallback attempts to load keystore with password, falling back to empty password
func (cm *CertificateManager) loadKeystoreWithFallback(path string, password string) (keystore.KeyStore, []byte, error) {
	passwordBytes := []byte(password)

	// Try with provided password first
	ks, err := cm.loadKeystore(path, passwordBytes)
	if err == nil {
		return ks, passwordBytes, nil
	}

	// If password is not empty and loading failed, try with empty password (PKCS12 password-less)
	if password != "" {
		logging.LogDebug("⚠️  Failed to load with provided password, trying empty password (PKCS12 password-less)...")
		ks, err = cm.loadKeystore(path, []byte(""))
		if err == nil {
			logging.LogDebug("✅ Successfully loaded keystore with empty password")
			return ks, []byte(""), nil
		}
	}

	return keystore.KeyStore{}, nil, fmt.Errorf("failed to load keystore with provided password or empty password: %w", err)
}

// loadKeystore loads a JKS/PKCS12 keystore from disk
func (cm *CertificateManager) loadKeystore(path string, password []byte) (keystore.KeyStore, error) {
	file, err := os.Open(path)
	if err != nil {
		return keystore.KeyStore{}, fmt.Errorf("failed to open keystore: %w", err)
	}
	defer file.Close()

	ks := keystore.New()
	err = ks.Load(file, password)
	if err != nil {
		return keystore.KeyStore{}, fmt.Errorf("failed to decode keystore: %w", err)
	}

	return ks, nil
}

// saveKeystore saves a keystore to disk
func (cm *CertificateManager) saveKeystore(ks keystore.KeyStore, path string, password []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create keystore file: %w", err)
	}
	defer file.Close()

	err = ks.Store(file, password)
	if err != nil {
		return fmt.Errorf("failed to encode keystore: %w", err)
	}

	return nil
}

// backupCacerts creates a backup copy of the cacerts file
func (cm *CertificateManager) backupCacerts(src, dst string) error {
	// Skip if backup already exists
	if _, err := os.Stat(dst); err == nil {
		logging.LogDebug("Backup already exists at %s, skipping", dst)
		return nil
	}

	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// addCertificateFromFile parses PEM file and adds the certificate to keystore
func (cm *CertificateManager) addCertificateFromFile(ks keystore.KeyStore, certEntry config.CertificateEntry) error {
	// Read certificate file
	certData, err := os.ReadFile(certEntry.Path)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse PEM certificate
	cert, err := cm.parsePEMCertificate(certData)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Add certificate to keystore with user-provided alias
	entry := keystore.TrustedCertificateEntry{
		CreationTime: time.Now(),
		Certificate: keystore.Certificate{
			Type:    "X.509",
			Content: cert.Raw,
		},
	}

	if err := ks.SetTrustedCertificateEntry(certEntry.Alias, entry); err != nil {
		return fmt.Errorf("failed to add certificate with alias %s: %w", certEntry.Alias, err)
	}

	return nil
}

// parsePEMCertificate parses a PEM-encoded certificate
func (cm *CertificateManager) parsePEMCertificate(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("PEM block is not a certificate (type: %s)", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse X.509 certificate: %w", err)
	}

	return cert, nil
}
