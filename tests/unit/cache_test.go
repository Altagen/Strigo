package unit

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDirEmpty(t *testing.T) {
	// Créer un répertoire temporaire
	tmpDir := t.TempDir()

	// Test 1: Répertoire vide
	emptyDir := filepath.Join(tmpDir, "empty")
	err := os.Mkdir(emptyDir, 0755)
	require.NoError(t, err)

	f, err := os.Open(emptyDir)
	require.NoError(t, err)
	defer f.Close()

	_, err = f.Readdirnames(1)
	assert.Error(t, err)
	// Le test principal: vérifier que c'est bien io.EOF
	assert.ErrorIs(t, err, io.EOF)

	// Test 2: Répertoire non vide
	nonEmptyDir := filepath.Join(tmpDir, "nonempty")
	err = os.Mkdir(nonEmptyDir, 0755)
	require.NoError(t, err)

	// Créer un fichier dedans
	testFile := filepath.Join(nonEmptyDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	f2, err := os.Open(nonEmptyDir)
	require.NoError(t, err)
	defer f2.Close()

	_, err = f2.Readdirnames(1)
	assert.NoError(t, err) // Pas d'erreur si non vide
}
