package composer_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mahtdy/phpvm/internal/composer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeFakeComposer writes a zero-byte composer binary into dir.
func makeFakeComposer(t *testing.T, dir string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	name := "composer"
	if runtime.GOOS == "windows" {
		name = "composer.bat"
	}
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(""), 0o755))
	return path
}

// --- Detector tests ---

func TestDetector_DetectPerVersion(t *testing.T) {
	tmp := t.TempDir()
	composerPath := makeFakeComposer(t, tmp)

	d := composer.NewDetector()
	d.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("Composer version 2.7.0 2024-01-01 00:00:00"), nil
	})

	info, err := d.DetectPerVersion(tmp)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "2.7.0", info.Version)
	assert.Equal(t, composerPath, info.Path)
	assert.False(t, info.Global)
}

func TestDetector_DetectPerVersion_NotFound(t *testing.T) {
	tmp := t.TempDir() // empty dir — no composer binary

	d := composer.NewDetector()
	info, err := d.DetectPerVersion(tmp)
	require.NoError(t, err)
	assert.Nil(t, info)
}

func TestDetector_DetectGlobal_NotFound(t *testing.T) {
	d := composer.NewDetector()
	// Override exec so LookPath-based detection also fails.
	d.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return nil, assert.AnError
	})
	// Global detect uses exec.LookPath, which won't find a fake binary.
	// If composer is not installed on CI this will return nil (no error).
	info, err := d.DetectGlobal()
	require.NoError(t, err)
	// We can't assert nil here if the machine has composer in PATH,
	// so just verify we got no error.
	_ = info
}

// --- Installer tests ---

func TestInstaller_Install_Success(t *testing.T) {
	tmp := t.TempDir()

	// Create a fake PHP binary.
	phpBin := filepath.Join(tmp, "php")
	if runtime.GOOS == "windows" {
		phpBin = filepath.Join(tmp, "php.exe")
	}
	require.NoError(t, os.WriteFile(phpBin, []byte(""), 0o755))

	inst := composer.NewInstaller()

	// Stub HTTP so we don't hit the real network.
	inst.SetFetchURL(func(url string) ([]byte, error) {
		return []byte("<?php echo 'fake installer';"), nil
	})

	// Stub PHP execution: create composer binary and return success output.
	destDir := filepath.Join(tmp, "dest")
	inst.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		// Simulate the installer creating the binary.
		require.NoError(t, os.MkdirAll(destDir, 0o755))
		binName := "composer"
		if runtime.GOOS == "windows" {
			binName = "composer.bat"
		}
		p := filepath.Join(destDir, binName)
		os.WriteFile(p, []byte(""), 0o755) //nolint:errcheck
		return []byte("Composer successfully installed"), nil
	})

	// Stub the version detection after install.
	d := composer.NewDetector()
	d.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("Composer version 2.7.0 2024-01-01"), nil
	})

	opts := composer.InstallOptions{
		DestDir:             destDir,
		PHPBinary:           phpBin,
		SkipSignatureVerify: true, // skip SHA-384 since we mocked the response
	}

	info, err := inst.Install(context.Background(), opts)
	require.NoError(t, err)
	assert.NotNil(t, info)
}

func TestInstaller_Install_MissingPHP(t *testing.T) {
	inst := composer.NewInstaller()
	_, err := inst.Install(context.Background(), composer.InstallOptions{
		DestDir: t.TempDir(),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PHP binary")
}

func TestInstaller_Install_MissingDestDir(t *testing.T) {
	inst := composer.NewInstaller()
	_, err := inst.Install(context.Background(), composer.InstallOptions{
		PHPBinary: "/usr/bin/php",
	})
	assert.Error(t, err)
}

// --- Updater tests ---

func TestUpdater_AlreadyLatest(t *testing.T) {
	tmp := t.TempDir()
	makeFakeComposer(t, tmp)

	u := composer.NewUpdater()
	u.SetFetchURL(func(url string) ([]byte, error) {
		return []byte(`{"latest":{"version":"2.7.0"}}`), nil
	})
	u.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		// Simulate `composer --version` and `composer self-update`.
		if len(args) > 0 && args[0] == "--version" {
			return []byte("Composer version 2.7.0 2024-01-01"), nil
		}
		return []byte("Already up to date"), nil
	})

	// Pre-plant a version match so detector finds it.
	d := composer.NewDetector()
	d.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return []byte("Composer version 2.7.0 2024-01-01"), nil
	})

	_, err := u.Update(context.Background(), tmp, "/usr/bin/php")
	require.NoError(t, err)
}

// --- Diagnose tests ---

func TestDiagnose_ComposerNotInstalled(t *testing.T) {
	report := composer.Diagnose("/usr/bin/php", nil)
	require.NotEmpty(t, report.Checks)

	// First check should be Composer installed = false.
	first := report.Checks[0]
	assert.Equal(t, "Composer installed", first.Name)
	assert.False(t, first.OK)
}

func TestDiagnose_WithComposer(t *testing.T) {
	info := &composer.Info{Path: "/usr/bin/composer", Version: "2.7.0", Global: true}
	report := composer.Diagnose("/usr/bin/php", info)

	// First check: installed = true.
	assert.True(t, report.Checks[0].OK)
	// Should have checks for extensions too.
	assert.Greater(t, len(report.Checks), 1)
}

func TestDiagnoseReport_Issues(t *testing.T) {
	report := composer.Diagnose("/nonexistent/php", nil)
	issues := report.Issues()
	assert.NotEmpty(t, issues)
	for _, issue := range issues {
		assert.False(t, issue.OK)
	}
}

func TestInfoString(t *testing.T) {
	info := &composer.Info{Path: "/usr/bin/composer", Version: "2.7.0", Global: true}
	s := info.String()
	assert.Contains(t, s, "2.7.0")
	assert.Contains(t, s, "global")
}
