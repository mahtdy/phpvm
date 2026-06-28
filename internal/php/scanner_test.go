package php_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mahtdy/phpvm/internal/php"
	"github.com/mahtdy/phpvm/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// phpBin returns the platform-appropriate PHP binary name.
func phpBin() string {
	if runtime.GOOS == "windows" {
		return "php.exe"
	}
	return "php"
}

// makeVersionDir creates a fake installed PHP version directory.
func makeVersionDir(t *testing.T, root, version string) {
	t.Helper()
	dir := filepath.Join(root, version)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	// Create a zero-byte fake PHP binary.
	f, err := os.Create(filepath.Join(dir, phpBin()))
	require.NoError(t, err)
	f.Close()
}

func TestScanInstalled_Empty(t *testing.T) {
	tmp := t.TempDir()
	scanner := php.NewScanner(filepath.Join(tmp, "versions"))

	versions, err := scanner.ScanInstalled()
	require.NoError(t, err)
	assert.Empty(t, versions, "no versions dir should return empty slice")
}

func TestScanInstalled_SortedAscending(t *testing.T) {
	tmp := t.TempDir()
	scanner := php.NewScanner(tmp)

	makeVersionDir(t, tmp, "8.4.1")
	makeVersionDir(t, tmp, "8.2.20")
	makeVersionDir(t, tmp, "8.3.10")

	versions, err := scanner.ScanInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 3)

	assert.Equal(t, "8.2.20", versions[0].Version.String())
	assert.Equal(t, "8.3.10", versions[1].Version.String())
	assert.Equal(t, "8.4.1", versions[2].Version.String())
}

func TestScanInstalled_SkipsNonVersionDirs(t *testing.T) {
	tmp := t.TempDir()
	scanner := php.NewScanner(tmp)

	makeVersionDir(t, tmp, "8.3.10")
	// Non-version directories should be ignored.
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "cache"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "bin"), 0o755))
	// Directory with version name but no PHP binary.
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "8.4.0"), 0o755))

	versions, err := scanner.ScanInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.Equal(t, "8.3.10", versions[0].Version.String())
}

func TestScanInstalled_Dir(t *testing.T) {
	tmp := t.TempDir()
	scanner := php.NewScanner(tmp)

	makeVersionDir(t, tmp, "8.3.10")

	versions, err := scanner.ScanInstalled()
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.Equal(t, filepath.Join(tmp, "8.3.10"), versions[0].Dir)
}

func TestFindBestMatch_ExactVersion(t *testing.T) {
	tmp := t.TempDir()
	scanner := php.NewScanner(tmp)

	makeVersionDir(t, tmp, "8.3.10")
	makeVersionDir(t, tmp, "8.4.1")

	sel, err := validation.Parse("8.3.10")
	require.NoError(t, err)

	iv, err := scanner.FindBestMatch(sel)
	require.NoError(t, err)
	assert.Equal(t, "8.3.10", iv.Version.String())
}

func TestFindBestMatch_MinorSelector(t *testing.T) {
	tmp := t.TempDir()
	scanner := php.NewScanner(tmp)

	makeVersionDir(t, tmp, "8.3.9")
	makeVersionDir(t, tmp, "8.3.10")
	makeVersionDir(t, tmp, "8.4.1")

	sel, err := validation.Parse("8.3")
	require.NoError(t, err)

	iv, err := scanner.FindBestMatch(sel)
	require.NoError(t, err)
	// Should pick the highest patch for 8.3.
	assert.Equal(t, "8.3.10", iv.Version.String())
}

func TestFindBestMatch_NotInstalled(t *testing.T) {
	tmp := t.TempDir()
	scanner := php.NewScanner(tmp)

	makeVersionDir(t, tmp, "8.3.10")

	sel, err := validation.Parse("8.4")
	require.NoError(t, err)

	_, err = scanner.FindBestMatch(sel)
	assert.Error(t, err)
}
