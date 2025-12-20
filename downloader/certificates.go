package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strigo/logging"
)

// CertificateManager handles certificate configuration
type CertificateManager struct{}

// NewCertificateManager creates a new CertificateManager instance
func NewCertificateManager() *CertificateManager {
	return &CertificateManager{}
}

// SetupJDKCertificates configures certificates for a JDK installation
func (cm *CertificateManager) SetupJDKCertificates(installPath string, config CertConfig) error {
	if !config.Enabled {
		logging.LogDebug("Certificate configuration is disabled")
		return nil
	}

	jdkCacertsPath := filepath.Join(installPath, config.JDKSecurityPath)
	if _, err := os.Stat(jdkCacertsPath); err != nil {
		return fmt.Errorf("JDK cacerts not found at %s: %w", jdkCacertsPath, err)
	}

	if _, err := os.Stat(config.SystemCacertsPath); err != nil {
		return fmt.Errorf("system cacerts not found at %s: %w", config.SystemCacertsPath, err)
	}

	// Backup original cacerts
	backupPath := jdkCacertsPath + ".original"
	if err := os.Link(jdkCacertsPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup cacerts: %w", err)
	}

	// Copy system cacerts to JDK
	if err := cm.copyFile(config.SystemCacertsPath, jdkCacertsPath); err != nil {
		return fmt.Errorf("failed to copy system cacerts: %w", err)
	}

	logging.LogDebug("✅ Certificates configured successfully")
	return nil
}

func (cm *CertificateManager) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}
