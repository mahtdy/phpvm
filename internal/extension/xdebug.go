package extension

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mahtdy/phpvm/internal/cli"
	phpini "github.com/mahtdy/phpvm/internal/ini"
	"github.com/mahtdy/phpvm/internal/logger"
)

// XdebugMode represents the xdebug.mode value.
type XdebugMode string

const (
	ModeOff      XdebugMode = "off"
	ModeDevelop  XdebugMode = "develop"
	ModeCoverage XdebugMode = "coverage"
	ModeProfile  XdebugMode = "profile"
	ModeTrace    XdebugMode = "trace"
	ModeDebug    XdebugMode = "debug"
)

// xdebugVersionRe extracts version from `php -r "echo phpversion('xdebug');"`.
var xdebugVersionRe = regexp.MustCompile(`(\d+\.\d+\.\d+)`)

// XdebugManager handles Xdebug-specific php.ini operations.
type XdebugManager struct {
	iniFile     *phpini.File
	execCommand func(string, ...string) ([]byte, error)
}

// NewXdebugManager loads the ini file at iniPath.
func NewXdebugManager(iniPath string) (*XdebugManager, error) {
	f, err := phpini.Load(iniPath)
	if err != nil {
		return nil, err
	}
	return &XdebugManager{iniFile: f, execCommand: defaultExec}, nil
}

// SetExecCommand replaces the executor (for testing).
func (x *XdebugManager) SetExecCommand(fn func(string, ...string) ([]byte, error)) {
	x.execCommand = fn
}

// IsInstalled returns true when a zend_extension=xdebug line is present.
func (x *XdebugManager) IsInstalled() bool {
	val, ok := x.iniFile.Get("zend_extension")
	if !ok {
		return false
	}
	return strings.Contains(strings.ToLower(val), "xdebug")
}

// CurrentMode returns the current xdebug.mode value (or "off" if not set).
func (x *XdebugManager) CurrentMode() XdebugMode {
	val, ok := x.iniFile.Get("xdebug.mode")
	if !ok {
		return ModeOff
	}
	return XdebugMode(strings.TrimSpace(val))
}

// SwitchMode sets xdebug.mode in php.ini and saves.
func (x *XdebugManager) SwitchMode(mode XdebugMode) error {
	if !x.IsInstalled() {
		return cli.NewValidationError(
			"Xdebug is not installed — run: phpvm xdebug install", nil,
		)
	}

	x.iniFile.Set("xdebug.mode", string(mode))
	if err := x.iniFile.Save(); err != nil {
		return cli.NewError("failed to update xdebug.mode", err)
	}
	logger.Info(fmt.Sprintf("Xdebug mode set to: %s", mode))
	return nil
}

// Enable adds or uncomments the zend_extension=xdebug line and sets mode=develop.
func (x *XdebugManager) Enable() error {
	if !x.IsInstalled() {
		return cli.NewValidationError(
			"Xdebug extension not found — run: phpvm xdebug install", nil,
		)
	}
	if err := x.SwitchMode(ModeDevelop); err != nil {
		return err
	}
	logger.Info("Xdebug enabled (mode=develop).")
	return nil
}

// Disable sets xdebug.mode=off in php.ini.
func (x *XdebugManager) Disable() error {
	return x.SwitchMode(ModeOff)
}

// DetectedVersion returns the installed Xdebug version string by querying PHP.
func (x *XdebugManager) DetectedVersion(phpBinary string) (string, error) {
	out, err := x.execCommand(phpBinary, "-r", "echo phpversion('xdebug');")
	if err != nil {
		return "", cli.NewError("failed to query Xdebug version", err)
	}
	m := xdebugVersionRe.FindString(string(out))
	if m == "" {
		return "", cli.NewFileNotFoundError("Xdebug does not appear to be loaded", nil)
	}
	return m, nil
}

// ValidModes returns all valid XdebugMode values.
func ValidModes() []XdebugMode {
	return []XdebugMode{ModeOff, ModeDevelop, ModeCoverage, ModeProfile, ModeTrace, ModeDebug}
}

// IsValidMode reports whether mode is a recognised Xdebug mode string.
func IsValidMode(mode string) bool {
	for _, m := range ValidModes() {
		if string(m) == strings.ToLower(mode) {
			return true
		}
	}
	return false
}
