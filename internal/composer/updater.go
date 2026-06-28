package composer

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/logger"
)

// latestVersionRe extracts version from getcomposer.org/versions JSON.
// {"latest": {"version": "2.7.0", ...}}
var latestVersionRe = regexp.MustCompile(`"version"\s*:\s*"([^"]+)"`)

// Updater checks for and applies Composer updates.
type Updater struct {
	detector    *Detector
	installer   *Installer
	fetchURL    func(url string) ([]byte, error)
	execCommand func(name string, args ...string) ([]byte, error)
}

// NewUpdater creates an Updater with default settings.
func NewUpdater() *Updater {
	u := &Updater{
		detector:  NewDetector(),
		installer: NewInstaller(),
	}
	u.fetchURL = u.installer.defaultFetch
	u.execCommand = runCommand
	return u
}

// Update upgrades Composer to the latest stable version in destDir.
// phpBinary is needed to run the installer if a full reinstall is required.
func (u *Updater) Update(ctx context.Context, phpVersionDir, phpBinary string) (*Info, error) {
	// 1. Detect current installation.
	current, err := u.detector.DetectPerVersion(phpVersionDir)
	if err != nil {
		return nil, err
	}
	if current == nil {
		// Try global.
		current, _ = u.detector.DetectGlobal()
	}

	currentVersion := ""
	if current != nil {
		currentVersion = current.Version
	}

	// 2. Fetch the latest available version.
	latestVersion, err := u.fetchLatestVersion()
	if err != nil {
		return nil, err
	}

	if currentVersion == latestVersion {
		logger.Info(fmt.Sprintf("Composer %s is already up to date.", currentVersion))
		return current, nil
	}

	logger.Info(fmt.Sprintf("Updating Composer %s → %s...", currentVersion, latestVersion))

	// 3. Use `composer self-update` if composer is available; otherwise reinstall.
	if current != nil {
		out, err := u.execCommand(current.Path, "self-update")
		if err == nil {
			logger.Info(fmt.Sprintf("Composer updated to %s.", latestVersion))
			current.Version = latestVersion
			return current, nil
		}
		logger.Warn("self-update failed, falling back to reinstall", "output", strings.TrimSpace(string(out)))
	}

	// 4. Fallback: reinstall via installer.
	destDir := phpVersionDir
	if destDir == "" && current != nil {
		destDir = current.Path
	}
	if destDir == "" {
		return nil, cli.NewValidationError("cannot determine Composer install location", nil)
	}

	return u.installer.Install(ctx, InstallOptions{
		DestDir:             destDir,
		PHPBinary:           phpBinary,
		SkipSignatureVerify: false,
	})
}

// fetchLatestVersion queries getcomposer.org for the latest stable version.
func (u *Updater) fetchLatestVersion() (string, error) {
	data, err := u.fetchURL("https://getcomposer.org/versions")
	if err != nil {
		return "", cli.NewNetworkError("failed to fetch Composer versions", err)
	}

	m := latestVersionRe.FindSubmatch(data)
	if m == nil {
		return "", cli.NewError("could not parse latest Composer version from getcomposer.org", nil)
	}
	return string(m[1]), nil
}

// SetFetchURL replaces the HTTP fetcher (for testing).
func (u *Updater) SetFetchURL(fn func(url string) ([]byte, error)) {
	u.fetchURL = fn
}

// SetExecCommand replaces the executor (for testing).
func (u *Updater) SetExecCommand(fn func(name string, args ...string) ([]byte, error)) {
	u.execCommand = fn
}
