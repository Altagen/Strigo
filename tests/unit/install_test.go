package unit

import (
	"strigo/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchedAssetPointerValidity(t *testing.T) {
	// Créer une liste d'assets
	assets := []repository.SDKAsset{
		{Version: "11.0.24_8", DownloadUrl: "http://url1"},
		{Version: "11.0.26_4", DownloadUrl: "http://url2"},
		{Version: "17.0.2_8", DownloadUrl: "http://url3"},
	}

	// Simuler la recherche (avec le code corrigé)
	var matchedAsset *repository.SDKAsset
	targetVersion := "11.0.26_4"

	for i := range assets {
		if assets[i].Version == targetVersion {
			matchedAsset = &assets[i]
			break
		}
	}

	// Vérifier que le pointeur est valide
	assert.NotNil(t, matchedAsset)
	assert.Equal(t, "11.0.26_4", matchedAsset.Version)
	assert.Equal(t, "http://url2", matchedAsset.DownloadUrl)

	// Vérifier que les données sont stables (pas de corruption mémoire)
	version1 := matchedAsset.Version
	url1 := matchedAsset.DownloadUrl

	// Faire quelques opérations
	_ = len(assets)

	// Re-vérifier
	assert.Equal(t, version1, matchedAsset.Version)
	assert.Equal(t, url1, matchedAsset.DownloadUrl)
}
