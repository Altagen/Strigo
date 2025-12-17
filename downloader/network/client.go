package network

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strigo/logging"
	"time"
)

// Client gère les opérations réseau
type Client struct {
	httpClient *http.Client
}

// NewClient crée une nouvelle instance de Client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetFileSize récupère la taille d'un fichier distant
func (c *Client) GetFileSize(url string) (int64, error) {
	resp, err := c.httpClient.Head(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid Content-Length: %w", err)
	}

	return size, nil
}

// DownloadFile télécharge un fichier depuis une URL
func (c *Client) DownloadFile(url, filepath string) error {
	logging.LogDebug("📡 Initiating network request to %s", url)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	logging.LogDebug("✅ Download completed. Wrote %d bytes", written)
	return nil
}
