package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	// Use a temp dir so no real config file exists.
	tmp := t.TempDir()
	t.Setenv("PHPVM_ROOT", tmp)

	cfg, err := config.Load("")
	require.NoError(t, err)

	assert.Equal(t, tmp, cfg.Root)
	assert.Equal(t, "", cfg.Current)
	assert.False(t, cfg.Telemetry)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoadFromFile(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.json")

	content := `{"current":"8.3.10","default_arch":"arm64","telemetry":true,"log_level":"debug"}`
	err := os.WriteFile(cfgPath, []byte(content), 0o600)
	require.NoError(t, err)

	t.Setenv("PHPVM_ROOT", tmp)
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, "8.3.10", cfg.Current)
	assert.Equal(t, "arm64", cfg.DefaultArch)
	assert.True(t, cfg.Telemetry)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestEnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("PHPVM_ROOT", tmp)
	t.Setenv("PHPVM_LOG_LEVEL", "warn")

	cfg, err := config.Load("")
	require.NoError(t, err)

	assert.Equal(t, "warn", cfg.LogLevel)
}

func TestSaveAndReload(t *testing.T) {
	tmp := t.TempDir()

	original := &config.Config{
		Root:        tmp,
		Current:     "8.4.1",
		DefaultArch: "x64",
		Telemetry:   false,
		LogLevel:    "info",
		NoColor:     false,
	}

	err := config.Save(original)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmp, "config.json")
	assert.FileExists(t, cfgPath)

	t.Setenv("PHPVM_ROOT", tmp)
	loaded, err := config.Load(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, original.Current, loaded.Current)
	assert.Equal(t, original.DefaultArch, loaded.DefaultArch)
	assert.Equal(t, original.Telemetry, loaded.Telemetry)
}

func TestDerivedPaths(t *testing.T) {
	root := filepath.Join("home", "user", ".phpvm")
	cfg := &config.Config{Root: root}

	assert.Equal(t, filepath.Join(root, "versions"), cfg.VersionsDir())
	assert.Equal(t, filepath.Join(root, "downloads"), cfg.DownloadsDir())
	assert.Equal(t, filepath.Join(root, "cache"), cfg.CacheDir())
	assert.Equal(t, filepath.Join(root, "plugins"), cfg.PluginsDir())
}

func TestMissingConfigFileIsNotError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("PHPVM_ROOT", tmp)

	// No config file written — should get defaults without error.
	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}
