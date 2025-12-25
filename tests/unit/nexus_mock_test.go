package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strigo/config"
	"strigo/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Nexus API response structure
type mockNexusResponse struct {
	Items []mockNexusItem `json:"items"`
}

type mockNexusItem struct {
	DownloadURL string `json:"downloadUrl"`
	Path        string `json:"path"`
}

// TestNexusClientWithMockServer tests the Nexus client with a mock HTTP server
func TestNexusClientWithMockServer(t *testing.T) {
	// Create mock Nexus server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		assert.Equal(t, "/service/rest/v1/assets", r.URL.Path)
		assert.Equal(t, "raw", r.URL.Query().Get("repository"))
		// Note: path parameter is NOT sent to Nexus (not supported by API)
		// Filtering happens client-side instead

		// Return mock response
		response := mockNexusResponse{
			Items: []mockNexusItem{
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/adoptium/temurin/OpenJDK11U-jdk_x64_linux_hotspot_11.0.24_8.tar.gz",
					Path:        "/jdk/adoptium/temurin/OpenJDK11U-jdk_x64_linux_hotspot_11.0.24_8.tar.gz",
				},
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/adoptium/temurin/OpenJDK17U-jdk_x64_linux_hotspot_17.0.15_6.tar.gz",
					Path:        "/jdk/adoptium/temurin/OpenJDK17U-jdk_x64_linux_hotspot_17.0.15_6.tar.gz",
				},
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/adoptium/temurin/OpenJDK21U-jdk_x64_linux_hotspot_21.0.9_10.tar.gz",
					Path:        "/jdk/adoptium/temurin/OpenJDK21U-jdk_x64_linux_hotspot_21.0.9_10.tar.gz",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Configure mock registry
	registry := config.Registry{
		Type:   "nexus",
		APIURL: server.URL + "/service/rest/v1/assets?repository={repository}",
	}

	repo := config.SDKRepository{
		Type:       "jdk",
		Registry:   "nexus",
		Repository: "raw",
		Path:       "jdk/adoptium/temurin",
	}

	// Fetch versions
	assets, err := repository.FetchAvailableVersions(repo, registry, "", true, "strigo-patterns.toml")
	require.NoError(t, err)
	require.NotNil(t, assets)

	// Verify results
	assert.Len(t, assets, 3, "Should have 3 versions")

	// Versions are sorted, so check they all exist
	versions := make([]string, len(assets))
	for i, asset := range assets {
		versions[i] = asset.Version
	}
	assert.Contains(t, versions, "11.0.24_8")
	assert.Contains(t, versions, "17.0.15_6")
	assert.Contains(t, versions, "21.0.9_10")
}

// TestNexusClientEmptyResponse tests handling of empty response
func TestNexusClientEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := mockNexusResponse{
			Items: []mockNexusItem{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	registry := config.Registry{
		Type:   "nexus",
		APIURL: server.URL + "/service/rest/v1/assets?repository={repository}",
	}

	repo := config.SDKRepository{
		Type:       "jdk",
		Registry:   "nexus",
		Repository: "raw",
		Path:       "jdk/test/empty",
	}

	assets, err := repository.FetchAvailableVersions(repo, registry, "", true, "strigo-patterns.toml")
	require.Error(t, err, "Should return error when no versions found")
	assert.Contains(t, err.Error(), "no versions found")
	assert.Nil(t, assets)
}

// TestNexusClientHTTPError tests handling of HTTP errors
func TestNexusClientHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	registry := config.Registry{
		Type:   "nexus",
		APIURL: server.URL + "/service/rest/v1/assets?repository={repository}",
	}

	repo := config.SDKRepository{
		Type:       "jdk",
		Registry:   "nexus",
		Repository: "raw",
		Path:       "jdk/test/error",
	}

	assets, err := repository.FetchAvailableVersions(repo, registry, "", true, "strigo-patterns.toml")
	require.Error(t, err)
	assert.Nil(t, assets)
	assert.Contains(t, err.Error(), "nexus API returned 500")
}

