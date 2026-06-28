package php

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/logger"
	"github.com/mahtdy/phpvm/internal/path"
	"github.com/mahtdy/phpvm/internal/validation"
)

// Switcher handles PHP version switching.
type Switcher struct {
	cfg     *config.Config
	scanner *Scanner
	pathMgr path.Manager
}

// NewSwitcher creates a Switcher backed by the given configuration.
func NewSwitcher(cfg *config.Config) *Switcher {
	return &Switcher{
		cfg:     cfg,
		scanner: NewScanner(cfg.VersionsDir()),
		pathMgr: path.New(),
	}
}

// NewSwitcherWithPathMgr creates a Switcher with a custom path.Manager.
// This is intended for testing purposes.
func NewSwitcherWithPathMgr(cfg *config.Config, mgr path.Manager) *Switcher {
	return &Switcher{
		cfg:     cfg,
		scanner: NewScanner(cfg.VersionsDir()),
		pathMgr: mgr,
	}
}

// Switch activates the best matching installed PHP version for selector.
// It updates the config and the system PATH.
func (s *Switcher) Switch(selectorStr string) (*InstalledVersion, error) {
	selector, err := validation.Parse(selectorStr)
	if err != nil {
		return nil, err
	}

	iv, err := s.scanner.FindBestMatch(selector)
	if err != nil {
		return nil, err
	}

	// Update PATH to point to the new version's directory.
	if pathErr := s.pathMgr.Add(iv.Dir); pathErr != nil {
		logger.Warn("could not update PATH", "error", pathErr)
		// Non-fatal: user can add it manually.
	}

	// Persist the active version to config.
	s.cfg.Current = iv.Version.String()
	if err := config.Save(s.cfg); err != nil {
		return nil, cli.NewError("failed to save config after switch", err)
	}

	logger.Info(fmt.Sprintf("Switched to PHP %s", iv.Version.String()))
	return iv, nil
}

// EnsureVersionsDir creates the versions directory if it does not exist.
func EnsureVersionsDir(cfg *config.Config) error {
	dirs := []string{
		cfg.VersionsDir(),
		cfg.DownloadsDir(),
		cfg.CacheDir(),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o750); err != nil {
			return cli.NewPermissionError(
				fmt.Sprintf("cannot create directory %s", d), err,
			)
		}
	}
	return nil
}

// InstallDir returns the expected installation directory for a specific version.
func InstallDir(versionsDir string, v *validation.Version) string {
	return filepath.Join(versionsDir, v.String())
}
