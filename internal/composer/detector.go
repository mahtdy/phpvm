package composer

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/mahtdy/phpvm/internal/cli"
)

// composerVersionRe extracts the version from `composer --version` output.
// e.g. "Composer version 2.7.0 2024-02-09 17:52:04"
var composerVersionRe = regexp.MustCompile(`Composer version (\S+)`)

// composerBinaryNames lists candidate filenames for the Composer binary.
func composerBinaryNames() []string {
	if runtime.GOOS == "windows" {
		return []string{"composer.bat", "composer.cmd", "composer.phar", "composer"}
	}
	return []string{"composer", "composer.phar"}
}

// Detector locates Composer binaries on the system.
type Detector struct {
	// execCommand is overridable for tests.
	execCommand func(name string, args ...string) ([]byte, error)
}

// NewDetector creates a Detector with default settings.
func NewDetector() *Detector {
	return &Detector{execCommand: runCommand}
}

// DetectGlobal looks for Composer in the system PATH.
// Returns nil if Composer is not found (not an error).
func (d *Detector) DetectGlobal() (*Info, error) {
	for _, name := range composerBinaryNames() {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		ver, err := d.versionFromBinary(path)
		if err != nil {
			continue
		}
		return &Info{Path: path, Version: ver, Global: true}, nil
	}
	return nil, nil
}

// DetectPerVersion looks for Composer inside a specific PHP version directory.
// Returns nil if not found.
func (d *Detector) DetectPerVersion(phpVersionDir string) (*Info, error) {
	for _, name := range composerBinaryNames() {
		path := filepath.Join(phpVersionDir, name)
		ver, err := d.versionFromBinary(path)
		if err != nil {
			continue
		}
		return &Info{Path: path, Version: ver, Global: false}, nil
	}
	return nil, nil
}

// Detect finds the best available Composer.
// It checks the per-version directory first, then falls back to global PATH.
func (d *Detector) Detect(phpVersionDir string) (*Info, error) {
	if phpVersionDir != "" {
		if info, _ := d.DetectPerVersion(phpVersionDir); info != nil {
			return info, nil
		}
	}
	info, err := d.DetectGlobal()
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, cli.NewFileNotFoundError(
			"Composer not found — run: phpvm composer install", nil,
		)
	}
	return info, nil
}

// versionFromBinary runs `<binary> --version` and parses the output.
func (d *Detector) versionFromBinary(path string) (string, error) {
	out, err := d.execCommand(path, "--version", "--no-ansi")
	if err != nil {
		return "", cli.NewError(fmt.Sprintf("failed to run %s --version", path), err)
	}
	m := composerVersionRe.FindSubmatch(out)
	if m == nil {
		return "", cli.NewError(
			"could not parse Composer version from: "+strings.TrimSpace(string(out)), nil,
		)
	}
	return string(m[1]), nil
}

// SetExecCommand replaces the command executor (for testing).
func (d *Detector) SetExecCommand(fn func(name string, args ...string) ([]byte, error)) {
	d.execCommand = fn
}

// runCommand is the default executor.
func runCommand(name string, args ...string) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(name, args...)
	return cmd.Output()
}
