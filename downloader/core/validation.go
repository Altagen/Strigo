package core

import (
	"fmt"
	"os"
)

// Validator handles system validations
type Validator struct{}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateSpace checks available space
func (v *Validator) ValidateSpace(fileSize int64, directory string) error {
	return CheckDiskSpace(fileSize, directory)
}

// ValidateDirectories checks and creates necessary directories
func (v *Validator) ValidateDirectories(installPath string) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}
	return nil
}
