package php

import (
	"context"
	"fmt"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/logger"
	"github.com/mahtdy/phpvm/internal/validation"
	ver "github.com/mahtdy/phpvm/internal/version"
)

// UpdateResult describes the outcome of an update operation.
type UpdateResult struct {
	OldVersion string
	NewVersion string
	Updated    bool // false when already on latest
}

// Updater checks for and applies PHP patch updates.
type Updater struct {
	cfg       *config.Config
	scanner   *Scanner
	installer *Installer
	resolver  *ver.Resolver
}

// NewUpdater creates an Updater backed by cfg.
func NewUpdater(cfg *config.Config) *Updater {
	u := &Updater{
		cfg:       cfg,
		scanner:   NewScanner(cfg.VersionsDir()),
		installer: NewInstaller(cfg),
		resolver:  ver.NewResolver(cfg.CacheDir()),
	}
	return u
}

// SetResolver replaces the release resolver (for testing).
func (u *Updater) SetResolver(r *ver.Resolver) {
	u.resolver = r
	u.installer.SetResolver(r)
}

// Update checks for a newer patch release of the given version and installs it.
// selectorStr may be empty (defaults to current version) or a specific version string.
func (u *Updater) Update(ctx context.Context, selectorStr string, noProgress bool) (*UpdateResult, error) {
	if selectorStr == "" {
		if u.cfg.Current == "" {
			return nil, cli.NewValidationError(
				"no active PHP version — specify a version to update, e.g. phpvm update 8.3",
				nil,
			)
		}
		selectorStr = u.cfg.Current
	}

	sel, err := validation.Parse(selectorStr)
	if err != nil {
		return nil, err
	}

	// Find what is currently installed for this major.minor.
	installed, err := u.scanner.FindBestMatch(sel)
	if err != nil {
		return nil, err
	}
	oldVersion := installed.Version.String()

	logger.Info(fmt.Sprintf("Checking for updates to PHP %s...", oldVersion))

	// Resolve the latest available patch from php.net.
	latest, err := u.resolver.LatestPatch(sel)
	if err != nil {
		return nil, err
	}

	if !installed.Version.Less(latest.Version) {
		logger.Info(fmt.Sprintf("PHP %s is already up to date.", oldVersion))
		return &UpdateResult{OldVersion: oldVersion, NewVersion: oldVersion, Updated: false}, nil
	}

	newVersion := latest.Version.String()
	logger.Info(fmt.Sprintf("Updating PHP %s → %s...", oldVersion, newVersion))

	// Install the newer patch.
	iv, err := u.installer.Install(ctx, InstallOptions{
		VersionStr: newVersion,
		NoProgress: noProgress,
	})
	if err != nil {
		return nil, cli.NewError(
			fmt.Sprintf("failed to install PHP %s", newVersion), err,
		)
	}

	// If the old version was the active one, switch to the new version.
	if u.cfg.Current == oldVersion {
		switcher := NewSwitcher(u.cfg)
		if _, err := switcher.Switch(newVersion); err != nil {
			logger.Warn("update succeeded but could not switch active version", "error", err)
		}
	}

	logger.Info(fmt.Sprintf("PHP updated to %s.", iv.Version.String()))
	return &UpdateResult{OldVersion: oldVersion, NewVersion: newVersion, Updated: true}, nil
}

// UpdateAll updates every installed major.minor to its latest patch.
func (u *Updater) UpdateAll(ctx context.Context, noProgress bool) ([]*UpdateResult, error) {
	installed, err := u.scanner.ScanInstalled()
	if err != nil {
		return nil, err
	}
	if len(installed) == 0 {
		return nil, cli.NewNotInstalledError("any version")
	}

	var results []*UpdateResult
	// Group by major.minor and update each once.
	seen := make(map[string]bool)
	for _, iv := range installed {
		minor := fmt.Sprintf("%d.%d", iv.Version.Major, iv.Version.Minor)
		if seen[minor] {
			continue
		}
		seen[minor] = true

		result, err := u.Update(ctx, minor, noProgress)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to update PHP %s: %v", minor, err))
			continue
		}
		results = append(results, result)
	}
	return results, nil
}
