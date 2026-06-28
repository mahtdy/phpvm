package php_test

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/php"
	ver "github.com/mahtdy/phpvm/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildFakeZip builds a minimal zip containing a php.exe / php stub.
func buildFakeZip(t *testing.T) []byte {
	t.Helper()
	var buf []byte
	tmpFile, err := os.CreateTemp(t.TempDir(), "*.zip")
	require.NoError(t, err)
	defer tmpFile.Close()

	w := zip.NewWriter(tmpFile)
	binName := "php.exe"
	if runtime.GOOS != "windows" {
		binName = "php"
	}
	fw, err := w.Create(binName)
	require.NoError(t, err)
	fw.Write([]byte("#!/bin/sh\necho PHP 8.3.10")) //nolint:errcheck
	w.Close()

	buf, err = os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	return buf
}

func TestInstaller_AlreadyInstalled(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, DefaultArch: "x64", LogLevel: "info"}

	// Pre-create the version dir with a fake binary.
	makeVersionDir(t, cfg.VersionsDir(), "8.3.10")

	inst := php.NewInstaller(cfg)
	inst.SetResolver(fakeResolverWithURL(t, "8.3.10", ""))

	iv, err := inst.Install(context.Background(), php.InstallOptions{
		VersionStr: "8.3.10",
		NoProgress: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "8.3.10", iv.Version.String())
}

func TestInstaller_Force(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, DefaultArch: "x64", LogLevel: "info"}
	makeVersionDir(t, cfg.VersionsDir(), "8.3.10")

	zipData := buildFakeZip(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(zipData) //nolint:errcheck
	}))
	defer server.Close()

	inst := php.NewInstaller(cfg)
	inst.SetResolver(fakeResolverWithURL(t, "8.3.10", server.URL+"/php-8.3.10.zip"))

	iv, err := inst.Install(context.Background(), php.InstallOptions{
		VersionStr: "8.3.10",
		Force:      true,
		NoProgress: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "8.3.10", iv.Version.String())

	binName := "php.exe"
	if runtime.GOOS != "windows" {
		binName = "php"
	}
	assert.FileExists(t, filepath.Join(iv.Dir, binName))
}

func TestInstaller_InvalidVersion(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, LogLevel: "info"}
	inst := php.NewInstaller(cfg)

	_, err := inst.Install(context.Background(), php.InstallOptions{
		VersionStr: "invalid",
		NoProgress: true,
	})
	assert.Error(t, err)
}

// fakeResolverWithURL builds a resolver whose LatestPatch returns a release
// pointing at downloadURL so tests can use an httptest.Server.
func fakeResolverWithURL(t *testing.T, versionStr, downloadURL string) *ver.Resolver {
	t.Helper()
	r := ver.NewResolver(t.TempDir())
	r.SetFetchURL(func(_ string) ([]byte, error) {
		// Build a fake windows-style listing that the resolver can parse.
		// The filename must encode the version so winZipRe matches it.
		filename := "php-" + versionStr + "-nts-Win32-vs16-x64.zip"
		page := `<a href="` + filename + `">` + filename + `</a>`
		return []byte(page), nil
	})
	// Override the download URL so it hits the test server instead of
	// windows.php.net.
	r.SetDownloadBaseURL(downloadURL)
	return r
}
