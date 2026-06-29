package extension_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/extension"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const xdebugIni = `zend_extension=xdebug.so
xdebug.mode=develop
memory_limit=256M
`

const noXdebugIni = `memory_limit=256M
extension=openssl
`

func makeTempIniX(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "php.ini")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	return p
}

func TestXdebug_IsInstalled_True(t *testing.T) {
	p := makeTempIniX(t, xdebugIni)
	x, err := extension.NewXdebugManager(p)
	require.NoError(t, err)
	assert.True(t, x.IsInstalled())
}

func TestXdebug_IsInstalled_False(t *testing.T) {
	p := makeTempIniX(t, noXdebugIni)
	x, err := extension.NewXdebugManager(p)
	require.NoError(t, err)
	assert.False(t, x.IsInstalled())
}

func TestXdebug_CurrentMode(t *testing.T) {
	p := makeTempIniX(t, xdebugIni)
	x, err := extension.NewXdebugManager(p)
	require.NoError(t, err)
	assert.Equal(t, extension.ModeDevelop, x.CurrentMode())
}

func TestXdebug_CurrentMode_Default(t *testing.T) {
	p := makeTempIniX(t, noXdebugIni)
	x, err := extension.NewXdebugManager(p)
	require.NoError(t, err)
	assert.Equal(t, extension.ModeOff, x.CurrentMode())
}

func TestXdebug_SwitchMode(t *testing.T) {
	p := makeTempIniX(t, xdebugIni)
	x, err := extension.NewXdebugManager(p)
	require.NoError(t, err)

	require.NoError(t, x.SwitchMode(extension.ModeCoverage))

	// Reload and verify.
	x2, err := extension.NewXdebugManager(p)
	require.NoError(t, err)
	assert.Equal(t, extension.ModeCoverage, x2.CurrentMode())
}

func TestXdebug_SwitchMode_NotInstalled(t *testing.T) {
	p := makeTempIniX(t, noXdebugIni)
	x, err := extension.NewXdebugManager(p)
	require.NoError(t, err)

	err = x.SwitchMode(extension.ModeDevelop)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestXdebug_Disable(t *testing.T) {
	p := makeTempIniX(t, xdebugIni)
	x, err := extension.NewXdebugManager(p)
	require.NoError(t, err)

	require.NoError(t, x.Disable())

	x2, err := extension.NewXdebugManager(p)
	require.NoError(t, err)
	assert.Equal(t, extension.ModeOff, x2.CurrentMode())
}

func TestXdebug_DetectedVersion(t *testing.T) {
	p := makeTempIniX(t, xdebugIni)
	x, err := extension.NewXdebugManager(p)
	require.NoError(t, err)

	x.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("3.3.2"), nil
	})

	ver, err := x.DetectedVersion("/usr/bin/php")
	require.NoError(t, err)
	assert.Equal(t, "3.3.2", ver)
}

func TestXdebug_ValidModes(t *testing.T) {
	modes := extension.ValidModes()
	assert.Contains(t, modes, extension.ModeOff)
	assert.Contains(t, modes, extension.ModeDevelop)
	assert.Contains(t, modes, extension.ModeCoverage)
}

func TestXdebug_IsValidMode(t *testing.T) {
	assert.True(t, extension.IsValidMode("develop"))
	assert.True(t, extension.IsValidMode("off"))
	assert.False(t, extension.IsValidMode("unknown"))
}
