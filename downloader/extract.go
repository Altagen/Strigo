package downloader

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"strigo/logging"

	"github.com/ulikunitz/xz"
)

// Extractor handles archive extraction
type Extractor struct{}

// NewExtractor creates a new Extractor instance
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract extracts an archive to a destination directory
func (e *Extractor) Extract(archivePath, destPath string) error {
	if !filepath.IsAbs(destPath) {
		return fmt.Errorf("destination path must be absolute")
	}

	logging.LogDebug(" Starting extraction of %s to %s", filepath.Base(archivePath), destPath)

	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"):
		return e.extractTarGz(archivePath, destPath)
	case strings.HasSuffix(archivePath, ".tar.xz"):
		return e.extractTarXz(archivePath, destPath)
	default:
		return fmt.Errorf("unsupported archive format")
	}
}

func (e *Extractor) extractTarGz(tarPath, destPath string) error {
	logging.LogDebug(" Opening tar.gz archive: %s", filepath.Base(tarPath))
	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	return e.extractTar(tar.NewReader(gzr), destPath)
}

func (e *Extractor) extractTarXz(tarPath, destPath string) error {
	logging.LogDebug(" Opening tar.xz archive: %s", filepath.Base(tarPath))
	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	xzr, err := xz.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	return e.extractTar(tar.NewReader(xzr), destPath)
}

func (e *Extractor) extractTar(tr *tar.Reader, destPath string) error {
	var filesExtracted int
	var totalSize int64

	logging.LogDebug(" Extracting files...")
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		target := filepath.Join(destPath, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid tar path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := e.extractFile(tr, target, header.Mode); err != nil {
				return fmt.Errorf("failed to extract file: %w", err)
			}
			filesExtracted++
			totalSize += header.Size
		}
	}
	logging.LogDebug(" Extraction completed: %d files extracted, total size: %d bytes", filesExtracted, totalSize)
	return nil
}

func (e *Extractor) extractFile(tr io.Reader, path string, mode int64) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(mode))
	if err != nil {
		return err
	}
	defer f.Close()

	written, err := io.Copy(f, tr)
	if err != nil {
		return err
	}
	logging.LogDebug(" Extracted: %s (%d bytes)", filepath.Base(path), written)
	return nil
}
