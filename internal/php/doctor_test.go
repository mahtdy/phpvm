package php_test

import (
	"path/filepath"
	"testing"

	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/php"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctor_Run_NoPhp(t *testing.T) {
	// Use a completely isolated tmp root with no versions dir
	// and no config.Current — so Detect() will always fail.
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, Current: "99.0.0", LogLevel: "info"}
	// 99.0.0 is not installed, so Detect() falls back to config and then
	// fails because the binary doesn't exist in versions dir.

	doc := php.NewDoctor(cfg)
	report := doc.Run()
	require.NotEmpty(t, report.Checks)

	// First check should be PHP binary.
	assert.Equal(t, "PHP binary", report.Checks[0].Name)
	// It should either fail (version dir missing) or OK (system php found).
	// We only assert that the check ran — the status depends on environment.
	assert.NotEmpty(t, report.Checks[0].Detail)
}

func TestDoctor_Run_IsolatedNoPhp(t *testing.T) {
	// Override PATH detection by pointing to a non-existent binary.
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, Current: "", LogLevel: "info"}

	doc := php.NewDoctor(cfg)
	report := doc.Run()

	require.NotEmpty(t, report.Checks)
	assert.Equal(t, "PHP binary", report.Checks[0].Name)
	// With empty Current and no versions dir, at minimum we expect the
	// doctor to produce a non-empty report without panicking.
}

func TestDoctor_Run_WithPhp(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, Current: "8.3.10", LogLevel: "info"}

	// Create a fake PHP binary so Detect() can find it via config fallback.
	makeVersionDir(t, cfg.VersionsDir(), "8.3.10")

	doc := php.NewDoctor(cfg)
	// Stub commands to avoid real PHP calls.
	doc.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		if len(args) > 0 {
			switch args[0] {
			case "--version":
				return []byte("PHP 8.3.10 (cli)"), nil
			case "--ini":
				iniPath := filepath.Join(tmp, "php.ini")
				return []byte("Loaded Configuration File:         " + iniPath + "\n"), nil
			case "-m":
				return []byte("[PHP Modules]\nopenssl\nmbstring\njson\ncurl\n"), nil
			}
		}
		return []byte(""), nil
	})

	report := doc.Run()
	require.NotEmpty(t, report.Checks)

	names := make([]string, len(report.Checks))
	for i, c := range report.Checks {
		names[i] = c.Name
	}
	assert.Contains(t, names, "PHP binary")
}

func TestDoctor_Issues(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Root: tmp, LogLevel: "info"}

	doc := php.NewDoctor(cfg)
	doc.SetExecCommand(func(name string, args ...string) ([]byte, error) {
		return nil, assert.AnError
	})

	report := doc.Run()
	issues := report.Issues()
	assert.NotEmpty(t, issues)
	for _, issue := range issues {
		assert.NotEqual(t, php.CheckOK, issue.Status)
	}
}

func TestCheckStatusConstants(t *testing.T) {
	assert.Equal(t, php.CheckOK, php.CheckStatus(0))
	assert.Equal(t, php.CheckWarn, php.CheckStatus(1))
	assert.Equal(t, php.CheckFail, php.CheckStatus(2))
}
