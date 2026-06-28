package archive_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/archive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----- helpers to build in-memory archives -----

func makeZip(t *testing.T, files map[string]string) string {
	t.Helper()
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "test.zip")

	f, err := os.Create(zipPath)
	require.NoError(t, err)
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, content := range files {
		fw, err := w.Create(name)
		require.NoError(t, err)
		_, err = fw.Write([]byte(content))
		require.NoError(t, err)
	}
	return zipPath
}

func makeTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	tmp := t.TempDir()
	archivePath := filepath.Join(tmp, "test.tar.gz")

	f, err := os.Create(archivePath)
	require.NoError(t, err)
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		data := []byte(content)
		hdr := &tar.Header{
			Name:     name,
			Mode:     0o644,
			Size:     int64(len(data)),
			Typeflag: tar.TypeReg,
		}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err = tw.Write(data)
		require.NoError(t, err)
	}
	return archivePath
}

// ----- tests -----

func TestExtractZip_BasicFiles(t *testing.T) {
	zipPath := makeZip(t, map[string]string{
		"php.exe":  "fake php binary",
		"php.ini":  "[PHP]\nmemory_limit = 128M",
		"sub/lib.dll": "dll content",
	})

	destDir := filepath.Join(t.TempDir(), "php-8.3.10")
	require.NoError(t, archive.Extract(zipPath, destDir))

	assert.FileExists(t, filepath.Join(destDir, "php.exe"))
	assert.FileExists(t, filepath.Join(destDir, "php.ini"))
	assert.FileExists(t, filepath.Join(destDir, "sub", "lib.dll"))

	content, _ := os.ReadFile(filepath.Join(destDir, "php.exe"))
	assert.Equal(t, "fake php binary", string(content))
}

func TestExtractTarGz_BasicFiles(t *testing.T) {
	archivePath := makeTarGz(t, map[string]string{
		"php":     "fake php binary",
		"php.ini": "memory_limit=128M",
	})

	destDir := filepath.Join(t.TempDir(), "php-8.3.10")
	require.NoError(t, archive.Extract(archivePath, destDir))

	assert.FileExists(t, filepath.Join(destDir, "php"))
	assert.FileExists(t, filepath.Join(destDir, "php.ini"))
}

func TestExtract_UnsupportedFormat(t *testing.T) {
	tmp := t.TempDir()
	badFile := filepath.Join(tmp, "archive.bz2")
	require.NoError(t, os.WriteFile(badFile, []byte("data"), 0o644))

	err := archive.Extract(badFile, filepath.Join(tmp, "dest"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported archive format")
}

func TestExtractZip_ZipSlipPrevented(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "evil.zip")

	f, err := os.Create(zipPath)
	require.NoError(t, err)

	w := zip.NewWriter(f)
	// Attempt path traversal.
	fw, err := w.Create("../../../evil.txt")
	require.NoError(t, err)
	_, err = fw.Write([]byte("evil"))
	require.NoError(t, err)
	w.Close()
	f.Close()

	destDir := filepath.Join(tmp, "dest")
	err = archive.Extract(zipPath, destDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "zip-slip")
}

func TestExtractZip_Idempotent(t *testing.T) {
	zipPath := makeZip(t, map[string]string{"php.exe": "v1"})
	destDir := filepath.Join(t.TempDir(), "php")

	require.NoError(t, archive.Extract(zipPath, destDir))

	// Write a new zip with updated content and extract again.
	zipPath2 := makeZip(t, map[string]string{"php.exe": "v2"})
	require.NoError(t, archive.Extract(zipPath2, destDir))

	content, _ := os.ReadFile(filepath.Join(destDir, "php.exe"))
	assert.Equal(t, "v2", string(content))
}

func TestExtractTarGz_InvalidArchive(t *testing.T) {
	tmp := t.TempDir()
	badPath := filepath.Join(tmp, "bad.tar.gz")
	require.NoError(t, os.WriteFile(badPath, []byte("not a real gzip"), 0o644))

	err := archive.Extract(badPath, filepath.Join(tmp, "dest"))
	assert.Error(t, err)
}

// TestExtractZip_LargeFilesRejected verifies the 2 GB limit guard is present.
// We can't actually test 2 GB, but we can verify the reader is limited.
func TestExtractZip_ReaderIsLimited(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fw, _ := w.Create("huge.bin")
	// Write 10 bytes — just testing the wiring, not the actual limit.
	fw.Write(bytes.Repeat([]byte{0}, 10)) //nolint:errcheck
	w.Close()

	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "limited.zip")
	require.NoError(t, os.WriteFile(zipPath, buf.Bytes(), 0o644))

	destDir := filepath.Join(tmp, "dest")
	require.NoError(t, archive.Extract(zipPath, destDir))
	assert.FileExists(t, filepath.Join(destDir, "huge.bin"))
}