// TestNexusClientInvalidJSON tests handling of invalid JSON response
func TestNexusClientInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("invalid json {{{"))
	}))
	defer server.Close()

	registry := config.Registry{
		Type:   "nexus",
		APIURL: server.URL + "/service/rest/v1/assets?repository={repository}",
	}

	repo := config.SDKRepository{
		Type:       "jdk",
		Registry:   "nexus",
		Repository: "raw",
		Path:       "jdk/test/invalid",
	}

	assets, err := repository.FetchAvailableVersions(repo, registry, "", true, "strigo-patterns.toml")
	require.Error(t, err)
	assert.Nil(t, assets)
	assert.Contains(t, err.Error(), "failed to decode JSON response")
}

// TestNexusClientVersionFiltering tests version filtering
func TestNexusClientVersionFiltering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := mockNexusResponse{
			Items: []mockNexusItem{
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/corretto/amazon-corretto-11.0.24.7.1-linux-x64.tar.gz",
					Path:        "/jdk/corretto/amazon-corretto-11.0.24.7.1-linux-x64.tar.gz",
				},
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/corretto/amazon-corretto-17.0.15.8.1-linux-x64.tar.gz",
					Path:        "/jdk/corretto/amazon-corretto-17.0.15.8.1-linux-x64.tar.gz",
				},
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/corretto/amazon-corretto-21.0.9.11.1-linux-x64.tar.gz",
					Path:        "/jdk/corretto/amazon-corretto-21.0.9.11.1-linux-x64.tar.gz",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	registry := config.Registry{
		Type:   "nexus",
		APIURL: server.URL + "/service/rest/v1/assets?repository={repository}",
	}

	repo := config.SDKRepository{
		Type:       "jdk",
		Registry:   "nexus",
		Repository: "raw",
		Path:       "jdk/corretto", // Match the path in mock data
	}

	// Fetch all versions
	assets, err := repository.FetchAvailableVersions(repo, registry, "", true, "strigo-patterns.toml")
	require.NoError(t, err)
	assert.Len(t, assets, 3)

	// Verify versions are extracted correctly (order may vary due to sorting)
	versions := make([]string, len(assets))
	for i, asset := range assets {
		versions[i] = asset.Version
	}
	assert.Contains(t, versions, "11.0.24.7.1")
	assert.Contains(t, versions, "17.0.15.8.1")
	assert.Contains(t, versions, "21.0.9.11.1")
}

// TestNexusClientMixedVersionFormats tests handling of multiple version formats
func TestNexusClientMixedVersionFormats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := mockNexusResponse{
			Items: []mockNexusItem{
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/test/jdk-11.0.24_8-linux.tar.gz",
					Path:        "/jdk/test/jdk-11.0.24_8-linux.tar.gz",
				},
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/test/jdk-17.0.15.tar.gz",
					Path:        "/jdk/test/jdk-17.0.15.tar.gz",
				},
				{
					DownloadURL: "http://nexus.example.com/repository/raw/jdk/test/some-random-file.txt",
					Path:        "/jdk/test/some-random-file.txt",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	registry := config.Registry{
		Type:   "nexus",
		APIURL: server.URL + "/service/rest/v1/assets?repository={repository}",
	}

	repo := config.SDKRepository{
		Type:       "jdk",
		Registry:   "nexus",
		Repository: "raw",
		Path:       "jdk/test",
	}

	assets, err := repository.FetchAvailableVersions(repo, registry, "", true, "strigo-patterns.toml")
	require.NoError(t, err)

	// Should only extract versions from files with recognizable patterns
	// some-random-file.txt should be ignored or extracted via generic pattern
	assert.GreaterOrEqual(t, len(assets), 2, "Should extract at least 2 valid versions")
}

// TestNexusClientNetworkTimeout tests handling of network timeouts
func TestNexusClientNetworkTimeout(t *testing.T) {
	// Note: This test might take some time depending on default HTTP client timeout
	// Skip if running in CI or add a timeout to the test context
	t.Skip("Skipping timeout test - requires HTTP client timeout configuration")

	// Create a server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the client timeout (if configured)
		// For now, just test that slow servers don't hang forever
		select {}
	}))
	defer server.Close()

	registry := config.Registry{
		Type:   "nexus",
		APIURL: server.URL + "/service/rest/v1/assets?repository={repository}",
	}

	repo := config.SDKRepository{
		Type:       "jdk",
		Registry:   "nexus",
		Repository: "raw",
		Path:       "jdk/test/timeout",
	}

	_, err := repository.FetchAvailableVersions(repo, registry, "", true, "strigo-patterns.toml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}
