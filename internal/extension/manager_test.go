package extension_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/extension"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleIni = `extension=openssl
extension=mbstring
; extension=redis
memory_limit = 128M
`

func makeTempIni(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "php.ini")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	return p
}

func TestManager_List(t *testing.T) {
	p := makeTempIni(t, sampleIni)
	m, err := extension.NewManager(p)
	require.NoError(t, err)

	list := m.List()
	// extension key appears once in the ini (last one wins in our simple parser)
	// so at least one extension should be listed.
	assert.NotEmpty(t, list)
}

func TestManager_ListLoaded(t *testing.T) {
	p := makeTempIni(t, sampleIni)
	m, err := extension.NewManager(p)
	require.NoError(t, err)

	m.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("[PHP Modules]\nopenssl\nmbstring\njson\n\n[Zend Modules]\n"), nil
	})

	loaded, err := m.ListLoaded("/usr/bin/php")
	require.NoError(t, err)

	names := make([]string, len(loaded))
	for i, s := range loaded {
		names[i] = s.Name
	}
	assert.Contains(t, names, "openssl")
	assert.Contains(t, names, "mbstring")
	assert.Contains(t, names, "json")
}

func TestManager_GetStatus_Loaded(t *testing.T) {
	p := makeTempIni(t, sampleIni)
	m, err := extension.NewManager(p)
	require.NoError(t, err)

	m.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("[PHP Modules]\nopenssl\n"), nil
	})

	s, err := m.GetStatus("/usr/bin/php", "openssl")
	require.NoError(t, err)
	assert.True(t, s.Enabled)
	assert.Equal(t, "openssl", s.Name)
}

func TestManager_GetStatus_NotLoaded(t *testing.T) {
	p := makeTempIni(t, sampleIni)
	m, err := extension.NewManager(p)
	require.NoError(t, err)

	m.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("[PHP Modules]\nopenssl\n"), nil
	})

	s, err := m.GetStatus("/usr/bin/php", "redis")
	require.NoError(t, err)
	assert.False(t, s.Enabled)
}

func TestManager_Enable(t *testing.T) {
	p := makeTempIni(t, "memory_limit = 128M\n")
	m, err := extension.NewManager(p)
	require.NoError(t, err)

	require.NoError(t, m.Enable("redis"))

	// Reload and verify.
	m2, err := extension.NewManager(p)
	require.NoError(t, err)
	list := m2.List()
	found := false
	for _, s := range list {
		if s.Path == "redis" {
			found = true
		}
	}
	assert.True(t, found, "redis extension should be in php.ini")
}

func TestManager_Disable(t *testing.T) {
	p := makeTempIni(t, "extension=redis\nmemory_limit = 128M\n")
	m, err := extension.NewManager(p)
	require.NoError(t, err)

	require.NoError(t, m.Disable("redis"))

	// After reload the extension key should be gone.
	m2, err := extension.NewManager(p)
	require.NoError(t, err)
	list := m2.List()
	for _, s := range list {
		assert.NotEqual(t, "redis", s.Path)
	}
}
