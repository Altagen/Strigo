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
	// Create temporary directory
	tmpDir := t.TempDir()

	// Test 1: Empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	err := os.Mkdir(emptyDir, 0755)
	require.NoError(t, err)

	f, err := os.Open(emptyDir)
	require.NoError(t, err)
	defer f.Close()

	_, err = f.Readdirnames(1)
	assert.Error(t, err)
	// Main test: verify that it's io.EOF
	assert.ErrorIs(t, err, io.EOF)

	// Test 2: Non-empty directory
	nonEmptyDir := filepath.Join(tmpDir, "nonempty")
	err = os.Mkdir(nonEmptyDir, 0755)
	require.NoError(t, err)

	// Create a file inside
	testFile := filepath.Join(nonEmptyDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	f2, err := os.Open(nonEmptyDir)
	require.NoError(t, err)
	defer f2.Close()

	_, err = f2.Readdirnames(1)
	assert.NoError(t, err) // No error if not empty
}
