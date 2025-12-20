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

// Client handles network operations
type Client struct {
	httpClient *http.Client
	username   string
	password   string
}

// NewClient creates a new Client instance without authentication
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithAuth creates a new Client instance with HTTP Basic Authentication
func NewClientWithAuth(username, password string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		username: username,
		password: password,
	}
}

// GetFileSize retrieves the size of a remote file
func (c *Client) GetFileSize(url string) (int64, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Basic Auth if credentials are provided
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
		logging.LogDebug("🔐 Using Basic Auth for file size check")
	}

	resp, err := c.httpClient.Do(req)
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

// DownloadFile downloads a file from a URL
func (c *Client) DownloadFile(url, filepath string) error {
	logging.LogDebug("📡 Initiating network request to %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add Basic Auth if credentials are provided
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
		logging.LogDebug("🔐 Using Basic Auth for download")
	}

	resp, err := c.httpClient.Do(req)
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
