package php_test

import (
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/php"
	"github.com/mahtdy/phpvm/internal/path"
	"github.com/mahtdy/phpvm/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwitch_Success(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{
		Root:        tmp,
		Current:     "",
		DefaultArch: "x64",
		LogLevel:    "info",
	}

	versionsDir := cfg.VersionsDir()
	makeVersionDir(t, versionsDir, "8.3.10")
	makeVersionDir(t, versionsDir, "8.4.1")

	switcher := php.NewSwitcherWithPathMgr(cfg, &mockPathMgr{})

	iv, err := switcher.Switch("8.3")
	require.NoError(t, err)
	assert.Equal(t, "8.3.10", iv.Version.String())
	assert.Equal(t, "8.3.10", cfg.Current)
}

func TestSwitch_NotInstalled(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, LogLevel: "info"}

	makeVersionDir(t, cfg.VersionsDir(), "8.3.10")

	switcher := php.NewSwitcherWithPathMgr(cfg, &mockPathMgr{})

	_, err := switcher.Switch("8.5")
	assert.Error(t, err)
}

func TestSwitch_InvalidVersion(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, LogLevel: "info"}

	switcher := php.NewSwitcherWithPathMgr(cfg, &mockPathMgr{})

	_, err := switcher.Switch("not-a-version")
	assert.Error(t, err)
}

func TestInstallDir(t *testing.T) {
	v, err := validation.Parse("8.3.10")
	require.NoError(t, err)
	dir := php.InstallDir("/versions", v)
	assert.Equal(t, filepath.Join("/versions", "8.3.10"), dir)
}

// mockPathMgr is a no-op path.Manager for tests.
type mockPathMgr struct {
	added   string
	removed bool
}

func (m *mockPathMgr) Add(dir string) error        { m.added = dir; return nil }
func (m *mockPathMgr) Remove() error               { m.removed = true; return nil }
func (m *mockPathMgr) Current() ([]string, error)  { return nil, nil }
func (m *mockPathMgr) Verify(string) (bool, error) { return true, nil }

// Ensure mockPathMgr satisfies the interface at compile time.
var _ path.Manager = (*mockPathMgr)(nil)
