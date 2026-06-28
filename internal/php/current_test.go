package php_test

import (
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/php"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetect_FromInstalledDir(t *testing.T) {
	tmp := t.TempDir()
	makeVersionDir(t, tmp, "8.3.10")

	detector := php.NewCurrentDetector(tmp, "8.3.10")
	// Override execCommand so it mimics `php --version` output.
	detector.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("PHP 8.3.10 (cli) (built: Jan  1 2025)\nCopyright..."), nil
	})

	v, path, err := detector.Detect()
	require.NoError(t, err)
	assert.Equal(t, "8.3.10", v.String())
	assert.NotEmpty(t, path)
}

func TestDetect_FallsBackToConfig(t *testing.T) {
	tmp := t.TempDir()
	makeVersionDir(t, tmp, "8.4.1")

	detector := php.NewCurrentDetector(tmp, "8.4.1")
	// Simulate PHP not in PATH.
	detector.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return nil, assert.AnError
	})

	v, _, err := detector.Detect()
	require.NoError(t, err)
	assert.Equal(t, "8.4.1", v.String())
}

func TestDetect_NoVersionError(t *testing.T) {
	tmp := t.TempDir()

	detector := php.NewCurrentDetector(tmp, "")
	detector.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return nil, assert.AnError
	})

	_, _, err := detector.Detect()
	assert.Error(t, err)
}

func TestPhpBinaryPath(t *testing.T) {
	path := php.PhpBinaryPath(filepath.Join("home", ".phpvm", "versions", "8.3.10"))
	assert.Contains(t, path, "8.3.10")
	assert.Contains(t, path, "php")
}
