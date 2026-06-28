package downloader_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/downloader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sha256Hex computes the hex SHA-256 of data.
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func TestDownload_Success(t *testing.T) {
	content := []byte("fake php archive content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content) //nolint:errcheck
	}))
	defer server.Close()

	tmp := t.TempDir()
	opts := downloader.Options{
		URL:        server.URL + "/php-8.3.10.zip",
		DestDir:    tmp,
		NoProgress: true,
	}

	result, err := downloader.Download(context.Background(), opts)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), result.Size)
	assert.FileExists(t, result.Path)

	got, _ := os.ReadFile(result.Path)
	assert.Equal(t, content, got)
}

func TestDownload_ChecksumVerify(t *testing.T) {
	content := []byte("archive data")
	expected := sha256Hex(content)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content) //nolint:errcheck
	}))
	defer server.Close()

	tmp := t.TempDir()
	opts := downloader.Options{
		URL:            server.URL + "/file.zip",
		DestDir:        tmp,
		ExpectedSHA256: expected,
		NoProgress:     true,
	}

	_, err := downloader.Download(context.Background(), opts)
	require.NoError(t, err)
}

func TestDownload_ChecksumMismatch(t *testing.T) {
	content := []byte("archive data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content) //nolint:errcheck
	}))
	defer server.Close()

	tmp := t.TempDir()
	opts := downloader.Options{
		URL:            server.URL + "/file.zip",
		DestDir:        tmp,
		ExpectedSHA256: "0000000000000000000000000000000000000000000000000000000000000000",
		NoProgress:     true,
	}

	_, err := downloader.Download(context.Background(), opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")

	// Partial file should be removed on mismatch.
	files, _ := filepath.Glob(filepath.Join(tmp, "*"))
	assert.Empty(t, files)
}

func TestDownload_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmp := t.TempDir()
	opts := downloader.Options{
		URL:        server.URL + "/missing.zip",
		DestDir:    tmp,
		NoProgress: true,
	}

	_, err := downloader.Download(context.Background(), opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestDownload_EmptyURL(t *testing.T) {
	_, err := downloader.Download(context.Background(), downloader.Options{
		DestDir:    t.TempDir(),
		NoProgress: true,
	})
	assert.Error(t, err)
}

func TestDownload_Resume(t *testing.T) {
	content := []byte("0123456789abcdef")
	partialContent := content[8:] // server sends second half

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			w.Header().Set("Content-Range", "bytes 8-15/16")
			w.WriteHeader(http.StatusPartialContent)
			w.Write(partialContent) //nolint:errcheck
		} else {
			w.Write(content) //nolint:errcheck
		}
	}))
	defer server.Close()

	tmp := t.TempDir()
	destPath := filepath.Join(tmp, "file.zip")
	// Write first half to simulate partial download.
	require.NoError(t, os.WriteFile(destPath, content[:8], 0o640))

	opts := downloader.Options{
		URL:        server.URL + "/file.zip",
		DestDir:    tmp,
		Filename:   "file.zip",
		NoProgress: true,
	}

	result, err := downloader.Download(context.Background(), opts)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), result.Size)
}
