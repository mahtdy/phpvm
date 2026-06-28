package php_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/php"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemove_Success(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, Current: "8.4.1", LogLevel: "info"}

	makeVersionDir(t, cfg.VersionsDir(), "8.3.10")
	makeVersionDir(t, cfg.VersionsDir(), "8.4.1")

	remover := php.NewRemover(cfg)
	iv, err := remover.Remove("8.3")
	require.NoError(t, err)
	assert.Equal(t, "8.3.10", iv.Version.String())

	// Directory should be gone.
	_, statErr := os.Stat(filepath.Join(cfg.VersionsDir(), "8.3.10"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestRemove_ActiveVersionBlocked(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, Current: "8.3.10", LogLevel: "info"}

	makeVersionDir(t, cfg.VersionsDir(), "8.3.10")

	remover := php.NewRemover(cfg)
	_, err := remover.Remove("8.3.10")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "active version")

	// Directory must still exist.
	assert.DirExists(t, filepath.Join(cfg.VersionsDir(), "8.3.10"))
}

func TestRemove_NotInstalled(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, LogLevel: "info"}

	remover := php.NewRemover(cfg)
	_, err := remover.Remove("8.5")
	assert.Error(t, err)
}

func TestRemove_InvalidVersion(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, LogLevel: "info"}

	remover := php.NewRemover(cfg)
	_, err := remover.Remove("bad-version")
	assert.Error(t, err)
}
