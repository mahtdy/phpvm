package php

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/validation"
)

// phpVersionOutputRe matches "PHP 8.3.10 ..." from `php --version` output.
var phpVersionOutputRe = regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)

// CurrentDetector detects the active PHP version.
type CurrentDetector struct {
	versionsDir string
	configCurrent string // version string stored in config
	// execCommand is overridable for testing.
	execCommand func(name string, args ...string) ([]byte, error)
}

// NewCurrentDetector creates a CurrentDetector.
// configCurrent is the version string from Config.Current (may be empty).
func NewCurrentDetector(versionsDir, configCurrent string) *CurrentDetector {
	return &CurrentDetector{
		versionsDir:   versionsDir,
		configCurrent: configCurrent,
		execCommand:   runCommand,
	}
}

// Detect returns the active PHP version.
// Resolution order:
//  1. Run `php --version` on the system PHP binary.
//  2. Fall back to the version stored in config.
func (d *CurrentDetector) Detect() (*validation.Version, string, error) {
	// Try to find PHP in PATH first.
	phpPath, err := exec.LookPath(phpBinaryName())
	if err == nil {
		ver, err := d.detectFromBinary(phpPath)
		if err == nil {
			return ver, phpPath, nil
		}
	}

	// Fall back to config current value.
	if d.configCurrent != "" {
		v, err := validation.Parse(d.configCurrent)
		if err == nil {
			// Verify the version directory actually exists.
			dir := filepath.Join(d.versionsDir, v.String())
			binPath := filepath.Join(dir, phpBinaryName())
			if _, statErr := os.Stat(binPath); statErr == nil {
				return v, binPath, nil
			}
		}
	}

	return nil, "", cli.NewFileNotFoundError(
		"no active PHP version found — run: phpvm use <version>", nil,
	)
}

// SetExecCommand replaces the command executor. Used in tests.
func (d *CurrentDetector) SetExecCommand(fn func(name string, args ...string) ([]byte, error)) {
	d.execCommand = fn
}

// detectFromBinary runs `php --version` and parses the output.
func (d *CurrentDetector) detectFromBinary(phpPath string) (*validation.Version, error) {
	out, err := d.execCommand(phpPath, "--version")
	if err != nil {
		return nil, cli.NewError(
			fmt.Sprintf("failed to run %s --version", phpPath), err,
		)
	}

	m := phpVersionOutputRe.FindSubmatch(out)
	if m == nil {
		return nil, cli.NewError(
			"could not parse PHP version from: "+strings.TrimSpace(string(out)), nil,
		)
	}

	return validation.Parse(string(m[1]))
}

// runCommand is the default command executor.
func runCommand(name string, args ...string) ([]byte, error) {
	//nolint:gosec // name and args come from trusted internal paths
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// PhpBinaryPath returns the path to the PHP binary for the given version directory.
func PhpBinaryPath(versionDir string) string {
	return filepath.Join(versionDir, phpBinaryName())
}

// VersionDirName returns the canonical directory name for a version.
// Uses the full dotted string: "8.3.10".
func VersionDirName(v *validation.Version) string {
	if runtime.GOOS == "windows" && v.IsTS() {
		return v.String() + "-ts"
	}
	return v.String()
}
