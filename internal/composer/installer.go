package composer

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mahtdy/phpvm/internal/cli"
	"github.com/mahtdy/phpvm/internal/logger"
)

const (
	// installerURL is the official Composer installer.
	installerURL = "https://getcomposer.org/installer"
	// installerSigURL is the SHA-384 signature URL.
	installerSigURL = "https://composer.github.io/installer.sig"
	// httpTimeout for Composer installer downloads.
	httpTimeout = 5 * time.Minute
)

// InstallOptions configures a Composer installation.
type InstallOptions struct {
	// DestDir is where composer.phar will be placed (PHP version dir or global bin).
	DestDir string
	// PHPBinary is the PHP executable used to run the installer.
	PHPBinary string
	// Global installs into the system PATH location rather than DestDir.
	Global bool
	// SkipSignatureVerify skips the SHA-384 check (not recommended).
	SkipSignatureVerify bool
}

// Installer downloads and installs Composer.
type Installer struct {
	httpClient  *http.Client
	fetchURL    func(url string) ([]byte, error)
	execCommand func(name string, args ...string) ([]byte, error)
}

// NewInstaller creates an Installer with default settings.
func NewInstaller() *Installer {
	inst := &Installer{
		httpClient: &http.Client{Timeout: httpTimeout},
	}
	inst.fetchURL = inst.defaultFetch
	inst.execCommand = runCommand
	return inst
}

// Install downloads and installs composer.phar into opts.DestDir.
func (inst *Installer) Install(ctx context.Context, opts InstallOptions) (*Info, error) {
	if opts.PHPBinary == "" {
		return nil, cli.NewValidationError("PHP binary path is required to install Composer", nil)
	}
	if opts.DestDir == "" {
		return nil, cli.NewValidationError("destination directory is required", nil)
	}

	if err := os.MkdirAll(opts.DestDir, 0o750); err != nil {
		return nil, cli.NewPermissionError("cannot create composer destination dir", err)
	}

	logger.Info("Downloading Composer installer...")

	// 1. Download the installer script.
	installerData, err := inst.fetchURL(installerURL)
	if err != nil {
		return nil, err
	}

	// 2. Verify SHA-384 signature.
	if !opts.SkipSignatureVerify {
		if err := inst.verifySignature(installerData); err != nil {
			return nil, err
		}
		logger.Info("Composer installer signature verified.")
	}

	// 3. Write installer to a temp file.
	tmpFile, err := os.CreateTemp("", "composer-setup-*.php")
	if err != nil {
		return nil, cli.NewError("failed to create temp installer file", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(installerData); err != nil {
		tmpFile.Close()
		return nil, cli.NewError("failed to write installer", err)
	}
	tmpFile.Close()

	// 4. Run the installer: php setup.php --install-dir=<dest> --filename=composer
	destArg := "--install-dir=" + opts.DestDir
	logger.Info(fmt.Sprintf("Installing Composer into %s...", opts.DestDir))

	out, err := inst.execCommand(opts.PHPBinary, tmpFile.Name(), destArg, "--filename=composer")
	if err != nil {
		return nil, cli.NewError(
			fmt.Sprintf("Composer installer failed: %s", string(out)), err,
		)
	}

	// 5. Detect the installed binary.
	composerPath := filepath.Join(opts.DestDir, composerBinName())
	if _, err := os.Stat(composerPath); err != nil {
		return nil, cli.NewFileNotFoundError(
			"Composer binary not found after installation", err,
		)
	}

	// 6. Detect version.
	detector := NewDetector()
	info, err := detector.DetectPerVersion(opts.DestDir)
	if err != nil || info == nil {
		// Fallback: return without version.
		return &Info{Path: composerPath, Global: opts.Global}, nil
	}

	logger.Info(fmt.Sprintf("Composer %s installed at %s", info.Version, composerPath))
	return info, nil
}

// verifySignature downloads the expected SHA-384 hash and compares.
func (inst *Installer) verifySignature(installerData []byte) error {
	sigData, err := inst.fetchURL(installerSigURL)
	if err != nil {
		return cli.NewNetworkError("failed to download Composer signature", err)
	}

	expected := string(sigData)
	// Trim whitespace / newlines.
	for len(expected) > 0 && (expected[len(expected)-1] == '\n' || expected[len(expected)-1] == '\r') {
		expected = expected[:len(expected)-1]
	}

	h := sha512.New384()
	io.WriteString(h, string(installerData)) //nolint:errcheck
	actual := hex.EncodeToString(h.Sum(nil))

	if actual != expected {
		return cli.NewError(
			fmt.Sprintf("Composer installer signature mismatch: expected %s, got %s",
				expected, actual), nil,
		)
	}
	return nil
}

// defaultFetch performs an HTTP GET.
func (inst *Installer) defaultFetch(url string) ([]byte, error) {
	resp, err := inst.httpClient.Get(url) //nolint:gosec
	if err != nil {
		return nil, cli.NewNetworkError(fmt.Sprintf("failed to fetch %s", url), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, cli.NewNetworkError(
			fmt.Sprintf("HTTP %d fetching %s", resp.StatusCode, url), nil,
		)
	}
	return io.ReadAll(resp.Body)
}

// SetFetchURL replaces the HTTP fetcher (for testing).
func (inst *Installer) SetFetchURL(fn func(url string) ([]byte, error)) {
	inst.fetchURL = fn
}

// SetExecCommand replaces the command executor (for testing).
func (inst *Installer) SetExecCommand(fn func(name string, args ...string) ([]byte, error)) {
	inst.execCommand = fn
}

// composerBinName returns the expected composer binary name after installation.
func composerBinName() string {
	if runtime.GOOS == "windows" {
		return "composer.bat"
	}
	return "composer"
}
