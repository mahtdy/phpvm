package ini_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/ini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleIni = `; php.ini sample
[PHP]
memory_limit = 128M
max_execution_time = 30
upload_max_filesize = 2M

; Extensions
extension=openssl
`

func writeTempIni(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "php.ini")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	return p
}

func TestLoad_ParsesKeyValues(t *testing.T) {
	p := writeTempIni(t, sampleIni)
	f, err := ini.Load(p)
	require.NoError(t, err)

	val, ok := f.Get("memory_limit")
	assert.True(t, ok)
	assert.Equal(t, "128M", val)

	val, ok = f.Get("max_execution_time")
	assert.True(t, ok)
	assert.Equal(t, "30", val)
}

func TestLoad_CaseInsensitiveGet(t *testing.T) {
	p := writeTempIni(t, "Memory_Limit = 256M\n")
	f, err := ini.Load(p)
	require.NoError(t, err)

	_, ok := f.Get("memory_limit")
	assert.True(t, ok)
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := ini.Load("/nonexistent/php.ini")
	assert.Error(t, err)
}

func TestSet_UpdatesExisting(t *testing.T) {
	p := writeTempIni(t, sampleIni)
	f, err := ini.Load(p)
	require.NoError(t, err)

	f.Set("memory_limit", "512M")

	val, _ := f.Get("memory_limit")
	assert.Equal(t, "512M", val)
}

func TestSet_AddsNew(t *testing.T) {
	p := writeTempIni(t, sampleIni)
	f, err := ini.Load(p)
	require.NoError(t, err)

	f.Set("date.timezone", "UTC")
	val, ok := f.Get("date.timezone")
	assert.True(t, ok)
	assert.Equal(t, "UTC", val)
}

func TestDelete_RemovesKey(t *testing.T) {
	p := writeTempIni(t, sampleIni)
	f, err := ini.Load(p)
	require.NoError(t, err)

	f.Delete("memory_limit")
	_, ok := f.Get("memory_limit")
	assert.False(t, ok)
}

func TestSave_RoundTrip(t *testing.T) {
	p := writeTempIni(t, sampleIni)
	f, err := ini.Load(p)
	require.NoError(t, err)

	f.Set("memory_limit", "256M")
	f.Set("new_key", "new_value")
	require.NoError(t, f.Save())

	// Reload and verify.
	f2, err := ini.Load(p)
	require.NoError(t, err)

	v, _ := f2.Get("memory_limit")
	assert.Equal(t, "256M", v)
	v2, _ := f2.Get("new_key")
	assert.Equal(t, "new_value", v2)
}

func TestBackup_CreatesFile(t *testing.T) {
	p := writeTempIni(t, sampleIni)
	f, err := ini.Load(p)
	require.NoError(t, err)

	backupPath, err := f.Backup()
	require.NoError(t, err)
	assert.FileExists(t, backupPath)

	data, _ := os.ReadFile(backupPath)
	assert.Equal(t, sampleIni, string(data))
}

func TestAll_ReturnsAllKeys(t *testing.T) {
	p := writeTempIni(t, sampleIni)
	f, err := ini.Load(p)
	require.NoError(t, err)

	all := f.All()
	assert.Contains(t, all, "memory_limit")
	assert.Contains(t, all, "max_execution_time")
}

func TestLocator_FindForVersion(t *testing.T) {
	tmp := t.TempDir()
	iniPath := filepath.Join(tmp, "php.ini")
	require.NoError(t, os.WriteFile(iniPath, []byte(sampleIni), 0o644))

	l := ini.NewLocator()
	l.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("Loaded Configuration File:         " + iniPath + "\n"), nil
	})

	found, err := l.Find(filepath.Join(tmp, "php"))
	require.NoError(t, err)
	assert.Equal(t, iniPath, found)
}
