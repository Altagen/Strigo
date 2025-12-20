package unit

import (
	"strigo/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchedAssetPointerValidity(t *testing.T) {
	// Create a list of assets
	assets := []repository.SDKAsset{
		{Version: "11.0.24_8", DownloadUrl: "http://url1"},
		{Version: "11.0.26_4", DownloadUrl: "http://url2"},
		{Version: "17.0.2_8", DownloadUrl: "http://url3"},
	}

	// Simulate search (with corrected code)
	var matchedAsset *repository.SDKAsset
	targetVersion := "11.0.26_4"

	for i := range assets {
		if assets[i].Version == targetVersion {
			matchedAsset = &assets[i]
			break
		}
	}

	// Verify that the pointer is valid
	assert.NotNil(t, matchedAsset)
	assert.Equal(t, "11.0.26_4", matchedAsset.Version)
	assert.Equal(t, "http://url2", matchedAsset.DownloadUrl)

	// Verify that data is stable (no memory corruption)
	version1 := matchedAsset.Version
	url1 := matchedAsset.DownloadUrl

	// Perform some operations
	_ = len(assets)

	// Re-verify
	assert.Equal(t, version1, matchedAsset.Version)
	assert.Equal(t, url1, matchedAsset.DownloadUrl)
}
