package php

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mahtdy/phpvm/internal/archive"
	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/config"
	"github.com/mahtdy/phpvm/internal/downloader"
	"github.com/mahtdy/phpvm/internal/logger"
	"github.com/mahtdy/phpvm/internal/validation"
	ver "github.com/mahtdy/phpvm/internal/version"
)

// InstallOptions configures a PHP installation.
type InstallOptions struct {
	// VersionStr is the user-supplied version selector (e.g. "8.4", "8.4.1").
	VersionStr string
	// Variant selects "ts" or "nts" (default: "nts").
	Variant string
	// Arch overrides the CPU architecture (default: from config).
	Arch string
	// Force reinstalls even if the version is already present.
	Force bool
	// NoProgress disables the terminal progress bar.
	NoProgress bool
}

// Installer orchestrates the full install pipeline:
// Resolve → Download → Verify → Extract → Configure.
type Installer struct {
	cfg      *config.Config
	resolver *ver.Resolver
}

// NewInstaller creates an Installer backed by cfg.
func NewInstaller(cfg *config.Config) *Installer {
	return &Installer{
		cfg:      cfg,
		resolver: ver.NewResolver(cfg.CacheDir()),
	}
}

// SetResolver replaces the release resolver (for testing).
func (inst *Installer) SetResolver(r *ver.Resolver) {
	inst.resolver = r
}

// Install downloads and installs a PHP version.
func (inst *Installer) Install(ctx context.Context, opts InstallOptions) (*InstalledVersion, error) {
	// 1. Parse and validate the version selector.
	sel, err := validation.Parse(opts.VersionStr)
	if err != nil {
		return nil, err
	}
	if opts.Variant != "" {
		sel.Variant = opts.Variant
	}

	logger.Info(fmt.Sprintf("Resolving PHP %s...", sel.String()))

	// 2. Resolve to the best available release.
	release, err := inst.resolver.LatestPatch(sel)
	if err != nil {
		return nil, err
	}

	resolvedVersion := release.Version.String()
	installDir := filepath.Join(inst.cfg.VersionsDir(), resolvedVersion)

	// 3. Check if already installed.
	binPath := filepath.Join(installDir, phpBinaryName())
	if _, err := os.Stat(binPath); err == nil && !opts.Force {
		logger.Info(fmt.Sprintf("PHP %s is already installed. Use --force to reinstall.", resolvedVersion))
		return &InstalledVersion{Version: release.Version, Dir: installDir}, nil
	}

	// 4. Ensure directories exist.
	if err := EnsureVersionsDir(inst.cfg); err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Downloading PHP %s (%s/%s)...",
		resolvedVersion, platformLabel(), archLabel(opts.Arch, inst.cfg)))

	// 5. Download the archive.
	dlResult, err := downloader.Download(ctx, downloader.Options{
		URL:            release.URL,
		DestDir:        inst.cfg.DownloadsDir(),
		Filename:       release.Filename,
		ExpectedSHA256: release.SHA256,
		NoProgress:     opts.NoProgress,
	})
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Extracting to %s...", installDir))

	// 6. Extract atomically.
	if err := archive.Extract(dlResult.Path, installDir); err != nil {
		return nil, err
	}

	// 7. Verify the binary exists after extraction.
	if _, err := os.Stat(binPath); err != nil {
		return nil, cli.NewArchiveError(
			fmt.Sprintf("PHP binary not found after extraction at %s", binPath), err,
		)
	}

	logger.Info(fmt.Sprintf("PHP %s installed successfully.", resolvedVersion))
	logger.Info(fmt.Sprintf("Run 'phpvm use %s' to activate it.", resolvedVersion))

	return &InstalledVersion{Version: release.Version, Dir: installDir}, nil
}

func platformLabel() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	default:
		return "Linux"
	}
}

func archLabel(override string, cfg *config.Config) string {
	if override != "" {
		return override
	}
	return cfg.DefaultArch
}
