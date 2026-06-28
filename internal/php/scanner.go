// Package php provides core PHP version management: scanning, switching,
// current-version detection, local pinning, and framework detection.
package php

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/validation"
)

// phpBinaryName is the PHP executable name for the current platform.
func phpBinaryName() string {
	if runtime.GOOS == "windows" {
		return "php.exe"
	}
	return "php"
}

// InstalledVersion combines a parsed version with its on-disk path.
type InstalledVersion struct {
	Version *validation.Version
	Dir     string // absolute path to the version's root directory
}

// Scanner scans the phpvm versions directory for installed PHP builds.
type Scanner struct {
	versionsDir string
}

// NewScanner creates a Scanner for the given versions directory.
func NewScanner(versionsDir string) *Scanner {
	return &Scanner{versionsDir: versionsDir}
}

// ScanInstalled returns all installed PHP versions sorted in ascending semver order.
// A subdirectory is considered a valid PHP installation if it contains
// the php / php.exe binary directly inside it.
func (s *Scanner) ScanInstalled() ([]*InstalledVersion, error) {
	entries, err := os.ReadDir(s.versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No versions directory yet — return empty list, not an error.
			return nil, nil
		}
		return nil, cli.NewFileNotFoundError(
			"cannot read versions directory: "+s.versionsDir, err,
		)
	}

	var result []*InstalledVersion

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		dirName := e.Name()
		v, err := validation.Parse(dirName)
		if err != nil {
			// Directory name is not a valid version — skip silently.
			continue
		}

		binPath := filepath.Join(s.versionsDir, dirName, phpBinaryName())
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			// No PHP binary inside — skip.
			continue
		}

		result = append(result, &InstalledVersion{
			Version: v,
			Dir:     filepath.Join(s.versionsDir, dirName),
		})
	}

	// Sort ascending by semver.
	rawVersions := make([]*validation.Version, len(result))
	for i, iv := range result {
		rawVersions[i] = iv.Version
	}
	// Insertion-sort result slice parallel to rawVersions.
	sortInstalled(result)

	return result, nil
}

// FindBestMatch finds the installed version that best matches a selector.
// "Best match" means the highest-patch version that satisfies the selector.
// e.g. selector "8.3" matches "8.3.10" and "8.3.9" — returns "8.3.10".
func (s *Scanner) FindBestMatch(selector *validation.Version) (*InstalledVersion, error) {
	installed, err := s.ScanInstalled()
	if err != nil {
		return nil, err
	}

	var best *InstalledVersion
	for _, iv := range installed {
		if iv.Version.Matches(selector) {
			if best == nil || best.Version.Less(iv.Version) {
				best = iv
			}
		}
	}

	if best == nil {
		return nil, cli.NewNotInstalledError(selector.Raw)
	}
	return best, nil
}

// sortInstalled sorts a slice of InstalledVersion ascending by semver (in-place).
func sortInstalled(vs []*InstalledVersion) {
	n := len(vs)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && vs[j].Version.Less(vs[j-1].Version); j-- {
			vs[j], vs[j-1] = vs[j-1], vs[j]
		}
	}
}
